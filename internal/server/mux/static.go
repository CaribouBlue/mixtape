package mux

import (
	"net/http"

	"github.com/CaribouBlue/mixtape/internal/config"
	"github.com/CaribouBlue/mixtape/internal/server/middleware"
)

type StaticMux struct {
	Mux[StaticMuxOpts, StaticMuxServices]
}

func (mux *StaticMux) Opts() MuxOpts {
	return mux.opts.MuxOpts
}

type StaticMuxOpts struct {
	MuxOpts
}

type StaticMuxServices struct {
	MuxServices
}

func NewStaticMux(opts StaticMuxOpts, services StaticMuxServices, middleware []middleware.Middleware, children []ChildMux) *StaticMux {
	mux := &StaticMux{
		*NewMux(
			opts,
			services,
			children,
			middleware,
		),
	}

	appDataPath := config.GetConfigValue(config.ConfAppDataPath)
	if appDataPath == "" {
		appDataPath = "."
	}

	mux.Handle("/", http.FileServer(http.Dir(appDataPath+"/static")))

	return mux
}
