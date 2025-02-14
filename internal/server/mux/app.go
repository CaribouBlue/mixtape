package mux

import (
	"net/http"

	"github.com/CaribouBlue/top-spot/internal/core"
	"github.com/CaribouBlue/top-spot/internal/server/middleware"
	"github.com/CaribouBlue/top-spot/internal/server/utils"
	"github.com/CaribouBlue/top-spot/internal/templates"
)

type AppMux struct {
	*http.ServeMux
	Opts       AppMuxOpts
	Children   AppMuxChildren
	Middleware []middleware.Middleware
}

type AppMuxOpts struct {
	PathPrefix string
}

type AppMuxChildren struct {
	SessionMux *SessionMux
	ProfileMux *ProfileMux
}

func NewAppMux(opts AppMuxOpts, mw []middleware.Middleware, children AppMuxChildren) *AppMux {
	mux := &AppMux{
		http.NewServeMux(),
		opts,
		children,
		mw,
	}

	mux.Handle("/home", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(utils.UserCtxKey).(*core.UserEntity)

		utils.HandleHtmlResponse(r, w, templates.Home(*user))
	}))

	sessionPathPrefix := mux.Children.SessionMux.Opts.PathPrefix
	mux.Handle(sessionPathPrefix+"/", http.StripPrefix(sessionPathPrefix, mux.Children.SessionMux))

	profilePathPrefix := mux.Children.ProfileMux.Opts.PathPrefix
	mux.Handle(profilePathPrefix+"/", http.StripPrefix(profilePathPrefix, mux.Children.ProfileMux))

	return mux
}

func (mux *AppMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	middleware.Apply(mux.ServeMux, mux.Middleware...).ServeHTTP(w, r)
}
