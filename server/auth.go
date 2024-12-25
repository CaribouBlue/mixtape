package server

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/CaribouBlue/top-spot/model"
)

const (
	authMuxPathPrefix       = "/auth"
	userId            int64 = 6666
)

func authLoginHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserCtxKey).(*model.UserModel)
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
		http.Redirect(w, r, appMuxPathPrefix, http.StatusFound)
	}
}

func authSpotifyHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserCtxKey).(*model.UserModel)
	spotify := authorizedSpotifyClient(user)

	userAuthUrl, err := spotify.GetUserAuthUrl()
	if err != nil {
		log.Fatal(err)
	}

	http.Redirect(w, r, userAuthUrl, http.StatusFound)
}

func authSpotifyRedirectHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserCtxKey).(*model.UserModel)
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

	http.Redirect(w, r, authMuxPathPrefix+"/user", http.StatusFound)
}

func registerAuthMux(parentMux *http.ServeMux) {
	authMux := http.NewServeMux()
	authMux.Handle("/user", handlerFuncWithMiddleware(authLoginHandler))
	authMux.Handle("/spotify", handlerFuncWithMiddleware(authSpotifyHandler))
	authMux.Handle("/spotify/redirect", handlerFuncWithMiddleware(authSpotifyRedirectHandler))

	authMuxWithMiddleware := applyMiddleware(authMux, withUser)

	parentMux.Handle(authMuxPathPrefix+"/", http.StripPrefix(authMuxPathPrefix, authMuxWithMiddleware))
}
