package mux

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/CaribouBlue/top-spot/internal/core"
	"github.com/CaribouBlue/top-spot/internal/server/middleware"
	serverUtils "github.com/CaribouBlue/top-spot/internal/server/utils"
	"github.com/CaribouBlue/top-spot/internal/templates"
	"github.com/CaribouBlue/top-spot/internal/utils"
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

	mux.Handle("GET /{sessionId}/tracks/search", http.HandlerFunc(mux.handleSessionTracksSearch))

	mux.Handle("POST /{sessionId}/playlist", http.HandlerFunc(mux.handleCreateSessionPlaylist))
	mux.Handle("GET /{sessionId}/playlist", http.HandlerFunc(mux.handleGetSessionPlaylist))

	mux.Handle("GET /{sessionId}/phase-duration", http.HandlerFunc(mux.handleGetSessionPhaseDuration))

	mux.Handle("POST /{sessionId}/submission", http.HandlerFunc(mux.handleCreateSessionSubmission))
	mux.Handle("GET /{sessionId}/submission-counter", http.HandlerFunc(mux.handleGetSessionSubmissionCounter))
	mux.Handle("GET /{sessionId}/vote-counter", http.HandlerFunc(mux.handleGetSessionVoteCounter))

	mux.Handle("GET /{sessionId}/submission/{submissionId}", http.HandlerFunc(mux.handleGetSessionSubmission))
	mux.Handle("DELETE /{sessionId}/submission/{submissionId}", http.HandlerFunc(mux.handleDeleteSessionSubmission))

	mux.Handle("GET /{sessionId}/submission/{submissionId}/candidate", http.HandlerFunc(mux.handleGetSessionSubmissionCandidate))

	mux.Handle("GET /{sessionId}/result/{resultId}", http.HandlerFunc(mux.handleGetSessionResult))

	mux.Handle("POST /{sessionId}/vote", http.HandlerFunc(mux.handleCreateSessionVote))

	mux.Handle("DELETE /{sessionId}/vote/{voteId}", http.HandlerFunc(mux.handleDeleteSessionVote))

	return mux
}

func (mux *SessionMux) beforeEveryRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		musicRepo := mux.Repos.MusicRepoFactory(r)
		userRepo := mux.Repos.UserRepo
		sessionRepo := mux.Repos.SessionRepo

		mux.services.MusicService = core.NewMusicService(musicRepo)
		mux.services.UserService = core.NewUserService(userRepo)
		mux.services.SessionService = core.NewSessionService(sessionRepo, mux.services.MusicService)

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
	sessionValues := utils.Map(*sessions, func(session core.SessionEntity) core.SessionEntity { return session })

	component := templates.UserSessions(*user, sessionValues)
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
		http.Error(w, "Failed to get session view", http.StatusInternalServerError)
		return
	}

	component := templates.SessionPage(*sessionView, *user)
	serverUtils.HandleHtmlResponse(r, w, component)
}

func (mux *SessionMux) handleSessionTracksSearch(w http.ResponseWriter, r *http.Request) {
	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	r.ParseForm()
	query := r.Form.Get("query")

	candidates, err := mux.services.SessionService.SearchSubmissionTracks(sessionId, query)
	if err != nil {
		http.Error(w, "Failed to search tracks", http.StatusInternalServerError)
		return
	}

	// for _, c := range *candidates {
	// 	log.Default().Println("track: ", c.Track.Id, c.Track.Name)
	// }

	serverUtils.HandleHtmlResponse(r, w, templates.SubmissionSearchResults(*candidates))
}

