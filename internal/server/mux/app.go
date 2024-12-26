package mux

import (
	"net/http"
)

type AppMux struct {
	*http.ServeMux
}

func NewAppMux() *AppMux {
	mux := &AppMux{http.NewServeMux()}
	mux.RegisterHandlers()
	return mux
}

func (mux *AppMux) RegisterHandlers() {
	sessionMuxHandler := NewSessionMux()
	mux.Handle(sessionPathPrefix+"/", http.StripPrefix(sessionPathPrefix, sessionMuxHandler))

	profileMuxHandler := NewProfileMux()
	mux.Handle(profilePathPrefix+"/", http.StripPrefix(profilePathPrefix, profileMuxHandler))
}
