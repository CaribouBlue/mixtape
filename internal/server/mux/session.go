package mux

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/CaribouBlue/mixtape/internal/core"
	"github.com/CaribouBlue/mixtape/internal/server/middleware"
	"github.com/CaribouBlue/mixtape/internal/server/response"
	"github.com/CaribouBlue/mixtape/internal/server/utils"
	serverUtils "github.com/CaribouBlue/mixtape/internal/server/utils"
	"github.com/CaribouBlue/mixtape/internal/templates"
)

type SessionMux struct {
	Mux[SessionMuxOpts, SessionMuxServices]
}

func (mux *SessionMux) Opts() MuxOpts {
	return mux.opts.MuxOpts
}

type SessionMuxOpts struct {
	MuxOpts
}

type SessionMuxServices struct {
	MuxServices
	SessionServiceInitializer MuxServiceInitializer[*SessionMux, *core.SessionService]
	sessionService            *core.SessionService
	MusicServiceInitializer   MuxServiceInitializer[*SessionMux, *core.MusicService]
	musicService              *core.MusicService
	UserService               *core.UserService
}

func (services *SessionMuxServices) SessionService() (*core.SessionService, error) {
	if services.sessionService == nil {
		return nil, errors.New("session service not initialized")
	}
	return services.sessionService, nil
}

func (services *SessionMuxServices) MusicService() (*core.MusicService, error) {
	if services.musicService == nil {
		return nil, errors.New("music service not initialized")
	}
	return services.musicService, nil
}

