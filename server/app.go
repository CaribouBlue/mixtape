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
	if err != nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session := db.NewGameSessionDataModel()
	session.SetId(sessionId)
	err = session.GetById()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	templates.Session(templateModel).Render(r.Context(), w)
}

func appSessionTracksSearchHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequestContext(r)
	if err != nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	spotifyClient, err := getSpotifyClientFromRequestContext(r)
	if err != nil {
		http.Error(w, "Spotify not found in context", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session := db.NewGameSessionDataModel()
	session.SetId(sessionId)
	err = session.GetById()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	r.ParseForm()
	query := r.Form.Get("query")

	templateModel := templates.NewSessionTemplateModel(*session, *user)

	if query != "" {
		searchResults, err := spotifyClient.SearchTracks(query)
		if err != nil {
			http.Error(w, "Failed to search Spotify", http.StatusInternalServerError)
			return
		}

		templateModel.SearchResult = *searchResults
	}

	templates.SubmissionSearchResults(templateModel).Render(r.Context(), w)
}

func appSessionSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequestContext(r)
	if err != nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	session := db.NewGameSessionDataModel()
	session.SetId(sessionId)
	err = session.GetById()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	r.ParseForm()
	trackId := r.Form.Get("trackId")
	submission := db.NewSubmissionDataModel(user.GetId(), trackId)

	err = session.AddSubmission(*submission)
	if err != nil {
		if err == db.ErrSubmissionsMaxedOut {
			http.Error(w, "Max submissions reached", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to add submission", http.StatusInternalServerError)
		return
	}

	err = session.Update()
	if err != nil {
		http.Error(w, "Failed to add submission", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	templates.NewSubmission(templateModel, submission.Id).Render(r.Context(), w)
}

func appSessionDeleteSubmissionHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequestContext(r)
	if err != nil {
		http.Error(w, "User not found in context", http.StatusInternalServerError)
		return
	}

	sessionId, err := strconv.ParseInt(r.PathValue("sessionId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	submissionId := r.PathValue("submissionId")
	if err != nil {
		http.Error(w, "Invalid submission ID", http.StatusBadRequest)
		return
	}

	session := db.NewGameSessionDataModel()
	session.SetId(sessionId)
	err = session.GetById()
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get session", http.StatusInternalServerError)
		return
	}

	err = session.DeleteSubmission(submissionId, user.GetId())
	if err != nil {
		http.Error(w, "Failed to delete submission", http.StatusInternalServerError)
		return
	}

	err = session.Update()
	if err != nil {
		http.Error(w, "Failed to delete submission", http.StatusInternalServerError)
		return
	}

	templateModel := templates.NewSessionTemplateModel(*session, *user)
	templates.DeleteSubmission(templateModel).Render(r.Context(), w)
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
	appMux.Handle("GET /session/{sessionId}", http.HandlerFunc(appSessionHandler))
	appMux.Handle("POST /session/{sessionId}/tracks", handlerFuncWithMiddleware(appSessionTracksSearchHandler, withSpotify))
	appMux.Handle("POST /session/{sessionId}/submission", handlerFuncWithMiddleware(appSessionSubmissionHandler))
	appMux.Handle("DELETE /session/{sessionId}/submission/{submissionId}", handlerFuncWithMiddleware(appSessionDeleteSubmissionHandler))

	appMux.Handle("GET /profile", handlerFuncWithMiddleware(appProfileHandler, withSpotify))

	appMuxWithMiddleware := applyMiddleware(appMux, enforceAuthentication)

	parentMux.Handle(appMuxPathPrefix+"/", http.StripPrefix(appMuxPathPrefix, appMuxWithMiddleware))
}
