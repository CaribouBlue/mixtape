package mux

import (
	"encoding/json"
	"net/http"

	"github.com/CaribouBlue/mixtape/internal/log/rlog"
	"github.com/CaribouBlue/mixtape/internal/server/middleware"
	"github.com/CaribouBlue/mixtape/internal/server/utils"
)

type ProfileMux struct {
	Mux[ProfileMuxOpts, ProfileMuxServices]
}

func (mux *ProfileMux) Opts() MuxOpts {
	return mux.opts.MuxOpts
}

type ProfileMuxOpts struct {
	MuxOpts
}

type ProfileMuxServices struct {
	MuxServices
}

func NewProfileMux(opts ProfileMuxOpts, services ProfileMuxServices, middleware []middleware.Middleware, children []ChildMux) *ProfileMux {
	mux := &ProfileMux{
		*NewMux(
			opts,
			services,
			children,
			middleware,
		),
	}

	mux.Handle("GET /", http.HandlerFunc(mux.handleProfilePage))

	return mux
}

func (mux *ProfileMux) handleProfilePage(w http.ResponseWriter, r *http.Request) {
	user, err := utils.ContextValue(r.Context(), utils.UserCtxKey)
	if err != nil {
		rlog.Logger(r).Error().Err(err).Msg("Failed to get user from context")
		http.Error(w, "Could not get user data", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
	}
}
