package mux

import (
	"net/http"

	"github.com/CaribouBlue/top-spot/internal/server/middleware"
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
}

func NewRootMux() *RootMux {
	mux := &RootMux{http.NewServeMux()}
	mux.RegisterHandlers()
	return mux
}

func (mux *RootMux) RegisterHandlers() {
	withUser := middleware.WithUser(middleware.WithUserOpts{DefaultUserId: defaultUserId})
	withEnforcedAuthentication := middleware.WithEnforcedAuthentication(middleware.WithEnforcedAuthenticationOpts{UnauthenticatedRedirectPath: authPathPrefix + "/user"})

	authMuxHandler := middleware.Apply(NewAuthMux(), withUser)
	mux.Handle(authPathPrefix+"/", http.StripPrefix(authPathPrefix, authMuxHandler))

	appMuxHandler := middleware.Apply(NewAppMux(), withUser, withEnforcedAuthentication)
	mux.Handle(appPathPrefix+"/", http.StripPrefix(appPathPrefix, appMuxHandler))
}
