package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/CaribouBlue/top-spot/db"
	"github.com/CaribouBlue/top-spot/templates"
)

const (
	appMuxPathPrefix string = "/app"
)

func appLandingHandler(w http.ResponseWriter, r *http.Request) {
	templates.Landing().Render(r.Context(), w)
}

func createAppSessionHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var session *db.GameSessionDataModel = db.NewGameSessionDataModel()
	if err := json.NewDecoder(r.Body).Decode(session); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	err := session.GetById()
	if err == sql.ErrNoRows {
		err = session.Insert()
	} else if err == nil {
		http.Error(w, "Session already exists", http.StatusConflict)
		return
	}

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

func appSessionHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequestContext(r)

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session := db.NewGameSessionDataModel()
	session.SetId(id)
	err = session.GetById()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	props := templates.SessionProps{Session: *session, User: *user}
	templates.Session(props).Render(r.Context(), w)
}

func appSpotifySearchHandler(w http.ResponseWriter, r *http.Request) {
	spotifyClient, err := getSpotifyClientFromRequestContext(r)
	if err != nil {
		http.Error(w, "Spotify not found in context", http.StatusInternalServerError)
		return
	}

	r.ParseForm()
	query := r.Form.Get("query")

	props := templates.SpotifySearchResultsProps{}

	if query != "" {
		searchResults, err := spotifyClient.SearchTracks(query)
		if err != nil {
			http.Error(w, "Failed to search Spotify", http.StatusInternalServerError)
			return
		}

		props.Results = *searchResults
	}

	templates.SpotifySearchResults(props).Render(r.Context(), w)
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

	appMux.Handle("POST /session", http.HandlerFunc(createAppSessionHandler))
	appMux.Handle("GET /session/{id}", http.HandlerFunc(appSessionHandler))

	appMux.Handle("POST /spotify-search", handlerFuncWithMiddleware(appSpotifySearchHandler, withSpotify))

	appMux.Handle("GET /profile", handlerFuncWithMiddleware(appProfileHandler, withSpotify))

	appMuxWithMiddleware := applyMiddleware(appMux, enforceAuthentication)

	parentMux.Handle(appMuxPathPrefix+"/", http.StripPrefix(appMuxPathPrefix, appMuxWithMiddleware))
}