func NewSessionMux(opts SessionMuxOpts, services SessionMuxServices, mw []middleware.Middleware, children []ChildMux) *SessionMux {
	mux := &SessionMux{
		*NewMux(
			opts,
			services,
			children,
			mw,
		),
	}

	mux.BeforeEachRequest = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var err error

			mux.Services.musicService, err = mux.Services.MusicServiceInitializer(mux, r)
			if err != nil {
				response.HandleErrorResponse(w, "Failed to init mux", http.StatusInternalServerError, r, err)
				return
			}

			mux.Services.sessionService, err = mux.Services.SessionServiceInitializer(mux, r)
			if err != nil {
				response.HandleErrorResponse(w, "Failed to init mux", http.StatusInternalServerError, r, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	mux.Handle("GET /", http.HandlerFunc(mux.handlePageSessions))
	mux.Handle("POST /", http.HandlerFunc(mux.handleCreateSession))

	mux.Handle("GET /maker", http.HandlerFunc(mux.handlePageSessionMaker))

	mux.Handle("GET /{sessionId}", http.HandlerFunc(mux.handlePageSession))

	mux.Handle("GET /{sessionId}/phase-duration", http.HandlerFunc(mux.handleGetPhaseDuration))
	mux.Handle("GET /{sessionId}/submission-search", http.HandlerFunc(mux.handleSearchSubmissions))

	mux.Handle("POST /{sessionId}/player/me", http.HandlerFunc(mux.handleJoinSession))
	mux.Handle("POST /{sessionId}/player/me/finalize-submissions", http.HandlerFunc(mux.handleFinalizeSubmissions))
	mux.Handle("POST /{sessionId}/player/me/playlist", http.HandlerFunc(mux.handleCreatePlayerPlaylist))

	mux.Handle("POST /{sessionId}/candidate", http.HandlerFunc(mux.handleSubmitCandidate))
	mux.Handle("DELETE /{sessionId}/candidate/{candidateId}", http.HandlerFunc(mux.handleRemoveCandidate))
	mux.Handle("POST /{sessionId}/candidate/{candidateId}/vote", http.HandlerFunc(mux.handleCreateCandidateVote))
	mux.Handle("DELETE /{sessionId}/candidate/{candidateId}/vote", http.HandlerFunc(mux.handleDeleteCandidateVote))

	return mux
}

func (mux *SessionMux) handlePageSessions(w http.ResponseWriter, r *http.Request) {
	user, err := utils.ContextValue(r.Context(), utils.UserCtxKey)
	if err != nil {
		response.HandleErrorResponse(w, "Could not get user data", http.StatusUnauthorized, r, err)
		return
	}

	sessions, err := mux.Services.sessionService.GetSessionsListForUser(user.Id)
	if err != nil {
		response.HandleErrorResponse(w, "Failed to get sessions", http.StatusInternalServerError, r, err)
		return
	}

	component := templates.UserSessions(*sessions)
	response.HandleHtmlResponse(r, w, component)
}

func (mux *SessionMux) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	user, err := utils.ContextValue(r.Context(), utils.UserCtxKey)
	if err != nil {
		response.HandleErrorResponse(w, "Could not get user data", http.StatusUnauthorized, r, err)
		return
	}

	err = r.ParseForm()
	if err != nil {
		response.HandleErrorResponse(w, "Failed to parse form", http.StatusBadRequest, r, err)
		return
	}

	name := r.Form.Get("name")

	session, err := mux.Services.sessionService.CreateSession(core.NewSessionEntity(name, user.Id))
	if err != nil {
		response.HandleErrorResponse(w, "Failed to create session", http.StatusInternalServerError, r, err)
		return
	}

	response.HandleRedirect(w, r, fmt.Sprintf("/app/session/%d", session.Id))
}

func (mux *SessionMux) handlePageSessionMaker(w http.ResponseWriter, r *http.Request) {
	user, err := utils.ContextValue(r.Context(), utils.UserCtxKey)
	if err != nil {
		response.HandleErrorResponse(w, "Could not get user data", http.StatusUnauthorized, r, err)
		return
	}

	if !user.IsAdmin {
		response.HandleErrorResponse(w, "Forbidden", http.StatusForbidden, r, err)
		return
	}

	response.HandleHtmlResponse(r, w, templates.SessionMakerPage(*user))
}

func (mux *SessionMux) handlePageSession(w http.ResponseWriter, r *http.Request) {
	user, err := utils.ContextValue(r.Context(), utils.UserCtxKey)
	if err != nil {
		response.HandleErrorResponse(w, "Could not get user data", http.StatusUnauthorized, r, err)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		response.HandleErrorResponse(w, "Invalid session ID", http.StatusBadRequest, r, err)
		return
	}

	sessionView, err := mux.Services.sessionService.GetSessionView(sessionId, user.Id)
	if err != nil {
		response.HandleErrorResponse(w, "Failed to get session view", http.StatusInternalServerError, r, err)
		return
	}

	component := templates.SessionPage(*sessionView)
	response.HandleHtmlResponse(r, w, component)
}

func (mux *SessionMux) handleSearchSubmissions(w http.ResponseWriter, r *http.Request) {
	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		response.HandleErrorResponse(w, "Invalid session ID", http.StatusBadRequest, r, err)
		return
	}

	r.ParseForm()
	query := r.Form.Get("query")

	submissions, err := mux.Services.sessionService.SearchCandidateSubmissions(sessionId, query)
	if err != nil {
		response.HandleErrorResponse(w, "Failed to search tracks", http.StatusInternalServerError, r, err)
		return
	}

	response.HandleHtmlResponse(r, w, templates.CandidateSubmissionSearchResults(*submissions))
}

