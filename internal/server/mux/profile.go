package mux

import (
	"encoding/json"
	"net/http"

	"github.com/CaribouBlue/top-spot/internal/entities/user"
	"github.com/CaribouBlue/top-spot/internal/server/middleware"
	"github.com/CaribouBlue/top-spot/internal/server/utils"
)

type ProfileMux struct {
	*http.ServeMux
	Opts       ProfileMuxOpts
	Services   ProfileMuxServices
	Children   ProfileMuxChildren
	Middleware []middleware.Middleware
}

type ProfileMuxOpts struct {
	PathPrefix string
}

type ProfileMuxServices struct{}

type ProfileMuxChildren struct{}

func NewProfileMux(opts ProfileMuxOpts, services ProfileMuxServices, middleware []middleware.Middleware, children ProfileMuxChildren) *ProfileMux {
	mux := &ProfileMux{
		http.NewServeMux(),
		opts,
		services,
		children,
		middleware,
	}

	mux.Handle("GET /", http.HandlerFunc(mux.handleProfilePage))

	return mux
}

func (mux *ProfileMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	middleware.Apply(mux.ServeMux, mux.Middleware...).ServeHTTP(w, r)
}

func (mux *ProfileMux) handleProfilePage(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(utils.UserCtxKey).(*user.User)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
	}
}
