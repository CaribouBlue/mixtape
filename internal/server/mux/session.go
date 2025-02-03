package mux

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/CaribouBlue/top-spot/internal/entities/music"
	"github.com/CaribouBlue/top-spot/internal/entities/session"
	"github.com/CaribouBlue/top-spot/internal/entities/user"
	"github.com/CaribouBlue/top-spot/internal/server/middleware"
	serverUtils "github.com/CaribouBlue/top-spot/internal/server/utils"
	"github.com/CaribouBlue/top-spot/internal/templates"
	"github.com/CaribouBlue/top-spot/internal/utils"
)

type SessionMux struct {
	*http.ServeMux
	Opts       SessionMuxOpts
	Services   SessionMuxServices
	Children   SessionMuxChildren
	Middleware []middleware.Middleware
}

type SessionMuxOpts struct {
	PathPrefix string
}

type SessionMuxServices struct {
	SessionService session.SessionService
	MusicService   music.MusicService
	UserService    user.UserService
}

type SessionMuxChildren struct{}

func NewSessionMux(opts SessionMuxOpts, services SessionMuxServices, middleware []middleware.Middleware, children SessionMuxChildren) *SessionMux {
	mux := &SessionMux{
		http.NewServeMux(),
		opts,
		services,
		children,
		middleware,
	}

	mux.Handle("GET /", http.HandlerFunc(mux.handleSessionListPage))
	mux.Handle("POST /", http.HandlerFunc(mux.handleCreateSession))

	mux.Handle("GET /maker", http.HandlerFunc(mux.handleSessionMakerPage))

	mux.Handle("GET /{sessionId}", http.HandlerFunc(mux.handleSessionPage))

	mux.Handle("POST /{sessionId}/tracks", http.HandlerFunc(mux.handleCreateSessionTrack))

	mux.Handle("POST /{sessionId}/playlist", http.HandlerFunc(mux.handleCreateSessionPlaylist))
	mux.Handle("GET /{sessionId}/playlist", http.HandlerFunc(mux.handleGetSessionPlaylist))

	mux.Handle("GET /{sessionId}/phase-duration", http.HandlerFunc(mux.handleGetSessionPhaseDuration))

	mux.Handle("POST /{sessionId}/submission", http.HandlerFunc(mux.handleCreateSessionSubmission))

	mux.Handle("GET /{sessionId}/submission/{submissionId}", http.HandlerFunc(mux.handleGetSessionSubmission))
	mux.Handle("DELETE /{sessionId}/submission/{submissionId}", http.HandlerFunc(mux.handleDeleteSessionSubmission))

	mux.Handle("GET /{sessionId}/submission/{submissionId}/candidate", http.HandlerFunc(mux.handleGetSessionSubmissionCandidate))

	mux.Handle("GET /{sessionId}/result/{resultId}", http.HandlerFunc(mux.handleGetSessionResult))

	mux.Handle("POST /{sessionId}/vote", http.HandlerFunc(mux.handleCreateSessionVote))

	mux.Handle("DELETE /{sessionId}/vote/{voteId}", http.HandlerFunc(mux.handleDeleteSessionVote))

	return mux
}

func (mux *SessionMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	middleware.Apply(mux.ServeMux, mux.Middleware...).ServeHTTP(w, r)
}

func (mux *SessionMux) handleSessionListPage(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*user.User)

	sessions, err := mux.Services.SessionService.GetAll()
	if err != nil {
		http.Error(w, "Failed to get sessions", http.StatusInternalServerError)
		return
	}
	sessionValues := utils.Map(sessions, func(session *session.Session) session.Session { return *session })

	component := templates.UserSessions(*user, sessionValues)
	serverUtils.HandleHtmlResponse(r, w, component)
}

func (mux *SessionMux) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(serverUtils.UserCtxKey).(*user.User)

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.Form.Get("name")

	session, err := mux.Services.SessionService.Create(u, name)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	serverUtils.HandleRedirect(w, r, fmt.Sprintf("/app/session/%d", session.Id))
}

func (mux *SessionMux) handleSessionMakerPage(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*user.User)
	serverUtils.HandleHtmlResponse(r, w, templates.SessionMakerPage(*user))
}