func (mux *SessionMux) handleSubmitCandidate(w http.ResponseWriter, r *http.Request) {
	user, err := utils.ContextValue(r.Context(), utils.UserCtxKey)
	if err != nil {
		response.HandleErrorResponse(w, "Could not get user data", http.StatusUnauthorized, r, err)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		response.HandleErrorResponse(w, "Invalid session ID", http.StatusBadRequest, r, err)
		return
	}

	r.ParseForm()
	trackId := r.Form.Get("trackId")

	submission, err := mux.Services.sessionService.SubmitCandidate(sessionId, user.Id, trackId)
	if err == core.ErrNoSubmissionsLeft {
		response.HandleErrorResponse(w, "No submissions left", http.StatusUnprocessableEntity, r, err)
		return
	} else if err == core.ErrDuplicateSubmission {
		response.HandleErrorResponse(w, "This song was already submitted", http.StatusUnprocessableEntity, r, err)
		return
	} else if err != nil {
		response.HandleErrorResponse(w, "Failed to add submission", http.StatusInternalServerError, r, err)
		return
	}

	session, err := mux.Services.sessionService.GetSessionView(sessionId, user.Id)
	if err != nil {
		response.HandleErrorResponse(w, "Failed to get session view", http.StatusInternalServerError, r, err)
		return
	}

	w.Header().Add("HX-Trigger", serverUtils.EventNewSubmission)
	response.HandleHtmlResponse(r, w, templates.AddSubmission(*session, *submission))
}

func (mux *SessionMux) handleJoinSession(w http.ResponseWriter, r *http.Request) {
	user, err := utils.ContextValue(r.Context(), utils.UserCtxKey)
	if err != nil {
		response.HandleErrorResponse(w, "Could not get user data", http.StatusUnauthorized, r, err)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		response.HandleErrorResponse(w, "Invalid session ID", http.StatusBadRequest, r, err)
		return
	}

	_, err = mux.Services.sessionService.JoinSession(sessionId, user.Id)
	if err != nil {
		response.HandleErrorResponse(w, "Failed to join session", http.StatusInternalServerError, r, err)
		return
	}

	sessionView, err := mux.Services.sessionService.GetSessionView(sessionId, user.Id)
	if err != nil {
		response.HandleErrorResponse(w, "Failed to get session view", http.StatusInternalServerError, r, err)
		return
	}

	response.HandleHtmlResponse(r, w, templates.SessionPage(*sessionView))
}

func (mux *SessionMux) handleFinalizeSubmissions(w http.ResponseWriter, r *http.Request) {
	user, err := utils.ContextValue(r.Context(), utils.UserCtxKey)
	if err != nil {
		response.HandleErrorResponse(w, "Could not get user data", http.StatusUnauthorized, r, err)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		response.HandleErrorResponse(w, "Invalid session ID", http.StatusBadRequest, r, err)
		return
	}

	err = mux.Services.sessionService.FinalizePlayerSubmissions(sessionId, user.Id)
	if err != nil {
		response.HandleErrorResponse(w, "Failed to finalize submissions", http.StatusInternalServerError, r, err)
		return
	}

	sessionView, err := mux.Services.sessionService.GetSessionView(sessionId, user.Id)
	if err != nil {
		response.HandleErrorResponse(w, "Failed to get session view", http.StatusInternalServerError, r, err)
		return
	}

	response.HandleHtmlResponse(r, w, templates.SessionPage(*sessionView))
}

func (mux *SessionMux) handleCreatePlayerPlaylist(w http.ResponseWriter, r *http.Request) {
	user, err := utils.ContextValue(r.Context(), utils.UserCtxKey)
	if err != nil {
		response.HandleErrorResponse(w, "Could not get user data", http.StatusUnauthorized, r, err)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		response.HandleErrorResponse(w, "Invalid session ID", http.StatusBadRequest, r, err)
		return
	}

	player, err := mux.Services.sessionService.CreatePlayerPlaylist(sessionId, user.Id)
	if err != nil {
		response.HandleErrorResponse(w, "Failed to create player playlist", http.StatusInternalServerError, r, err)
		return
	}

	response.HandleHtmlResponse(r, w, templates.PlaylistButton(sessionId, player.PlaylistUrl))
}

func (mux *SessionMux) handleGetPhaseDuration(w http.ResponseWriter, r *http.Request) {
	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		response.HandleErrorResponse(w, "Invalid session ID", http.StatusBadRequest, r, err)
		return
	}

	session, err := mux.Services.sessionService.GetSessionData(sessionId)
	if err != nil {
		response.HandleErrorResponse(w, "Failed to get session", http.StatusInternalServerError, r, err)
		return
	}

	templates.SessionPhaseDuration(*session).Render(r.Context(), w)
}

