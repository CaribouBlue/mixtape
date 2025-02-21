package mux

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/CaribouBlue/top-spot/internal/core"
	"github.com/CaribouBlue/top-spot/internal/server/middleware"
	serverUtils "github.com/CaribouBlue/top-spot/internal/server/utils"
	"github.com/CaribouBlue/top-spot/internal/templates"
)

type SessionMux struct {
	*http.ServeMux
	Opts       SessionMuxOpts
	Repos      SessionMuxRepos
	services   sessionMuxServices
	Middleware []middleware.Middleware
}

type SessionMuxOpts struct {
	PathPrefix string
}

type SessionMuxRepos struct {
	SessionRepo      core.SessionRepository
	MusicRepoFactory serverUtils.RequestBasedFactory[core.MusicRepository]
	UserRepo         core.UserRepository
}

type sessionMuxServices struct {
	SessionService *core.SessionService
	MusicService   *core.MusicService
	UserService    *core.UserService
}

func NewSessionMux(opts SessionMuxOpts, repos SessionMuxRepos, middleware []middleware.Middleware) *SessionMux {
	mux := &SessionMux{
		http.NewServeMux(),
		opts,
		repos,
		sessionMuxServices{},
		middleware,
	}

	mux.Handle("GET /", http.HandlerFunc(mux.handlePageSessions))
	mux.Handle("POST /", http.HandlerFunc(mux.handleCreateSession))

	mux.Handle("GET /maker", http.HandlerFunc(mux.handlePageSessionMaker))

	mux.Handle("GET /{sessionId}", http.HandlerFunc(mux.handlePageSession))

	mux.Handle("GET /{sessionId}/phase-duration", http.HandlerFunc(mux.handleGetPhaseDuration))
	mux.Handle("GET /{sessionId}/submission-search", http.HandlerFunc(mux.handleSearchSubmissions))
	mux.Handle("GET /{sessionId}/submission-counter", http.HandlerFunc(mux.handleGetSubmissionCounter))
	mux.Handle("GET /{sessionId}/vote-counter", http.HandlerFunc(mux.handleGetVoteCounter))

	mux.Handle("POST /{sessionId}/candidate", http.HandlerFunc(mux.handleSubmitCandidate))
	mux.Handle("DELETE /{sessionId}/candidate/{candidateId}", http.HandlerFunc(mux.handleRemoveCandidate))
	mux.Handle("POST /{sessionId}/candidate/{candidateId}/vote", http.HandlerFunc(mux.handleCreateCandidateVote))
	mux.Handle("DELETE /{sessionId}/candidate/{candidateId}/vote", http.HandlerFunc(mux.handleDeleteCandidateVote))

	return mux
}

func (mux *SessionMux) beforeEveryRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		musicRepo := mux.Repos.MusicRepoFactory(r)
		userRepo := mux.Repos.UserRepo
		sessionRepo := mux.Repos.SessionRepo

		mux.services.MusicService = core.NewMusicService(musicRepo)
		mux.services.UserService = core.NewUserService(userRepo)
		mux.services.SessionService = core.NewSessionService(sessionRepo, mux.services.UserService, mux.services.MusicService)

		next.ServeHTTP(w, r)
	})
}

func (mux *SessionMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	middleware.Apply(mux.ServeMux, append(mux.Middleware, mux.beforeEveryRequest)...).ServeHTTP(w, r)
}

func (mux *SessionMux) handlePageSessions(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)

	sessions, err := mux.services.SessionService.GetSessionsList()
	if err != nil {
		http.Error(w, "Failed to get sessions", http.StatusInternalServerError)
		return
	}

	component := templates.UserSessions(*user, *sessions)
	serverUtils.HandleHtmlResponse(r, w, component)
}

func (mux *SessionMux) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.Form.Get("name")

	session, err := mux.services.SessionService.CreateSession(core.NewSessionEntity(name, u.Id))
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	serverUtils.HandleRedirect(w, r, fmt.Sprintf("/app/session/%d", session.Id))
}

func (mux *SessionMux) handlePageSessionMaker(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)

	if !user.IsAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	serverUtils.HandleHtmlResponse(r, w, templates.SessionMakerPage(*user))
}

