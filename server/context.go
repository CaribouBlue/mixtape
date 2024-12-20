package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/CaribouBlue/top-spot/db"
	"github.com/CaribouBlue/top-spot/model"
	"github.com/CaribouBlue/top-spot/spotify"
)

type ContextKey int

const (
	SpotifyClientContextKey ContextKey = iota
	UserContextKey
	DbContextKey
)

func NewDefaultContext(db db.Db) context.Context {
	cxt := context.Background()
	cxt = context.WithValue(cxt, DbContextKey, db)
	return cxt
}

func getRequestContextValue[Value any](r *http.Request, key ContextKey) (Value, error) {
	value, ok := r.Context().Value(key).(Value)
	if !ok {
		return value, fmt.Errorf("value not found in context")
	}
	return value, nil
}

func getSpotifyClientFromRequestContext(r *http.Request) (*spotify.SpotifyClient, error) {
	return getRequestContextValue[*spotify.SpotifyClient](r, SpotifyClientContextKey)
}

func getUserFromRequestContext(r *http.Request) (*model.UserModel, error) {
	return getRequestContextValue[*model.UserModel](r, UserContextKey)
}

func getDbFromRequestContext(r *http.Request) (db.Db, error) {
	return getRequestContextValue[db.Db](r, DbContextKey)
}
