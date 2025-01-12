package mux

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/CaribouBlue/top-spot/internal/music"
	"github.com/CaribouBlue/top-spot/internal/server/middleware"
	"github.com/CaribouBlue/top-spot/internal/server/utils"
	"github.com/CaribouBlue/top-spot/internal/session"
	"github.com/CaribouBlue/top-spot/internal/templates"
	"github.com/CaribouBlue/top-spot/internal/user"
)

type SessionMux struct {
	*http.ServeMux
	sessionService session.SessionService
	musicService   music.MusicService
}

func NewSessionMux(sessionService session.SessionService, musicService music.MusicService) *SessionMux {
	mux := &SessionMux{
		http.NewServeMux(),
		sessionService,
		musicService,
	}
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
	user := r.Context().Value(middleware.UserCtxKey).(*user.User)

	sessions, err := mux.sessionService.GetAll()
	if err != nil {
		log.Print(err)
		http.Error(w, "Failed to get sessions", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewHomeTemplateModel(user, sessions)
	component := templates.Home(templateModel)
	utils.HandleHtmlResponse(r, w, component)
}

func (mux *SessionMux) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	session := &session.Session{}
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(session)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	err = mux.sessionService.Create(session)
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
	user := r.Context().Value(middleware.UserCtxKey).(*user.User)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session, err := mux.sessionService.GetOne(sessionId)
	if err != nil {
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
		component := templates.SessionPage(templateModel, "")
		utils.HandleHtmlResponse(r, w, component)
	}
}

func (mux *SessionMux) handleCreateSessionTrack(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*user.User)
	err := mux.musicService.Authenticate(user)
	if err != nil {
		http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session, err := mux.sessionService.GetOne(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	r.ParseForm()
	query := r.Form.Get("query")

	templateModel := templates.NewSessionTemplateModel(*session, *user)

	if query != "" {
		searchResults, err := mux.musicService.SearchTracks(query)
		if err != nil {
			http.Error(w, "Failed to search Spotify", http.StatusInternalServerError)
			return
		}

		templateModel.SearchResult = searchResults
	}

	templates.SubmissionSearchBar(templateModel, "").Render(r.Context(), w)
}

func (mux *SessionMux) handleCreateSessionPlaylist(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*user.User)
	err := mux.musicService.Authenticate(user)
	if err != nil {
		http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session, err := mux.sessionService.GetOne(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	playlist := &music.Playlist{
		Name: fmt.Sprintf("Top Spot Session: %s", session.Name),
	}
	trackIds := make([]string, len(session.Submissions))
	for i, submission := range session.Submissions {
		trackIds[i] = submission.TrackId
	}
	err = mux.musicService.CreatePlaylist(playlist, trackIds)
	if err != nil {
		http.Error(w, "Failed to create playlist", http.StatusInternalServerError)
		return
	}

	session, err = mux.sessionService.AddPlaylist(sessionId, playlist.Id, user.Id)
	if err != nil {
		http.Error(w, "Failed to add playlist to session", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	utils.HandleHtmlResponse(r, w, templates.PlaylistButton(templateModel, *playlist))
}

func (mux *SessionMux) handleGetSessionPlaylist(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*user.User)
	err := mux.musicService.Authenticate(user)
	if err != nil {
		http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	sesh, err := mux.sessionService.GetOne(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	var playlist *music.Playlist
	sessionPlaylist, err := mux.sessionService.GetPlaylist(sessionId, user.Id)
	if err == session.ErrPlaylistNotFound {
		playlist = &music.Playlist{}
	} else if err != nil {
		http.Error(w, "Failed to get playlist from session", http.StatusInternalServerError)
		return
	} else {
		playlist, err = mux.musicService.GetPlaylist(sessionPlaylist.Id)
		if err != nil {
			http.Error(w, "Failed to get playlist", http.StatusInternalServerError)
			return
		}
	}

	templateModel := templates.NewSessionTemplateModel(*sesh, *user)
	utils.HandleHtmlResponse(r, w, templates.PlaylistButton(templateModel, *playlist))
}

func (mux *SessionMux) handleCreateSessionSubmission(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*user.User)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	sesh, err := mux.sessionService.GetOne(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	r.ParseForm()
	trackId := r.Form.Get("trackId")
	submission := &session.Submission{
		UserId:  user.Id,
		TrackId: trackId,
	}
	sesh, err = mux.sessionService.AddSubmission(sesh.Id, submission)
	if err != nil {
		http.Error(w, "Failed to add submission", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*sesh, *user)
	templates.NewSubmission(templateModel, *submission).Render(r.Context(), w)
}

func (mux *SessionMux) handleGetSessionPhaseDuration(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*user.User)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session, err := mux.sessionService.GetOne(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	templates.SessionPhaseDuration(templateModel).Render(r.Context(), w)
}

func (mux *SessionMux) handleGetSessionSubmission(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*user.User)
	err := mux.musicService.Authenticate(user)
	if err != nil {
		http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
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

	session, err := mux.sessionService.GetOne(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	submission, err := mux.sessionService.GetSubmission(sessionId, submissionId)
	if err != nil {
		http.Error(w, "Failed to get submission", http.StatusInternalServerError)
		return
	}

	track, err := mux.musicService.GetTrack(submission.TrackId)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	templates.SubmissionItem(templateModel, *submission, *track).Render(r.Context(), w)
}

func (mux *SessionMux) handleDeleteSessionSubmission(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*user.User)

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

	session, err := mux.sessionService.RemoveSubmission(sessionId, submissionId)
	if err != nil {
		http.Error(w, "Failed to delete submission", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	templates.DeleteSubmission(templateModel).Render(r.Context(), w)
}

func (mux *SessionMux) handleGetSessionSubmissionCandidate(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*user.User)

	err := mux.musicService.Authenticate(user)
	if err != nil {
		http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
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

	session, err := mux.sessionService.GetOne(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	submission, err := mux.sessionService.GetSubmission(sessionId, submissionId)
	if err != nil {
		http.Error(w, "Failed to get submission", http.StatusInternalServerError)
		return
	}

	track, err := mux.musicService.GetTrack(submission.TrackId)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	templates.VoteCandidate(templateModel, *submission, *track).Render(r.Context(), w)
}

func (mux *SessionMux) handleCreateSessionVote(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*user.User)

	err := mux.musicService.Authenticate(user)
	if err != nil {
		http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	r.ParseForm()
	submissionId := r.Form.Get("submissionId")
	submission, err := mux.sessionService.GetSubmission(sessionId, submissionId)
	if err != nil {
		http.Error(w, "Failed to get submission", http.StatusInternalServerError)
		return
	}

	track, err := mux.musicService.GetTrack(submission.TrackId)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
		return
	}

	vote := &session.Vote{
		UserId:       user.Id,
		SubmissionId: submissionId,
	}
	sesh, err := mux.sessionService.AddVote(sessionId, vote)
	if err != nil {
		http.Error(w, "Failed to add vote", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*sesh, *user)
	templates.VoteCandidate(templateModel, *submission, *track).Render(r.Context(), w)
}

func (mux *SessionMux) handleDeleteSessionVote(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*user.User)

	err := mux.musicService.Authenticate(user)
	if err != nil {
		http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
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

	vote, err := mux.sessionService.GetVote(sessionId, voteId)
	if err != nil {
		http.Error(w, "Failed to get vote", http.StatusInternalServerError)
		return
	}

	submission, err := mux.sessionService.GetSubmission(sessionId, vote.SubmissionId)
	if err != nil {
		http.Error(w, "Failed to get submission", http.StatusInternalServerError)
		return
	}

	track, err := mux.musicService.GetTrack(submission.TrackId)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
		return
	}

	session, err := mux.sessionService.RemoveVote(sessionId, voteId)
	if err != nil {
		http.Error(w, "Failed to delete vote", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	templates.VoteCandidate(templateModel, *submission, *track).Render(r.Context(), w)
}
