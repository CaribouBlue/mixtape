package mux

import (
	"net/http"

	"github.com/CaribouBlue/top-spot/internal/music"
	"github.com/CaribouBlue/top-spot/internal/session"
	"github.com/CaribouBlue/top-spot/internal/user"
)

type AppMux struct {
	*http.ServeMux
	userService    user.UserService
	musicService   music.MusicService
	sessionService session.SessionService
}

func NewAppMux(userService user.UserService, musicService music.MusicService, sessionService session.SessionService) *AppMux {
	mux := &AppMux{
		http.NewServeMux(),
		userService,
		musicService,
		sessionService,
	}
	mux.RegisterHandlers()
	return mux
}

func (mux *AppMux) RegisterHandlers() {
	sessionMuxHandler := NewSessionMux(mux.sessionService, mux.musicService)
	mux.Handle(sessionPathPrefix+"/", http.StripPrefix(sessionPathPrefix, sessionMuxHandler))

	profileMuxHandler := NewProfileMux(mux.userService)
	mux.Handle(profilePathPrefix+"/", http.StripPrefix(profilePathPrefix, profileMuxHandler))
}
