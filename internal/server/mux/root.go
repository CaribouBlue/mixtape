package mux

import (
	"net/http"

	"github.com/CaribouBlue/mixtape/internal/core"
	"github.com/CaribouBlue/mixtape/internal/server/middleware"
	"github.com/CaribouBlue/mixtape/internal/server/response"
)

type RootMux struct {
	Mux[RootMuxOpts, RootMuxServices]
}

func (mux *RootMux) Opts() MuxOpts {
	return mux.opts.MuxOpts
}

type RootMuxOpts struct {
	MuxOpts
}

type RootMuxServices struct {
	MuxServices
	UserService *core.UserService
}

func NewRootMux(opts RootMuxOpts, services RootMuxServices, middleware []middleware.Middleware, children []ChildMux) *RootMux {
	mux := &RootMux{
		*NewMux(
			opts,
			services,
			children,
			middleware,
		),
	}

	mux.Handle("/{$}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response.HandleRedirect(w, r, "/app")
	}))

	return mux
}
