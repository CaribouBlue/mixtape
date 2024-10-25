package server

import (
	"fmt"
	"net/http"

	"github.com/CaribouBlue/top-spot/db"
	"github.com/CaribouBlue/top-spot/spotify"
)

type RequestContextKey int

const (
	SpotifyClientRequestContextKey RequestContextKey = iota
	UserRequestContextKey
)

func getRequestContextValue[Value any](r *http.Request, key RequestContextKey) (Value, error) {
	value, ok := r.Context().Value(key).(Value)
	if !ok {
		return value, fmt.Errorf("value not found in context")
	}
	return value, nil
}

func getSpotifyClientFromRequestContext(r *http.Request) (*spotify.SpotifyClient, error) {
	return getRequestContextValue[*spotify.SpotifyClient](r, SpotifyClientRequestContextKey)
}

func getUserFromRequestContext(r *http.Request) (*db.UserDataModel, error) {
	return getRequestContextValue[*db.UserDataModel](r, UserRequestContextKey)
}