func (mux *SessionMux) handleSessionPage(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*user.User)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session, err := mux.Services.SessionService.GetOne(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	acceptHeader := r.Header.Get("Accept")
	switch strings.ToLower(acceptHeader) {
	case "application/json":
		serverUtils.HandleJsonResponse(w, session)
	case "text/html":
	default:
		component := templates.SessionPage(*session, *user)
		serverUtils.HandleHtmlResponse(r, w, component)
	}
}

func (mux *SessionMux) handleCreateSessionTrack(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*user.User)
	err := mux.Services.MusicService.Authenticate(user)
	if err != nil {
		http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session, err := mux.Services.SessionService.GetOne(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	r.ParseForm()
	query := r.Form.Get("query")

	searchResults := make([]music.Track, 0)
	if query != "" {
		tracks, err := mux.Services.MusicService.SearchTracks(query)
		if err != nil {
			http.Error(w, "Failed to search Spotify", http.StatusInternalServerError)
			return
		}
		searchResults = utils.Map(tracks, func(track *music.Track) music.Track {
			return *track
		})
	}

	templates.SubmissionSearchBar(*session, *user, searchResults, "").Render(r.Context(), w)
}

func (mux *SessionMux) handleCreateSessionPlaylist(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*user.User)
	err := mux.Services.MusicService.Authenticate(user)
	if err != nil {
		http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session, err := mux.Services.SessionService.GetOne(sessionId)
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
	err = mux.Services.MusicService.CreatePlaylist(playlist, trackIds)
	if err != nil {
		http.Error(w, "Failed to create playlist", http.StatusInternalServerError)
		return
	}

	session, err = mux.Services.SessionService.AddPlaylist(sessionId, playlist.Id, user.Id)
	if err != nil {
		http.Error(w, "Failed to add playlist to session", http.StatusInternalServerError)
		return
	}

	serverUtils.HandleHtmlResponse(r, w, templates.PlaylistButton(*session, *playlist))
}

func (mux *SessionMux) handleGetSessionPlaylist(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*user.User)
	err := mux.Services.MusicService.Authenticate(user)
	if err != nil {
		http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	s, err := mux.Services.SessionService.GetOne(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	var playlist *music.Playlist
	sessionPlaylist, err := mux.Services.SessionService.GetPlaylist(sessionId, user.Id)
	if err == session.ErrPlaylistNotFound {
		playlist = &music.Playlist{}
	} else if err != nil {
		http.Error(w, "Failed to get playlist from session", http.StatusInternalServerError)
		return
	} else {
		playlist, err = mux.Services.MusicService.GetPlaylist(sessionPlaylist.Id)
		if err != nil {
			http.Error(w, "Failed to get playlist", http.StatusInternalServerError)
			return
		}
	}

	serverUtils.HandleHtmlResponse(r, w, templates.PlaylistButton(*s, *playlist))
}

func (mux *SessionMux) handleCreateSessionSubmission(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*user.User)

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	s, err := mux.Services.SessionService.GetOne(sessionId)
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
	s, err = mux.Services.SessionService.AddSubmission(s.Id, submission)
	if err != nil {
		http.Error(w, "Failed to add submission", http.StatusInternalServerError)
		return
	}

	templates.NewSubmission(*s, *user, *submission).Render(r.Context(), w)
}

func (mux *SessionMux) handleGetSessionPhaseDuration(w http.ResponseWriter, r *http.Request) {
	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session, err := mux.Services.SessionService.GetOne(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	templates.SessionPhaseDuration(*session).Render(r.Context(), w)
}

func (mux *SessionMux) handleGetSessionSubmission(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*user.User)
	err := mux.Services.MusicService.Authenticate(user)
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

	session, err := mux.Services.SessionService.GetOne(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	submission, err := mux.Services.SessionService.GetSubmission(sessionId, submissionId)
	if err != nil {
		http.Error(w, "Failed to get submission", http.StatusInternalServerError)
		return
	}

	track, err := mux.Services.MusicService.GetTrack(submission.TrackId)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
		return
	}

	templates.SubmissionItem(*session, *submission, *track).Render(r.Context(), w)
}

func (mux *SessionMux) handleDeleteSessionSubmission(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*user.User)

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

	session, err := mux.Services.SessionService.RemoveSubmission(sessionId, submissionId)
	if err != nil {
		http.Error(w, "Failed to delete submission", http.StatusInternalServerError)
		return
	}

	templates.DeleteSubmission(*session, *user).Render(r.Context(), w)
}

