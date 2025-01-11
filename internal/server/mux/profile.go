package mux

import (
	"encoding/json"
	"net/http"

	"github.com/CaribouBlue/top-spot/internal/server/middleware"
	"github.com/CaribouBlue/top-spot/internal/user"
)

type ProfileMux struct {
	*http.ServeMux
	userService user.UserService
}

func NewProfileMux(userService user.UserService) *ProfileMux {
	mux := &ProfileMux{
		http.NewServeMux(),
		userService,
	}
	mux.RegisterHandlers()
	return mux
}

func (mux *ProfileMux) RegisterHandlers() {
	mux.Handle("GET /", http.HandlerFunc(mux.handleProfilePage))
}

func (mux *ProfileMux) handleProfilePage(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*user.User)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
	}
}