func (mux *SessionMux) handleCreateSessionPlaylist(w http.ResponseWriter, r *http.Request) {
	// user := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)
	// err := mux.Services.MusicService.Authenticate(user)
	// if err != nil {
	// 	http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
	// 	return
	// }

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session, err := mux.services.SessionService.GetSession(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	submissions, err := mux.services.SessionService.GetAllSubmissions(sessionId)
	if err != nil {
		http.Error(w, "Failed to get submissions", http.StatusInternalServerError)
		return
	}

	trackIds := make([]string, len(*submissions))
	for i, submission := range *submissions {
		trackIds[i] = submission.TrackId
	}

	playlist, err := mux.services.MusicService.CreatePlaylist(fmt.Sprintf("Top Spot Session: %s", session.Name), trackIds)
	if err != nil {
		http.Error(w, "Failed to create playlist", http.StatusInternalServerError)
		return
	}

	// playlist := &music.Playlist{
	// 	Name: fmt.Sprintf("Top Spot Session: %s", session.Name),
	// }
	// trackIds := make([]string, len(session.Submissions))
	// for i, submission := range session.Submissions {
	// 	trackIds[i] = submission.TrackId
	// }
	// err = mux.Services.MusicService.CreatePlaylist(playlist, trackIds)
	// if err != nil {
	// 	http.Error(w, "Failed to create playlist", http.StatusInternalServerError)
	// 	return
	// }

	// session, err = mux.Services.SessionService.AddPlaylist(sessionId, playlist.Id, user.Id)
	// if err != nil {
	// 	http.Error(w, "Failed to add playlist to session", http.StatusInternalServerError)
	// 	return
	// }

	serverUtils.HandleHtmlResponse(r, w, templates.PlaylistButton(*session, *playlist))
}

func (mux *SessionMux) handleGetSessionPlaylist(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)
	// err := mux.Services.MusicService.Authenticate(user)
	// if err != nil {
	// 	http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
	// 	return
	// }

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	s, err := mux.services.SessionService.GetSession(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	var playlist *core.PlaylistEntity
	sessionPlaylist, err := mux.services.SessionService.GetUserPlaylist(sessionId, user.Id)
	if err == core.ErrPlaylistNotFound {
		playlist = &core.PlaylistEntity{}
	} else if err != nil {
		http.Error(w, "Failed to get playlist from session", http.StatusInternalServerError)
		return
	} else {
		playlist, err = mux.services.MusicService.GetPlaylistById(sessionPlaylist.PlaylistId)
		if err != nil {
			http.Error(w, "Failed to get playlist", http.StatusInternalServerError)
			return
		}
	}

	serverUtils.HandleHtmlResponse(r, w, templates.PlaylistButton(*s, *playlist))
}

func (mux *SessionMux) handleCreateSessionSubmission(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session, err := mux.services.SessionService.GetSession(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	r.ParseForm()
	trackId := r.Form.Get("trackId")
	submission, err := mux.services.SessionService.AddUserSubmission(session.Id, user.Id, trackId)
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

	track, err := mux.services.MusicService.GetTrackById(submission.TrackId)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Trigger", serverUtils.EventNewSubmission)
	serverUtils.HandleHtmlResponse(r, w, templates.AddSubmission(*session, *user, *submission, *track))
}

func (mux *SessionMux) handleGetSessionSubmissionCounter(w http.ResponseWriter, r *http.Request) {
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

func (mux *SessionMux) handleGetSessionVoteCounter(w http.ResponseWriter, r *http.Request) {
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

func (mux *SessionMux) handleGetSessionPhaseDuration(w http.ResponseWriter, r *http.Request) {
	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session, err := mux.services.SessionService.GetSession(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	templates.SessionPhaseDuration(*session).Render(r.Context(), w)
}

func (mux *SessionMux) handleGetSessionSubmission(w http.ResponseWriter, r *http.Request) {
	// user := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)
	// err := mux.Services.MusicService.Authenticate(user)
	// if err != nil {
	// 	http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
	// 	return
	// }

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	submissionIdParam := r.PathValue("submissionId")
	submissionId, err := strconv.ParseInt(submissionIdParam, 10, 64)
	if err != nil {
		http.Error(w, "Invalid submission ID", http.StatusBadRequest)
		return
	}

	session, err := mux.services.SessionService.GetSession(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	submission, err := mux.services.SessionService.GetSubmissionById(sessionId, submissionId)
	if err != nil {
		http.Error(w, "Failed to get submission", http.StatusInternalServerError)
		return
	}

	track, err := mux.services.MusicService.GetTrackById(submission.TrackId)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
		return
	}

	templates.SubmissionItem(*session, *submission, *track).Render(r.Context(), w)
}

func (mux *SessionMux) handleDeleteSessionSubmission(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	submissionIdParam := r.PathValue("submissionId")
	submissionId, err := strconv.ParseInt(submissionIdParam, 10, 64)
	if err != nil {
		http.Error(w, "Invalid submission ID", http.StatusBadRequest)
		return
	}

	err = mux.services.SessionService.RemoveUserSubmission(sessionId, user.Id, submissionId)
	if err != nil {
		http.Error(w, "Failed to delete submission", http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Trigger", serverUtils.EventDeleteSubmission)
	w.WriteHeader(http.StatusOK)
}

func (mux *SessionMux) handleGetSessionSubmissionCandidate(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)

	// err := mux.Services.MusicService.Authenticate(user)
	// if err != nil {
	// 	http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
	// 	return
	// }

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	submissionIdParam := r.PathValue("submissionId")
	submissionId, err := strconv.ParseInt(submissionIdParam, 10, 64)
	if err != nil {
		http.Error(w, "Invalid submission ID", http.StatusBadRequest)
		return
	}

	session, err := mux.services.SessionService.GetSession(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	candidate, err := mux.services.SessionService.GetUserCandidate(sessionId, user.Id, submissionId)
	if err != nil {
		http.Error(w, "Failed to get submission", http.StatusInternalServerError)
		return
	}
	track, err := mux.services.MusicService.GetTrackById(candidate.Submission.TrackId)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
		return
	}

	templates.VoteCandidate(*session, *user, *candidate, *track).Render(r.Context(), w)
}

func (mux *SessionMux) handleCreateSessionVote(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)

	// err := mux.Services.MusicService.Authenticate(user)
	// if err != nil {
	// 	http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
	// 	return
	// }

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session, err := mux.services.SessionService.GetSession(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	r.ParseForm()
	submissionIdParam := r.Form.Get("submissionId")
	submissionId, err := strconv.ParseInt(submissionIdParam, 10, 64)
	if err != nil {
		http.Error(w, "Invalid submission ID", http.StatusBadRequest)
		return
	}

	candidate, err := mux.services.SessionService.VoteForCandidate(sessionId, user.Id, submissionId)
	if err == core.ErrNoVotesLeft {
		w.Header().Add("HX-Reswap", "innerHTML")
		http.Error(w, "No votes left", http.StatusUnprocessableEntity)
		return
	} else if err != nil {
		http.Error(w, "Failed to add vote", http.StatusInternalServerError)
		return
	}
	track, err := mux.services.MusicService.GetTrackById(candidate.Submission.TrackId)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Trigger", serverUtils.EventNewVote)
	templates.VoteCandidate(*session, *user, *candidate, *track).Render(r.Context(), w)
}

func (mux *SessionMux) handleDeleteSessionVote(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)

	// err := mux.Services.MusicService.Authenticate(user)
	// if err != nil {
	// 	http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
	// 	return
	// }

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session, err := mux.services.SessionService.GetSession(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	submissionIdParam := r.PathValue("submissionId")
	submissionId, err := strconv.ParseInt(submissionIdParam, 10, 64)
	if err != nil {
		http.Error(w, "Invalid submission ID", http.StatusBadRequest)
		return
	}

	candidate, err := mux.services.SessionService.RemoveVoteForCandidate(sessionId, user.Id, submissionId)
	if err != nil {
		http.Error(w, "Failed to remove vote", http.StatusInternalServerError)
		return
	}
	track, err := mux.services.MusicService.GetTrackById(candidate.Submission.TrackId)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Trigger", serverUtils.EventDeleteVote)
	templates.VoteCandidate(*session, *user, *candidate, *track).Render(r.Context(), w)
}

func (mux *SessionMux) handleGetSessionResult(w http.ResponseWriter, r *http.Request) {
	// user := r.Context().Value(serverUtils.UserCtxKey).(*core.UserEntity)

	// err := mux.Services.MusicService.Authenticate(user)
	// if err != nil {
	// 	http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
	// 	return
	// }

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	resultId, err := strconv.ParseInt(r.PathValue("resultId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid result ID", http.StatusBadRequest)
		return
	}

	session, err := mux.services.SessionService.GetSession(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	results, err := mux.services.SessionService.GetResults(sessionId)
	if err != nil {
		http.Error(w, "Failed to get result", http.StatusInternalServerError)
		return
	}

	var result *core.ResultDto
	for _, r := range *results {
		if r.Submission.Id == resultId {
			result = &r
			break
		}
	}

	if result == nil {
		http.Error(w, "Result not found", http.StatusNotFound)
		return
	}

	submission, err := mux.services.SessionService.GetSubmissionById(sessionId, result.Submission.Id)
	if err != nil {
		http.Error(w, "Failed to get submission", http.StatusInternalServerError)
		return
	}
	track, err := mux.services.MusicService.GetTrackById(submission.TrackId)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
		return
	}

	owner, err := mux.services.UserService.GetUserById(submission.UserId)
	if err != nil {
		http.Error(w, "Failed to get owner", http.StatusInternalServerError)
		return
	}

	templates.Result(*session, *result, *submission, *track, *owner).Render(r.Context(), w)
}
