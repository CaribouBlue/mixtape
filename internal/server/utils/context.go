package utils

import (
	"net/http"

	"github.com/google/uuid"
)

type ContextKey struct {
	name string
}

func (c ContextKey) String() string {
	return c.name
}

var (
	UserCtxKey            ContextKey = ContextKey{"user"}
	SpotifyClientCtxKey   ContextKey = ContextKey{"spotify_client"}
	RequestMetaDataCtxKey ContextKey = ContextKey{"request_meta_data"}
	MusicServiceCtxKey    ContextKey = ContextKey{"music_service"}
)

type RequestMetadata struct {
	RequestId     string
	IsHtmxRequest bool
}

func NewRequestMetadata(r *http.Request) RequestMetadata {
	return RequestMetadata{
		RequestId:     uuid.New().String(),
		IsHtmxRequest: r.Header.Get("HX-Request") != "",
	}
}
