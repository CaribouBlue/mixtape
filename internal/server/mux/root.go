package mux

import (
	"net/http"

	"github.com/CaribouBlue/top-spot/internal/music"
	"github.com/CaribouBlue/top-spot/internal/server/middleware"
	"github.com/CaribouBlue/top-spot/internal/session"
	"github.com/CaribouBlue/top-spot/internal/user"
)

const (
	authPathPrefix    string = "/auth"
	appPathPrefix     string = "/app"
	sessionPathPrefix string = "/session"
	profilePathPrefix string = "/profile"
	defaultUserId     int64  = 6666
)

type RootMux struct {
	*http.ServeMux
	userService    user.UserService
	musicService   music.MusicService
	sessionService session.SessionService
}

func NewRootMux(userService user.UserService, musicService music.MusicService, sessionService session.SessionService) *RootMux {
	mux := &RootMux{
		http.NewServeMux(),
		userService,
		musicService,
		sessionService,
	}
	mux.RegisterHandlers()
	return mux
}

func (mux *RootMux) RegisterHandlers() {
	withUser := middleware.WithUser(middleware.WithUserOpts{DefaultUserId: defaultUserId, UserService: mux.userService})
	withEnforcedAuthentication := middleware.WithEnforcedAuthentication(middleware.WithEnforcedAuthenticationOpts{
		UnauthenticatedRedirectPath: authPathPrefix + "/user",
		UserService:                 mux.userService,
	})

	authMuxHandler := middleware.Apply(NewAuthMux(mux.userService, mux.musicService), withUser)
	mux.Handle(authPathPrefix+"/", http.StripPrefix(authPathPrefix, authMuxHandler))

	appMuxHandler := middleware.Apply(NewAppMux(mux.userService, mux.musicService, mux.sessionService), withUser, withEnforcedAuthentication)
	mux.Handle(appPathPrefix+"/", http.StripPrefix(appPathPrefix, appMuxHandler))
}
