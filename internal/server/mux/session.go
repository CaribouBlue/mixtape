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
	"github.com/CaribouBlue/top-spot/internal/server/utils"
	"github.com/CaribouBlue/top-spot/internal/spotify"
	"github.com/CaribouBlue/top-spot/internal/templates"
)

type SessionMux struct {
	*http.ServeMux
}

func NewSessionMux() *SessionMux {
	mux := &SessionMux{http.NewServeMux()}
	mux.RegisterHandlers()
	return mux
}

func (mux *SessionMux) RegisterHandlers() {
	mux.Handle("GET /", http.HandlerFunc(mux.handleSessionListPage))
	mux.Handle("POST /", http.HandlerFunc(mux.handleCreateSession))

	mux.Handle("GET /{sessionId}", http.HandlerFunc(mux.handleSessionPage))

	mux.Handle("POST /{sessionId}/tracks", http.HandlerFunc(mux.handleCreateSessionTrack))

	mux.Handle("POST /{sessionId}/playlist", http.HandlerFunc(mux.handleCreateSessionPlaylist))
	mux.Handle("GET /{sessionId}/playlist", http.HandlerFunc(mux.handleGetSessionPlaylist))

	mux.Handle("GET /{sessionId}/phase-duration", http.HandlerFunc(mux.handleGetSessionPhaseDuration))

	mux.Handle("POST /{sessionId}/submission", http.HandlerFunc(mux.handleCreateSessionSubmission))

	mux.Handle("GET /{sessionId}/submission/{submissionId}", http.HandlerFunc(mux.handleGetSessionSubmission))
	mux.Handle("DELETE /{sessionId}/submission/{submissionId}", http.HandlerFunc(mux.handleDeleteSessionSubmission))

	mux.Handle("GET /{sessionId}/submission/{submissionId}/candidate", http.HandlerFunc(mux.handleGetSessionSubmissionCandidate))

	mux.Handle("POST /{sessionId}/vote", http.HandlerFunc(mux.handleCreateSessionVote))

	mux.Handle("DELETE /{sessionId}/vote/{voteId}", http.HandlerFunc(mux.handleDeleteSessionVote))
}

func (mux *SessionMux) handleSessionListPage(w http.ResponseWriter, r *http.Request) {
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
	utils.HandleHtmlResponse(r, w, component)
}

func (mux *SessionMux) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	db := db.Global()

	session := model.NewSessionModel(db)

	defer r.Body.Close()
	if err := session.Scan(r.Body); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	err := session.Create()
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

func (mux *SessionMux) handleSessionPage(w http.ResponseWriter, r *http.Request) {
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
		utils.HandleJsonResponse(w, session)
	case "text/html":
	default:
		templateModel := templates.NewSessionTemplateModel(*session, *user)
		component := templates.Session(templateModel, "")
		utils.HandleHtmlResponse(r, w, component)
	}
}

func (mux *SessionMux) handleCreateSessionTrack(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()
	spotifyClient := utils.AuthorizedSpotifyClient(user)

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

func (mux *SessionMux) handleCreateSessionPlaylist(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()
	spotifyClient := utils.AuthorizedSpotifyClient(user)

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
	utils.HandleHtmlResponse(r, w, templates.PlaylistButton(templateModel, *playlist))
}

func (mux *SessionMux) handleGetSessionPlaylist(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()
	spotifyClient := utils.AuthorizedSpotifyClient(user)

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

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	utils.HandleHtmlResponse(r, w, templates.PlaylistButton(templateModel, *playlist))
}

func (mux *SessionMux) handleCreateSessionSubmission(w http.ResponseWriter, r *http.Request) {
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

func (mux *SessionMux) handleGetSessionPhaseDuration(w http.ResponseWriter, r *http.Request) {
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
	templates.SessionPhaseDuration(templateModel).Render(r.Context(), w)
}

func (mux *SessionMux) handleGetSessionSubmission(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()
	spotifyClient := utils.AuthorizedSpotifyClient(user)

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
	templates.SubmissionItem(templateModel, *submission, *track).Render(r.Context(), w)
}

func (mux *SessionMux) handleDeleteSessionSubmission(w http.ResponseWriter, r *http.Request) {
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

func (mux *SessionMux) handleGetSessionSubmissionCandidate(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()
	spotifyClient := utils.AuthorizedSpotifyClient(user)

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
	templates.VoteCandidate(templateModel, *submission, *track).Render(r.Context(), w)
}

func (mux *SessionMux) handleCreateSessionVote(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()
	spotifyClient := utils.AuthorizedSpotifyClient(user)

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

	session.AddVote(*model.NewVoteModel(user.Id(), submissionId))

	err = session.Update()
	if err != nil {
		http.Error(w, "Failed to add vote", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	templates.VoteCandidate(templateModel, *submission, *track).Render(r.Context(), w)
}

func (mux *SessionMux) handleDeleteSessionVote(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	db := db.Global()
	spotifyClient := utils.AuthorizedSpotifyClient(user)

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

	submission, err := session.GetSubmission(vote.SubmissionId)
	if err != nil {
		http.Error(w, "Failed to get submission", http.StatusInternalServerError)
		return
	}

	track, err := spotifyClient.GetTrack(submission.TrackId)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
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
	templates.VoteCandidate(templateModel, *submission, *track).Render(r.Context(), w)
}
