package mux

import (
	"net/http"

	"github.com/CaribouBlue/mixtape/internal/server/middleware"
)

type Mux[Opts interface{}, Services interface{}] struct {
	*http.ServeMux
	opts              Opts
	Services          Services
	Children          []ChildMux
	Middleware        []middleware.Middleware
	BeforeEachRequest middleware.Middleware
}

type MuxOpts struct {
	PathPrefix string
}

type MuxServiceInitializer[Mux interface{}, Service interface{}] func(Mux, *http.Request) (Service, error)

type MuxServices struct {
}

type ChildMux interface {
	http.Handler
	Opts() MuxOpts
}

func NewMux[Opts interface{}, Services interface{}](opts Opts, services Services, children []ChildMux, mw []middleware.Middleware) *Mux[Opts, Services] {
	mux := &Mux[Opts, Services]{
		http.NewServeMux(),
		opts,
		services,
		children,
		mw,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})
		},
	}

	for _, child := range mux.Children {
		pathPrefix := child.Opts().PathPrefix
		mux.Handle(pathPrefix+"/", http.StripPrefix(pathPrefix, child))
	}

	return mux
}

func (mux *Mux[Opts, Services]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	middleware.Apply(mux.ServeMux, append(mux.Middleware, mux.BeforeEachRequest)...).ServeHTTP(w, r)
}
