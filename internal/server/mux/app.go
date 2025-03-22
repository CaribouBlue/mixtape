package mux

import (
	"net/http"

	"github.com/CaribouBlue/mixtape/internal/core"
	"github.com/CaribouBlue/mixtape/internal/server/middleware"
	"github.com/CaribouBlue/mixtape/internal/server/utils"
	"github.com/CaribouBlue/mixtape/internal/templates"
)

type AppMux struct {
	Mux[AppMuxOpts, AppMuxServices]
}

func (mux *AppMux) Opts() MuxOpts {
	return mux.opts.MuxOpts
}

type AppMuxOpts struct {
	MuxOpts
}

type AppMuxServices struct {
	MuxServices
}

func NewAppMux(opts AppMuxOpts, services AppMuxServices, middleware []middleware.Middleware, children []ChildMux) *AppMux {
	mux := &AppMux{
		*NewMux(
			opts,
			services,
			children,
			middleware,
		),
	}
	mux.Middleware = middleware

	mux.Handle("/home", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(utils.UserCtxKey).(*core.UserEntity)

		utils.HandleHtmlResponse(r, w, templates.Home(*user))
	}))

	return mux
}