func (mux *SessionMux) handleRemoveCandidate(w http.ResponseWriter, r *http.Request) {
	user, err := utils.ContextValue(r.Context(), utils.UserCtxKey)
	if err != nil {
		response.HandleErrorResponse(w, "Could not get user data", http.StatusUnauthorized, r, err)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		response.HandleErrorResponse(w, "Invalid session ID", http.StatusBadRequest, r, err)
		return
	}

	candidateId, err := strconv.ParseInt(r.PathValue("candidateId"), 10, 64)
	if err != nil {
		response.HandleErrorResponse(w, "Invalid candidate ID", http.StatusBadRequest, r, err)
		return
	}

	err = mux.Services.sessionService.RemoveCandidate(sessionId, user.Id, candidateId)
	if err != nil {
		response.HandleErrorResponse(w, "Failed to delete submission", http.StatusInternalServerError, r, err)
		return
	}

	session, err := mux.Services.sessionService.GetSessionView(sessionId, user.Id)
	if err != nil {
		response.HandleErrorResponse(w, "Failed to get session view", http.StatusInternalServerError, r, err)
		return
	}

	w.Header().Add("HX-Trigger", serverUtils.EventDeleteSubmission)
	response.HandleHtmlResponse(r, w, templates.RemoveSubmission(*session))
}

func (mux *SessionMux) handleCreateCandidateVote(w http.ResponseWriter, r *http.Request) {
	user, err := utils.ContextValue(r.Context(), utils.UserCtxKey)
	if err != nil {
		response.HandleErrorResponse(w, "Could not get user data", http.StatusUnauthorized, r, err)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		response.HandleErrorResponse(w, "Invalid session ID", http.StatusBadRequest, r, err)
		return
	}

	r.ParseForm()
	candidateId, err := strconv.ParseInt(r.PathValue("candidateId"), 10, 64)
	if err != nil {
		response.HandleErrorResponse(w, "Invalid candidate ID", http.StatusBadRequest, r, err)
		return
	}

	candidate, err := mux.Services.sessionService.VoteForCandidate(sessionId, user.Id, candidateId)
	if err == core.ErrNoVotesLeft {
		w.Header().Add("HX-Reswap", "innerHTML")
		response.HandleErrorResponse(w, "No votes left", http.StatusUnprocessableEntity, r, err)
		return
	} else if err != nil {
		response.HandleErrorResponse(w, "Failed to add vote", http.StatusInternalServerError, r, err)
		return
	}

	w.Header().Add("HX-Trigger", serverUtils.EventNewVote)
	templates.CandidateBallot(*candidate, true).Render(r.Context(), w)
}

func (mux *SessionMux) handleDeleteCandidateVote(w http.ResponseWriter, r *http.Request) {
	user, err := utils.ContextValue(r.Context(), utils.UserCtxKey)
	if err != nil {
		response.HandleErrorResponse(w, "Could not get user data", http.StatusUnauthorized, r, err)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		response.HandleErrorResponse(w, "Invalid session ID", http.StatusBadRequest, r, err)
		return
	}

	candidateId, err := strconv.ParseInt(r.PathValue("candidateId"), 10, 64)
	if err != nil {
		response.HandleErrorResponse(w, "Invalid candidate ID", http.StatusBadRequest, r, err)
		return
	}

	candidate, err := mux.Services.sessionService.RemoveVoteForCandidate(sessionId, user.Id, candidateId)
	if err != nil {
		response.HandleErrorResponse(w, "Failed to remove vote", http.StatusInternalServerError, r, err)
		return
	}

	w.Header().Add("HX-Trigger", serverUtils.EventDeleteVote)
	templates.CandidateBallot(*candidate, true).Render(r.Context(), w)
}