func (mux *SessionMux) handlePageSession(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	sessionView, err := mux.services.SessionService.GetSessionView(sessionId, user.Id)
	if err != nil {
		log.Default().Println("Failed to get session view: ", err)
		http.Error(w, "Failed to get session view", http.StatusInternalServerError)
		return
	}

	component := templates.SessionPage(*sessionView, *user)
	serverUtils.HandleHtmlResponse(r, w, component)
}

func (mux *SessionMux) handleSearchSubmissions(w http.ResponseWriter, r *http.Request) {
	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	r.ParseForm()
	query := r.Form.Get("query")

	submissions, err := mux.services.SessionService.SearchCandidateSubmissions(sessionId, query)
	if err != nil {
		http.Error(w, "Failed to search tracks", http.StatusInternalServerError)
		return
	}

	serverUtils.HandleHtmlResponse(r, w, templates.CandidateSubmissionSearchResults(*submissions))
}

func (mux *SessionMux) handleSubmitCandidate(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	r.ParseForm()
	trackId := r.Form.Get("trackId")

	submission, err := mux.services.SessionService.SubmitCandidate(sessionId, user.Id, trackId)
	if err == core.ErrNoSubmissionsLeft {
		http.Error(w, "No submissions left", http.StatusUnprocessableEntity)
		return
	} else if err == core.ErrDuplicateSubmission {
		http.Error(w, "This song was already submitted", http.StatusUnprocessableEntity)
		return
	} else if err != nil {
		http.Error(w, "Failed to add submission", http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Trigger", serverUtils.EventNewSubmission)
	serverUtils.HandleHtmlResponse(r, w, templates.AddSubmission(*submission))
}

func (mux *SessionMux) handleGetSubmissionCounter(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	sessionView, err := mux.services.SessionService.GetSessionView(sessionId, user.Id)
	if err != nil {
		http.Error(w, "Failed to get session view", http.StatusInternalServerError)
		return
	}

	serverUtils.HandleHtmlResponse(r, w, templates.SubmissionCounter(*sessionView))
}

func (mux *SessionMux) handleGetVoteCounter(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	sessionView, err := mux.services.SessionService.GetSessionView(sessionId, user.Id)
	if err != nil {
		http.Error(w, "Failed to get session view", http.StatusInternalServerError)
		return
	}

	serverUtils.HandleHtmlResponse(r, w, templates.VoteCounter(*sessionView))
}

func (mux *SessionMux) handleGetPhaseDuration(w http.ResponseWriter, r *http.Request) {
	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session, err := mux.services.SessionService.GetSessionData(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	templates.SessionPhaseDuration(*session).Render(r.Context(), w)
}

func (mux *SessionMux) handleRemoveCandidate(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	candidateId, err := strconv.ParseInt(r.PathValue("candidateId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid candidate ID", http.StatusBadRequest)
		return
	}

	err = mux.services.SessionService.RemoveCandidate(sessionId, user.Id, candidateId)
	if err != nil {
		http.Error(w, "Failed to delete submission", http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Trigger", serverUtils.EventDeleteSubmission)
	w.WriteHeader(http.StatusOK)
}

func (mux *SessionMux) handleCreateCandidateVote(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	r.ParseForm()
	candidateId, err := strconv.ParseInt(r.PathValue("candidateId"), 10, 64)
	if err != nil {
		log.Default().Println("Failed to parse candidate ID: ", err)
		http.Error(w, "Invalid candidate ID", http.StatusBadRequest)
		return
	}

	candidate, err := mux.services.SessionService.VoteForCandidate(sessionId, user.Id, candidateId)
	if err == core.ErrNoVotesLeft {
		w.Header().Add("HX-Reswap", "innerHTML")
		http.Error(w, "No votes left", http.StatusUnprocessableEntity)
		return
	} else if err != nil {
		http.Error(w, "Failed to add vote", http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Trigger", serverUtils.EventNewVote)
	templates.CandidateBallot(*candidate, true).Render(r.Context(), w)
}

func (mux *SessionMux) handleDeleteCandidateVote(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	candidateId, err := strconv.ParseInt(r.PathValue("candidateId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid candidate ID", http.StatusBadRequest)
		return
	}

	candidate, err := mux.services.SessionService.RemoveVoteForCandidate(sessionId, user.Id, candidateId)
	if err != nil {
		http.Error(w, "Failed to remove vote", http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Trigger", serverUtils.EventDeleteVote)
	templates.CandidateBallot(*candidate, true).Render(r.Context(), w)
}
