package utils

import (
	"context"
	"errors"
	"net/http"

	"github.com/CaribouBlue/mixtape/internal/core"
	"github.com/CaribouBlue/mixtape/internal/spotify"
	"github.com/google/uuid"
)

type ContextKey[Value interface{}] struct {
	name string
}

func (c ContextKey[Value]) String() string {
	return c.name
}

var (
	UserCtxKey            = ContextKey[*core.UserEntity]{"user"}
	SpotifyClientCtxKey   = ContextKey[*spotify.Client]{"spotify_client"}
	RequestMetaDataCtxKey = ContextKey[*RequestMetadata]{"request_meta_data"}
)

func ContextValue[T interface{}](ctx context.Context, key ContextKey[T]) (T, error) {
	var val T
	v, ok := ctx.Value(key).(T)
	if !ok {
		return val, errors.New("invalid value type in context")
	}
	return v, nil
}

func SetContextValue[T interface{}](ctx context.Context, key ContextKey[T], value T) context.Context {
	return context.WithValue(ctx, key, value)
}

type RequestMetadata struct {
	RequestId     string
	IsHtmxRequest bool
}

func NewRequestMetadata(r *http.Request) *RequestMetadata {
	return &RequestMetadata{
		RequestId:     uuid.New().String(),
		IsHtmxRequest: r.Header.Get("HX-Request") != "",
	}
}
