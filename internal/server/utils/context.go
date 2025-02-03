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
	RequestMetaDataCtxKey ContextKey = ContextKey{"request_meta_data"}
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
