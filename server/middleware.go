package server

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"slices"

	"github.com/CaribouBlue/top-spot/db"
	"github.com/CaribouBlue/top-spot/spotify"
)

type middleware func(http.Handler) http.Handler

func applyMiddleware(handler http.Handler, middlewares ...middleware) http.Handler {
	slices.Reverse(middlewares)
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

func handlerFuncWithMiddleware(handler http.HandlerFunc, middlewares ...middleware) http.Handler {
	return applyMiddleware(http.HandlerFunc(handler), middlewares...)
}

type WrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *WrappedWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func withRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrappedWriter := &WrappedWriter{w, http.StatusOK}

		path := r.URL.Path
		method := r.Method

		next.ServeHTTP(wrappedWriter, r)

		log.Println(wrappedWriter.statusCode, method, path)
	})
}

func withSpotify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := getSpotifyClientFromRequestContext(r)
		if err == nil {
			next.ServeHTTP(w, r)
			return
		}

		clientID := os.Getenv("SPOTIFY_CLIENT_ID")
		clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
		redirectUri := os.Getenv("SPOTIFY_REDIRECT_URI")
		scope := os.Getenv("SPOTIFY_SCOPE")

		spotifyClient := spotify.NewSpotifyClient(clientID, clientSecret, redirectUri, scope)

		user, err := getUserFromRequestContext(r)
		if err == nil {
			spotifyClient.SetAccessToken(user.SpotifyAccessToken)
		}

		ctx := context.WithValue(r.Context(), SpotifyClientRequestContextKey, spotifyClient)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := getUserFromRequestContext(r)
		if err == nil {
			next.ServeHTTP(w, r)
			return
		}

		user := db.NewUserDataModel()
		user.SetId(userId)
		err = user.GetById()
		if err == sql.ErrNoRows {
			// if user does not exist, continue with empty user data model
		} else if err != nil {
			http.Error(w, "Failed to get user", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), UserRequestContextKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func enforceAuthentication(next http.Handler) http.Handler {
	enforceAuthenticationHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := getUserFromRequestContext(r)
		if err != nil {
			http.Error(w, "User not found in context", http.StatusInternalServerError)
			return
		}

		isAuthenticated, err := user.IsAuthenticated()
		if err != nil {
			http.Error(w, "Failed to check authentication", http.StatusInternalServerError)
			return
		}

		if !isAuthenticated {
			http.Redirect(w, r, authMuxPathPrefix+"/user", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
	return applyMiddleware(enforceAuthenticationHandler, withUser)
}