func (mux *SessionMux) handleGetSessionSubmissionCandidate(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*user.User)

	err := mux.Services.MusicService.Authenticate(user)
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

	session, err := mux.Services.SessionService.GetOne(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	submission, err := mux.Services.SessionService.GetSubmission(sessionId, submissionId)
	if err != nil {
		http.Error(w, "Failed to get submission", http.StatusInternalServerError)
		return
	}

	track, err := mux.Services.MusicService.GetTrack(submission.TrackId)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
		return
	}

	templates.VoteCandidate(*session, *user, *submission, *track).Render(r.Context(), w)
}

func (mux *SessionMux) handleCreateSessionVote(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*user.User)

	err := mux.Services.MusicService.Authenticate(user)
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
	submission, err := mux.Services.SessionService.GetSubmission(sessionId, submissionId)
	if err != nil {
		http.Error(w, "Failed to get submission", http.StatusInternalServerError)
		return
	}

	track, err := mux.Services.MusicService.GetTrack(submission.TrackId)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
		return
	}

	vote := &session.Vote{
		UserId:       user.Id,
		SubmissionId: submissionId,
	}
	s, err := mux.Services.SessionService.AddVote(sessionId, vote)
	if err != nil {
		http.Error(w, "Failed to add vote", http.StatusInternalServerError)
		return
	}

	templates.VoteCandidate(*s, *user, *submission, *track).Render(r.Context(), w)
}

func (mux *SessionMux) handleDeleteSessionVote(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*user.User)

	err := mux.Services.MusicService.Authenticate(user)
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

	vote, err := mux.Services.SessionService.GetVote(sessionId, voteId)
	if err != nil {
		http.Error(w, "Failed to get vote", http.StatusInternalServerError)
		return
	}

	submission, err := mux.Services.SessionService.GetSubmission(sessionId, vote.SubmissionId)
	if err != nil {
		http.Error(w, "Failed to get submission", http.StatusInternalServerError)
		return
	}

	track, err := mux.Services.MusicService.GetTrack(submission.TrackId)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
		return
	}

	session, err := mux.Services.SessionService.RemoveVote(sessionId, voteId)
	if err != nil {
		http.Error(w, "Failed to delete vote", http.StatusInternalServerError)
		return
	}

	templates.VoteCandidate(*session, *user, *submission, *track).Render(r.Context(), w)
}

func (mux *SessionMux) handleGetSessionResult(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(serverUtils.UserCtxKey).(*user.User)

	err := mux.Services.MusicService.Authenticate(user)
	if err != nil {
		http.Error(w, "Failed to authenticate user", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	resultId := r.PathValue("resultId")
	if resultId == "" {
		http.Error(w, "Invalid result ID", http.StatusBadRequest)
		return
	}

	session, err := mux.Services.SessionService.GetOne(sessionId)
	if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	result, err := mux.Services.SessionService.GetResult(sessionId, resultId)
	if err != nil {
		http.Error(w, "Failed to get result", http.StatusInternalServerError)
		return
	}

	submission, err := mux.Services.SessionService.GetSubmission(sessionId, result.SubmissionId)
	if err != nil {
		http.Error(w, "Failed to get submission", http.StatusInternalServerError)
		return
	}

	track, err := mux.Services.MusicService.GetTrack(submission.TrackId)
	if err != nil {
		http.Error(w, "Failed to get track", http.StatusInternalServerError)
		return
	}

	owner, err := mux.Services.UserService.Get(submission.UserId)
	if err != nil {
		http.Error(w, "Failed to get owner", http.StatusInternalServerError)
		return
	}

	templates.Result(*session, *result, *submission, *track, *owner).Render(r.Context(), w)
}
