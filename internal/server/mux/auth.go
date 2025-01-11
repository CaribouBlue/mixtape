package mux

import (
	"net/http"

	"github.com/CaribouBlue/top-spot/internal/music"
	"github.com/CaribouBlue/top-spot/internal/music/spotify"
	"github.com/CaribouBlue/top-spot/internal/server/middleware"
	"github.com/CaribouBlue/top-spot/internal/user"
)

type AuthMux struct {
	*http.ServeMux
	userService  user.UserService
	musicService music.MusicService
}

func NewAuthMux(userService user.UserService, musicService music.MusicService) *AuthMux {
	mux := &AuthMux{
		ServeMux:     http.NewServeMux(),
		userService:  userService,
		musicService: musicService,
	}
	mux.RegisterHandlers()
	return mux
}

func (mux *AuthMux) RegisterHandlers() {
	mux.Handle("/user", http.HandlerFunc(mux.handleUserLogin))
	mux.Handle("/spotify", http.HandlerFunc(mux.handleSpotifyAuth))
	mux.Handle("/spotify/redirect", http.HandlerFunc(mux.handleSpotifyAuthRedirect))
}

func (mux *AuthMux) handleUserLogin(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(middleware.UserCtxKey).(*user.User)
	if u.Id == 0 {
		http.Error(w, "User does not exist", http.StatusInternalServerError)
		return
	}

	err := mux.musicService.Authenticate(u)
	if err != nil {
		http.Redirect(w, r, "/auth/spotify", http.StatusFound)
		return
	} else {
		http.Redirect(w, r, appPathPrefix, http.StatusFound)
		return
	}
}

func (mux *AuthMux) handleSpotifyAuth(w http.ResponseWriter, r *http.Request) {
	spotify := spotify.DefaultClient()
	userAuthUrl, err := spotify.GetUserAuthUrl()
	if err != nil {
		http.Error(w, "Failed to get user auth url", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, userAuthUrl, http.StatusFound)
}

func (mux *AuthMux) handleSpotifyAuthRedirect(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(middleware.UserCtxKey).(*user.User)
	u, err := mux.userService.Get(u.Id)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	spotify := spotify.DefaultClient()

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found in request", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, "State not found in request", http.StatusBadRequest)
		return
	}

	err = spotify.GetNewAccessToken(code)
	if err != nil {
		http.Error(w, "Failed to get new access token", http.StatusBadRequest)
		return
	}

	u.SpotifyAccessToken, err = spotify.GetValidAccessToken()
	if err != nil {
		http.Error(w, "Failed to get valid access token", http.StatusInternalServerError)
		return
	}
	mux.userService.Update(u)

	http.Redirect(w, r, authPathPrefix+"/user", http.StatusFound)
}
