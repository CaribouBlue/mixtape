package mux

import (
	"net/http"

	"github.com/CaribouBlue/top-spot/internal/server/middleware"
	"github.com/CaribouBlue/top-spot/internal/user"
)

type RootMux struct {
	*http.ServeMux
	Services   RootMuxServices
	Middleware []middleware.Middleware
	Children   RootMuxChildren
}

type RootMuxServices struct {
	UserService user.UserService
}

type RootMuxChildren struct {
	AuthMux *AuthMux
	AppMux  *AppMux
}

func NewRootMux(services RootMuxServices, middleware []middleware.Middleware, children RootMuxChildren) *RootMux {
	mux := &RootMux{
		http.NewServeMux(),
		services,
		middleware,
		children,
	}

	authPathPrefix := mux.Children.AuthMux.Opts.PathPrefix
	mux.Handle(authPathPrefix+"/", http.StripPrefix(authPathPrefix, mux.Children.AuthMux))

	appPathPrefix := mux.Children.AppMux.Opts.PathPrefix
	mux.Handle(appPathPrefix+"/", http.StripPrefix(appPathPrefix, mux.Children.AppMux))

	return mux
}

func (mux *RootMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	middleware.Apply(mux.ServeMux, mux.Middleware...).ServeHTTP(w, r)
}
