package mux

import (
	"net/http"

	"github.com/CaribouBlue/top-spot/internal/server/middleware"
)

type StaticMux struct {
	*http.ServeMux
	Opts       StaticMuxOpts
	Middleware []middleware.Middleware
}

type StaticMuxOpts struct {
	PathPrefix string
}

func NewStaticMux(opts StaticMuxOpts, middleware []middleware.Middleware) *StaticMux {
	mux := &StaticMux{
		ServeMux:   http.NewServeMux(),
		Opts:       opts,
		Middleware: middleware,
	}

	mux.Handle("/", http.FileServer(http.Dir("./static")))

	return mux
}

func (mux *StaticMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	middleware.Apply(mux.ServeMux, mux.Middleware...).ServeHTTP(w, r)
}
