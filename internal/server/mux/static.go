package mux

import (
	"net/http"

	"github.com/CaribouBlue/top-spot/internal/server/middleware"
)

type StaticMux struct {
	*http.ServeMux
	Opts       StaticMuxOpts
	Services   StaticMuxServices
	Children   StaticMuxChildren
	Middleware []middleware.Middleware
}

type StaticMuxOpts struct {
	PathPrefix string
}

type StaticMuxServices struct {
}

type StaticMuxChildren struct {
}

func NewStaticMux(opts StaticMuxOpts, services StaticMuxServices, middleware []middleware.Middleware, children StaticMuxChildren) *StaticMux {
	mux := &StaticMux{
		ServeMux:   http.NewServeMux(),
		Opts:       opts,
		Services:   services,
		Children:   children,
		Middleware: middleware,
	}

	mux.Handle("/", http.FileServer(http.Dir("./static")))

	return mux
}

func (mux *StaticMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	middleware.Apply(mux.ServeMux, mux.Middleware...).ServeHTTP(w, r)
}
