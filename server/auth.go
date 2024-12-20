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
	db, err := getDbFromRequestContext(r)
	if err != nil {
		http.Error(w, "Database not found in context", http.StatusInternalServerError)
		return
	}

	user := model.NewUserModel(db, model.WithId(userId))
	err = user.Read()
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
	spotify, err := getSpotifyClientFromRequestContext(r)
	if err != nil {
		http.Error(w, "Spotify client not found in context", http.StatusInternalServerError)
		return
	}

	userAuthUrl, err := spotify.GetUserAuthUrl()
	if err != nil {
		log.Fatal(err)
	}

	http.Redirect(w, r, userAuthUrl, http.StatusFound)
}

func authSpotifyRedirectHandler(w http.ResponseWriter, r *http.Request) {
	spotify, err := getSpotifyClientFromRequestContext(r)
	if err != nil {
		http.Error(w, "Spotify client not found in context", http.StatusInternalServerError)
		return
	}

	db, err := getDbFromRequestContext(r)
	if err != nil {
		http.Error(w, "Database not found in context", http.StatusInternalServerError)
		return
	}

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

	user := model.NewUserModel(db, model.WithId(userId))
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
	authMux.Handle("/user", http.HandlerFunc(authLoginHandler))
	authMux.Handle("/spotify", handlerFuncWithMiddleware(authSpotifyHandler, withUser, withSpotify))
	authMux.Handle("/spotify/redirect", handlerFuncWithMiddleware(authSpotifyRedirectHandler, withUser, withSpotify))

	authMuxWithMiddleware := applyMiddleware(authMux)

	parentMux.Handle(authMuxPathPrefix+"/", http.StripPrefix(authMuxPathPrefix, authMuxWithMiddleware))
}
