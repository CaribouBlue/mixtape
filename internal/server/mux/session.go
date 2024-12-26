package mux

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/CaribouBlue/top-spot/internal/db"
	"github.com/CaribouBlue/top-spot/internal/model"
	"github.com/CaribouBlue/top-spot/internal/server/middleware"
	"github.com/CaribouBlue/top-spot/internal/spotify"
	"github.com/CaribouBlue/top-spot/internal/templates"
)

func appHomeHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()

	session := model.NewSessionModel(db)

	sessions, err := session.ReadAll()
	if err != nil {
		http.Error(w, "Failed to get sessions", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewHomeTemplateModel(user, sessions)
	component := templates.Home(templateModel)
	handleHtmlResponse(r, w, component)
}

func createAppSessionHandler(w http.ResponseWriter, r *http.Request) {
	db := db.Global()

	session := model.NewSessionModel(db)

	defer r.Body.Close()
	if err := session.Scan(r.Body); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	err := session.Read()
	if err == sql.ErrNoRows {
		err = session.Create()
	} else if err == nil {
		http.Error(w, "Session already exists", http.StatusConflict)
		return
	}

	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(session); err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
	}
}

func appSessionHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session := model.NewSessionModel(db, model.WithId(sessionId))
	err = session.Read()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	acceptHeader := r.Header.Get("Accept")
	switch strings.ToLower(acceptHeader) {
	case "application/json":
		handleJsonResponse(w, session)
	case "text/html":
	default:
		templateModel := templates.NewSessionTemplateModel(*session, *user)
		component := templates.Session(templateModel, "")
		handleHtmlResponse(r, w, component)
	}
}

func appSessionTracksSearchHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()
	spotifyClient := authorizedSpotifyClient(user)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session := model.NewSessionModel(db, model.WithId(sessionId))
	err = session.Read()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	r.ParseForm()
	query := r.Form.Get("query")

	templateModel := templates.NewSessionTemplateModel(*session, *user)

	if query != "" {
		searchResults, err := spotifyClient.SearchTracks(query)
		if err != nil {
			http.Error(w, "Failed to search Spotify", http.StatusInternalServerError)
			return
		}

		templateModel.SearchResult = *searchResults
	}

	templates.SubmissionSearchBar(templateModel, "").Render(r.Context(), w)
}

func appSessionCreatePlaylistHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()
	spotifyClient := authorizedSpotifyClient(user)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session := model.NewSessionModel(db, model.WithId(sessionId))
	err = session.Read()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	playlist, err := spotifyClient.CreatePlaylist(session.PlaylistName())
	if err != nil {
		fmt.Printf("%s\n", err)
		http.Error(w, "Failed to create playlist", http.StatusInternalServerError)
		return
	}

	trackIds := make([]string, len(session.Data.Submissions))
	for i, submission := range session.Data.Submissions {
		trackIds[i] = submission.TrackId
	}
	err = spotifyClient.AddTracksToPlaylist(playlist.Id, trackIds)
	if err != nil {
		spotifyClient.UnfollowPlaylist(playlist.Id)
		http.Error(w, "Failed to add tracks to playlist", http.StatusInternalServerError)
		return
	}

	err = user.AddPlaylist(playlist.Id, session.Id())
	if err != nil {
		spotifyClient.UnfollowPlaylist(playlist.Id)
		http.Error(w, "Failed to add playlist to user", http.StatusInternalServerError)
		return
	}

	err = user.Update()
	if err != nil {
		spotifyClient.UnfollowPlaylist(playlist.Id)
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, model.UserModel{})
	handleHtmlResponse(r, w, templates.VotePlaylistButton(templateModel, playlist.ExternalUrls.Spotify))
}

func appSessionPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()
	spotifyClient := authorizedSpotifyClient(user)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session := model.NewSessionModel(db, model.WithId(sessionId))
	err = session.Read()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	var playlist *spotify.Playlist
	userPlaylist, err := user.GetSessionPlaylist(session.Id())
	if err == model.ErrUserPlaylistNotFound {
		playlist = &spotify.Playlist{}
	} else if err != nil {
		http.Error(w, "Failed to get playlist", http.StatusInternalServerError)
		return
	} else {
		playlist, err = spotifyClient.GetPlaylist(userPlaylist.Id)
		if err != nil {
			http.Error(w, "Failed to get playlist", http.StatusInternalServerError)
			return
		}
	}

	templateModel := templates.NewSessionTemplateModel(*session, model.UserModel{})
	templates.VotePlaylistButton(templateModel, playlist.Uri)
	handleHtmlResponse(r, w, templates.VotePlaylistButton(templateModel, playlist.ExternalUrls.Spotify))
}

func appSessionSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session := model.NewSessionModel(db, model.WithId(sessionId))
	err = session.Read()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	r.ParseForm()
	trackId := r.Form.Get("trackId")
	submission := model.NewSubmissionData(user.Id(), trackId)

	err = session.AddSubmission(*submission)
	if err != nil {
		if err == model.ErrSubmissionsMaxedOut {
			http.Error(w, "Max submissions reached", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to add submission", http.StatusInternalServerError)
		return
	}

	err = session.Update()
	if err != nil {
		http.Error(w, "Failed to add submission", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	templates.NewSubmission(templateModel, *submission).Render(r.Context(), w)
}

func appSessionTimeLeftHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session := model.NewSessionModel(db, model.WithId(sessionId))
	err = session.Read()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	templates.SessionPhaseTimeLeft(templateModel).Render(r.Context(), w)
}

func appSessionSubmissionDetailsHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()
	spotifyClient := authorizedSpotifyClient(user)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	submissionId := r.PathValue("submissionId")
	if submissionId == "" {
		http.Error(w, "Invalid submission ID", http.StatusBadRequest)
		return
	}

	session := model.NewSessionModel(db, model.WithId(sessionId))
	err = session.Read()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	submission, err := session.GetSubmission(submissionId)
	if err != nil {
		http.Error(w, "Failed to get submission", http.StatusInternalServerError)
		return
	}

	track, err := spotifyClient.GetTrack(submission.TrackId)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	templates.SubmissionListItem(templateModel, *submission, *track).Render(r.Context(), w)
}

func appSessionDeleteSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	submissionId := r.PathValue("submissionId")
	if submissionId == "" {
		http.Error(w, "Invalid submission ID", http.StatusBadRequest)
		return
	}

	session := model.NewSessionModel(db, model.WithId(sessionId))
	err = session.Read()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	err = session.DeleteSubmission(submissionId, user.Id())
	if err != nil {
		http.Error(w, "Failed to delete submission", http.StatusInternalServerError)
		return
	}

	err = session.Update()
	if err != nil {
		http.Error(w, "Failed to delete submission", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	templates.DeleteSubmission(templateModel).Render(r.Context(), w)
}

func appSessionSubmissionCandidateHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()
	spotifyClient := authorizedSpotifyClient(user)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	submissionId := r.PathValue("submissionId")
	if submissionId == "" {
		http.Error(w, "Invalid submission ID", http.StatusBadRequest)
		return
	}

	session := model.NewSessionModel(db, model.WithId(sessionId))
	err = session.Read()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	submission, err := session.GetSubmission(submissionId)
	if err != nil {
		http.Error(w, "Failed to get submission", http.StatusInternalServerError)
		return
	}

	track, err := spotifyClient.GetTrack(submission.TrackId)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	templates.VoteListCandidate(templateModel, *submission, *track).Render(r.Context(), w)
}

func appSessionVoteHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session := model.NewSessionModel(db, model.WithId(sessionId))
	err = session.Read()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	r.ParseForm()
	submissionId := r.Form.Get("submissionId")

	session.AddVote(*model.NewVoteModel(user.Id(), submissionId))

	err = session.Update()
	if err != nil {
		http.Error(w, "Failed to add vote", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	templates.LazyLoadVoteCandidate(templateModel, submissionId).Render(r.Context(), w)
}

func appSessionDeleteVoteHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	voteId := r.PathValue("voteId")
	if voteId == "" {
		http.Error(w, "Invalid vote ID", http.StatusBadRequest)
		return
	}

	session := model.NewSessionModel(db, model.WithId(sessionId))
	err = session.Read()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	vote, err := session.GetVote(voteId, user.Id())
	if err != nil {
		http.Error(w, "Failed to get vote", http.StatusInternalServerError)
		return
	}

	err = session.DeleteVote(voteId, user.Id())
	if err != nil {
		http.Error(w, "Failed to delete vote", http.StatusInternalServerError)
		return
	}

	err = session.Update()
	if err != nil {
		http.Error(w, "Failed to delete vote", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	templates.LazyLoadVoteCandidate(templateModel, vote.SubmissionId).Render(r.Context(), w)
}

type SessionMux struct {
	*http.ServeMux
}

func NewSessionMux() *SessionMux {
	mux := &SessionMux{http.NewServeMux()}
	mux.RegisterHandlers()
	return mux
}

func (mux *SessionMux) RegisterHandlers() {
	mux.Handle("GET /", http.HandlerFunc(appHomeHandler))
	mux.Handle("POST /", http.HandlerFunc(createAppSessionHandler))

	mux.Handle("GET /{sessionId}", http.HandlerFunc(appSessionHandler))
	mux.Handle("POST /{sessionId}/tracks", http.HandlerFunc(appSessionTracksSearchHandler))
	mux.Handle("POST /{sessionId}/playlist", http.HandlerFunc(appSessionCreatePlaylistHandler))
	mux.Handle("GET /{sessionId}/playlist", http.HandlerFunc(appSessionPlaylistHandler))

	mux.Handle("POST /{sessionId}/submission", http.HandlerFunc(appSessionSubmissionHandler))
	mux.Handle("GET /{sessionId}/submission/time-left", http.HandlerFunc(appSessionTimeLeftHandler))
	mux.Handle("GET /{sessionId}/submission/{submissionId}", http.HandlerFunc(appSessionSubmissionDetailsHandler))
	mux.Handle("DELETE /{sessionId}/submission/{submissionId}", http.HandlerFunc(appSessionDeleteSubmissionHandler))
	mux.Handle("GET /{sessionId}/submission/{submissionId}/candidate", http.HandlerFunc(appSessionSubmissionCandidateHandler))

	mux.Handle("POST /{sessionId}/vote", http.HandlerFunc(appSessionVoteHandler))
	mux.Handle("DELETE /{sessionId}/vote/{voteId}", http.HandlerFunc(appSessionDeleteVoteHandler))
	mux.Handle("GET /{sessionId}/vote/time-left", http.HandlerFunc(appSessionTimeLeftHandler))
}
