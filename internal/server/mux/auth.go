package mux

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/CaribouBlue/top-spot/internal/model"
	"github.com/CaribouBlue/top-spot/internal/server/middleware"
)

func handleUserLogin(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	err := user.Read()
	if err == sql.ErrNoRows {
		user.Create()
	} else if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
	}

	isAuthenticated, err := user.IsAuthenticated()
	if err != nil {
		http.Error(w, "Failed to check authentication", http.StatusInternalServerError)
		return
	}

	if !isAuthenticated {
		http.Redirect(w, r, "/auth/spotify", http.StatusFound)
		return
	} else {
		http.Redirect(w, r, appPathPrefix, http.StatusFound)
	}
}

func handleSpotifyAuth(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	spotify := authorizedSpotifyClient(user)

	userAuthUrl, err := spotify.GetUserAuthUrl()
	if err != nil {
		log.Fatal(err)
	}

	http.Redirect(w, r, userAuthUrl, http.StatusFound)
}

func handleSpotifyAuthRedirect(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	spotify := authorizedSpotifyClient(user)

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

	err := spotify.GetNewAccessToken(code)
	if err != nil {
		http.Error(w, "Failed to get new access token", http.StatusBadRequest)
		return
	}

	err = user.Read()
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
	}

	user.Data.SpotifyAccessToken, err = spotify.GetValidAccessToken()
	if err != nil {
		http.Error(w, "Failed to get valid access token", http.StatusInternalServerError)
		return
	}
	user.Update()

	http.Redirect(w, r, authPathPrefix+"/user", http.StatusFound)
}

type AuthMux struct {
	*http.ServeMux
}

func NewAuthMux() *AuthMux {
	mux := &AuthMux{http.NewServeMux()}
	mux.RegisterHandlers()
	return mux
}

func (mux *AuthMux) RegisterHandlers() {
	mux.Handle("/user", http.HandlerFunc(handleUserLogin))
	mux.Handle("/spotify", http.HandlerFunc(handleSpotifyAuth))
	mux.Handle("/spotify/redirect", http.HandlerFunc(handleSpotifyAuthRedirect))
}
