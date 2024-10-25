package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/CaribouBlue/top-spot/templates"
)

const (
	appMuxPathPrefix string = "/app"
)

func appLandingHandler(w http.ResponseWriter, r *http.Request) {
	templates.Landing().Render(r.Context(), w)
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
	appMux.Handle("GET /profile", handlerFuncWithMiddleware(appProfileHandler, withSpotify))

	appMuxWithMiddleware := applyMiddleware(appMux, enforceAuthentication)

	parentMux.Handle(appMuxPathPrefix+"/", http.StripPrefix(appMuxPathPrefix, appMuxWithMiddleware))
}
