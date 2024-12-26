package mux

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/CaribouBlue/top-spot/internal/model"
	"github.com/CaribouBlue/top-spot/internal/server/middleware"
)

type ProfileMux struct {
	*http.ServeMux
}

func NewProfileMux() *ProfileMux {
	mux := &ProfileMux{http.NewServeMux()}
	mux.RegisterHandlers()
	return mux
}

func (mux *ProfileMux) RegisterHandlers() {
	mux.Handle("GET /", http.HandlerFunc(mux.handleProfile))
}

func (mux *ProfileMux) handleProfile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserCtxKey).(*model.UserModel)
	spotify := authorizedSpotifyClient(user)

	profile, err := spotify.GetCurrentUserProfile()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to get current user profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(profile); err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
	}
}
