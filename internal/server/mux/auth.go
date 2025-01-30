package mux

import (
	"log"
	"net/http"

	"github.com/CaribouBlue/top-spot/internal/entities/music"
	"github.com/CaribouBlue/top-spot/internal/entities/user"
	"github.com/CaribouBlue/top-spot/internal/server/middleware"
	"github.com/CaribouBlue/top-spot/internal/server/utils"
	"github.com/CaribouBlue/top-spot/internal/spotify"
	"github.com/CaribouBlue/top-spot/internal/templates"
)

type AuthMux struct {
	*http.ServeMux
	Opts       AuthMuxOpts
	Services   AuthMuxServices
	Children   AuthMuxChildren
	Middleware []middleware.Middleware
}

type AuthMuxOpts struct {
	PathPrefix       string
	LoginSuccessPath string
}

type AuthMuxServices struct {
	UserService  user.UserService
	MusicService music.MusicService
}

type AuthMuxChildren struct {
}

func NewAuthMux(opts AuthMuxOpts, services AuthMuxServices, middleware []middleware.Middleware, children AuthMuxChildren) *AuthMux {
	mux := &AuthMux{
		http.NewServeMux(),
		opts,
		services,
		children,
		middleware,
	}

	mux.Handle("GET /user/login", http.HandlerFunc(mux.handleUserLoginPage))
	mux.Handle("POST /user/login", http.HandlerFunc(mux.handleUserLoginSubmit))

	mux.Handle("GET /user/sign-up", http.HandlerFunc(mux.handleUserSignUpPage))
	mux.Handle("POST /user/sign-up", http.HandlerFunc(mux.handleUserSignUp))

	mux.Handle("/login", http.HandlerFunc(mux.handleLogin))
	mux.Handle("/logout", http.HandlerFunc(mux.handleLogout))

	mux.Handle("/spotify", http.HandlerFunc(mux.handleSpotifyAuth))
	mux.Handle("/spotify/redirect", http.HandlerFunc(mux.handleSpotifyAuthRedirect))

	return mux
}

func (mux *AuthMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	middleware.Apply(mux.ServeMux, mux.Middleware...).ServeHTTP(w, r)
}

func (mux *AuthMux) handleUserLoginPage(w http.ResponseWriter, r *http.Request) {
	utils.HandleHtmlResponse(r, w, templates.UserLoginPage())
}

func (mux *AuthMux) handleUserSignUpPage(w http.ResponseWriter, r *http.Request) {
	utils.HandleHtmlResponse(r, w, templates.UserSignUpPage())
}

func (mux *AuthMux) handleUserSignUp(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	_, err := mux.Services.UserService.SignUp(username, password)
	if err != nil {
		http.Error(w, "Failed to sign up user", http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Redirect", mux.Opts.PathPrefix+"/user/login")
	w.WriteHeader(http.StatusCreated)
}

func (mux *AuthMux) handleUserLoginSubmit(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	u, err := mux.Services.UserService.Login(username, password)
	if err != nil {
		http.Error(w, "Failed to log in user", http.StatusInternalServerError)
		return
	}

	err = utils.SetAuthCookie(w, u)
	if err != nil {
		log.Default().Print(err)
		http.Error(w, "Failed to set auth cookie", http.StatusInternalServerError)
		return
	}

	err = mux.Services.MusicService.Authenticate(u)
	if err != nil {
		w.Header().Add("HX-Redirect", mux.Opts.PathPrefix+"/spotify")
		w.WriteHeader(http.StatusOK)
		return
	} else {
		w.Header().Add("HX-Redirect", mux.Opts.LoginSuccessPath)
		w.WriteHeader(http.StatusOK)
		return
	}
}

func (mux *AuthMux) handleLogin(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(middleware.UserCtxKey).(*user.User)

	if u.Id == 0 {
		http.Redirect(w, r, mux.Opts.PathPrefix+"/user/login", http.StatusFound)
		return
	}

	err := mux.Services.MusicService.Authenticate(u)
	if err != nil {
		http.Redirect(w, r, mux.Opts.PathPrefix+"/spotify", http.StatusFound)
		return
	} else {
		http.Redirect(w, r, mux.Opts.LoginSuccessPath, http.StatusFound)
		return
	}
}

func (mux *AuthMux) handleLogout(w http.ResponseWriter, r *http.Request) {
	utils.ClearAuthCookie(w)
	http.Redirect(w, r, mux.Opts.PathPrefix+"/user/login", http.StatusFound)
}

func (mux *AuthMux) handleSpotifyAuth(w http.ResponseWriter, r *http.Request) {
	spotify := spotify.DefaultClient()
	userAuthUrl, err := spotify.GetUserAuthUrl()
	if err != nil {
		http.Error(w, "Failed to get user auth url", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, userAuthUrl, http.StatusFound)
}

func (mux *AuthMux) handleSpotifyAuthRedirect(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(middleware.UserCtxKey).(*user.User)
	u, err := mux.Services.UserService.Get(u.Id)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	spotify := spotify.DefaultClient()

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found in request", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, "State not found in request", http.StatusBadRequest)
		return
	}

	err = spotify.GetNewAccessToken(code)
	if err != nil {
		http.Error(w, "Failed to get new access token", http.StatusBadRequest)
		return
	}

	u.SpotifyAccessToken, err = spotify.GetValidAccessToken()
	if err != nil {
		http.Error(w, "Failed to get valid access token", http.StatusInternalServerError)
		return
	}
	mux.Services.UserService.Update(u)

	http.Redirect(w, r, mux.Opts.PathPrefix+"/login", http.StatusFound)
}
