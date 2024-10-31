package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/CaribouBlue/top-spot/db"
	"github.com/CaribouBlue/top-spot/spotify"
	"github.com/CaribouBlue/top-spot/templates"
)

const (
	appMuxPathPrefix string = "/app"
)

func appLandingHandler(w http.ResponseWriter, r *http.Request) {
	templates.Landing().Render(r.Context(), w)
}

func createAppSessionHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var session *db.GameSessionDataModel = db.NewGameSessionDataModel()
	if err := json.NewDecoder(r.Body).Decode(session); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	err := session.GetById()
	if err == sql.ErrNoRows {
		err = session.Insert()
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
	user, err := getUserFromRequestContext(r)
	if err != nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session := db.NewGameSessionDataModel()
	session.SetId(sessionId)
	err = session.GetById()
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
	user, err := getUserFromRequestContext(r)
	if err != nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	spotifyClient, err := getSpotifyClientFromRequestContext(r)
	if err != nil {
		http.Error(w, "Spotify not found in context", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session := db.NewGameSessionDataModel()
	session.SetId(sessionId)
	err = session.GetById()
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
	user, err := getUserFromRequestContext(r)
	if err != nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	spotifyClient, err := getSpotifyClientFromRequestContext(r)
	if err != nil {
		http.Error(w, "Spotify not found in context", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session := db.NewGameSessionDataModel()
	session.SetId(sessionId)
	err = session.GetById()
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

	trackIds := make([]string, len(session.Submissions))
	for i, submission := range session.Submissions {
		trackIds[i] = submission.TrackId
	}
	err = spotifyClient.AddTracksToPlaylist(playlist.Id, trackIds)
	if err != nil {
		spotifyClient.UnfollowPlaylist(playlist.Id)
		http.Error(w, "Failed to add tracks to playlist", http.StatusInternalServerError)
		return
	}

	err = user.AddPlaylist(playlist.Id, session.Id)
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

	templateModel := templates.NewSessionTemplateModel(*session, db.UserDataModel{})
	handleHtmlResponse(r, w, templates.VotePlaylistButton(templateModel, playlist.ExternalUrls.Spotify))
}

func appSessionPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequestContext(r)
	if err != nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	spotifyClient, err := getSpotifyClientFromRequestContext(r)
	if err != nil {
		http.Error(w, "Spotify not found in context", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session := db.NewGameSessionDataModel()
	session.SetId(sessionId)
	err = session.GetById()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	var playlist *spotify.Playlist
	userPlaylist, err := user.GetPlaylistBySessionId(session.Id)
	if err == db.ErrUserPlaylistNotFound {
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

	templateModel := templates.NewSessionTemplateModel(*session, db.UserDataModel{})
	templates.VotePlaylistButton(templateModel, playlist.Uri)
	handleHtmlResponse(r, w, templates.VotePlaylistButton(templateModel, playlist.ExternalUrls.Spotify))
}

func appSessionSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequestContext(r)
	if err != nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session := db.NewGameSessionDataModel()
	session.SetId(sessionId)
	err = session.GetById()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	r.ParseForm()
	trackId := r.Form.Get("trackId")
	submission := db.NewSubmissionDataModel(user.GetId(), trackId)

	err = session.AddSubmission(*submission)
	if err != nil {
		if err == db.ErrSubmissionsMaxedOut {
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
	user, err := getUserFromRequestContext(r)
	if err != nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session := db.NewGameSessionDataModel()
	session.SetId(sessionId)
	err = session.GetById()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	var phase templates.SessionPhase
	path := r.URL.Path
	if strings.Contains(path, "submission") {
		phase = templates.SubmissionPhase
	} else if strings.Contains(path, "vote") {
		phase = templates.VotePhase
	} else {
		http.Error(w, "Invalid phase", http.StatusBadRequest)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	templates.SessionPhaseTimeLeft(templateModel, phase).Render(r.Context(), w)
}

func appSessionSubmissionDetailsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequestContext(r)
	if err != nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	spotifyClient, err := getSpotifyClientFromRequestContext(r)
	if err != nil {
		http.Error(w, "Spotify not found in context", http.StatusInternalServerError)
		return
	}

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

	session := db.NewGameSessionDataModel()
	session.SetId(sessionId)
	err = session.GetById()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	submission, err := session.GetUserSubmission(submissionId, user.GetId())
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
	user, err := getUserFromRequestContext(r)
	if err != nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

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

	session := db.NewGameSessionDataModel()
	session.SetId(sessionId)
	err = session.GetById()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	err = session.DeleteSubmission(submissionId, user.GetId())
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
	user, err := getUserFromRequestContext(r)
	if err != nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	spotifyClient, err := getSpotifyClientFromRequestContext(r)
	if err != nil {
		http.Error(w, "Spotify not found in context", http.StatusInternalServerError)
		return
	}

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

	session := db.NewGameSessionDataModel()
	session.SetId(sessionId)
	err = session.GetById()
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
	user, err := getUserFromRequestContext(r)
	if err != nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session := db.NewGameSessionDataModel()
	session.SetId(sessionId)
	err = session.GetById()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	r.ParseForm()
	submissionId := r.Form.Get("submissionId")

	session.AddVote(*db.NewVoteDataModel(user.GetId(), submissionId))

	err = session.Update()
	if err != nil {
		http.Error(w, "Failed to add vote", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	templates.LazyLoadVoteCandidate(templateModel, submissionId).Render(r.Context(), w)
}

func appSessionDeleteVoteHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequestContext(r)
	if err != nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

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

	session := db.NewGameSessionDataModel()
	session.SetId(sessionId)
	err = session.GetById()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	vote, err := session.GetVote(voteId, user.GetId())
	if err != nil {
		http.Error(w, "Failed to get vote", http.StatusInternalServerError)
		return
	}

	err = session.DeleteVote(voteId, user.GetId())
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

func appProfileHandler(w http.ResponseWriter, r *http.Request) {
	spotify, err := getSpotifyClientFromRequestContext(r)
	if err != nil {
		http.Error(w, "Spotify not found in context", http.StatusInternalServerError)
		return
	}

	profile, err := spotify.GetCurrentUserProfile()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to get current user profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(profile); err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
	}
}

func registerAppMux(parentMux *http.ServeMux) {
	appMux := http.NewServeMux()

	appMux.Handle("GET /", http.HandlerFunc(appLandingHandler))

	appMux.Handle("POST /session", http.HandlerFunc(createAppSessionHandler))
	appMux.Handle("GET /session/{sessionId}", http.HandlerFunc(appSessionHandler))
	appMux.Handle("POST /session/{sessionId}/tracks", handlerFuncWithMiddleware(appSessionTracksSearchHandler, withSpotify))
	appMux.Handle("POST /session/{sessionId}/playlist", handlerFuncWithMiddleware(appSessionCreatePlaylistHandler, withSpotify))
	appMux.Handle("GET /session/{sessionId}/playlist", handlerFuncWithMiddleware(appSessionPlaylistHandler, withSpotify))

	appMux.Handle("POST /session/{sessionId}/submission", handlerFuncWithMiddleware(appSessionSubmissionHandler))
	appMux.Handle("GET /session/{sessionId}/submission/time-left", handlerFuncWithMiddleware(appSessionTimeLeftHandler))
	appMux.Handle("GET /session/{sessionId}/submission/{submissionId}", handlerFuncWithMiddleware(appSessionSubmissionDetailsHandler, withSpotify))
	appMux.Handle("DELETE /session/{sessionId}/submission/{submissionId}", handlerFuncWithMiddleware(appSessionDeleteSubmissionHandler))
	appMux.Handle("GET /session/{sessionId}/submission/{submissionId}/candidate", handlerFuncWithMiddleware(appSessionSubmissionCandidateHandler, withSpotify))

	appMux.Handle("POST /session/{sessionId}/vote", handlerFuncWithMiddleware(appSessionVoteHandler))
	appMux.Handle("DELETE /session/{sessionId}/vote/{voteId}", handlerFuncWithMiddleware(appSessionDeleteVoteHandler))
	appMux.Handle("GET /session/{sessionId}/vote/time-left", handlerFuncWithMiddleware(appSessionTimeLeftHandler))

	appMux.Handle("GET /profile", handlerFuncWithMiddleware(appProfileHandler, withSpotify))

	appMuxWithMiddleware := applyMiddleware(appMux, enforceAuthentication)

	parentMux.Handle(appMuxPathPrefix+"/", http.StripPrefix(appMuxPathPrefix, appMuxWithMiddleware))
}
