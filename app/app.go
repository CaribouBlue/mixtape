package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"

	"github.com/CaribouBlue/top-spot/db"
	"github.com/CaribouBlue/top-spot/spotify"
)

type RequestContextKey int

const (
	SpotifyClientRequestContextKey RequestContextKey = iota
	UserRequestContextKey
)

const userId int64 = 666

func authLoginHandler(w http.ResponseWriter, r *http.Request) {
	user := db.NewUserDataModel()
	user.SetId(userId)

	err := user.GetById()
	if err == sql.ErrNoRows {
		user.Insert()
	} else if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
	}

	if user.SpotifyAccessToken.AccessToken == "" {
		http.Redirect(w, r, "/auth/spotify", http.StatusFound)
		return
	} else {
		http.Redirect(w, r, "/app", http.StatusFound)
	}
}

func authSpotifyHandler(w http.ResponseWriter, r *http.Request) {
	spotify, ok := r.Context().Value(SpotifyClientRequestContextKey).(*spotify.SpotifyClient)
	if !ok {
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
	spotify, ok := r.Context().Value(SpotifyClientRequestContextKey).(*spotify.SpotifyClient)
	if !ok {
		http.Error(w, "Spotify client not found in context", http.StatusInternalServerError)
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

	err := spotify.GetNewAccessToken(code)
	if err != nil {
		http.Error(w, "Failed to get new access token", http.StatusBadRequest)
		return
	}

	user := db.NewUserDataModel()
	user.SetId(userId)

	err = user.GetById()
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
	}

	user.SpotifyAccessToken, err = spotify.GetValidAccessToken()
	if err != nil {
		http.Error(w, "Failed to get valid access token", http.StatusInternalServerError)
		return
	}
	user.Update()

	http.Redirect(w, r, "/auth/user", http.StatusFound)
}

func withSpotify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := os.Getenv("SPOTIFY_CLIENT_ID")
		clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
		redirectUri := os.Getenv("SPOTIFY_REDIRECT_URI")
		scope := os.Getenv("SPOTIFY_SCOPE")

		spotifyClient := spotify.NewSpotifyClient(clientID, clientSecret, redirectUri, scope)

		user, ok := r.Context().Value(UserRequestContextKey).(*db.UserDataModel)
		if ok {
			spotifyClient.SetAccessToken(user.SpotifyAccessToken)
		}

		ctx := context.WithValue(r.Context(), SpotifyClientRequestContextKey, spotifyClient)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := db.NewUserDataModel()
		user.SetId(userId)

		isAuthenticated, err := user.IsAuthenticated()
		if err != nil {
			http.Error(w, "Failed to check authentication", http.StatusInternalServerError)
			return
		}

		if !isAuthenticated {
			http.Redirect(w, r, "/auth/user", http.StatusFound)
			return
		}

		ctx := context.WithValue(r.Context(), UserRequestContextKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func handleWithMiddleware(pattern string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) {
	slices.Reverse(middlewares)
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	http.Handle(pattern, handler)
}

func StartServer() {
	http.Handle("GET /auth/user", http.HandlerFunc(authLoginHandler))
	http.Handle("GET /auth/spotify", withSpotify(http.HandlerFunc(authSpotifyHandler)))
	http.Handle("GET /auth/spotify/redirect", withSpotify(http.HandlerFunc(authSpotifyRedirectHandler)))

	handleWithMiddleware("GET /app", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spotify, ok := r.Context().Value(SpotifyClientRequestContextKey).(*spotify.SpotifyClient)
		if !ok {
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
	}), withUser, withSpotify)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	serverAddr := fmt.Sprintf("localhost:%s", port)
	fmt.Println("Starting server at ", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
