package mux

import (
	"log"
	"net/http"

	"github.com/CaribouBlue/top-spot/internal/core"
	"github.com/CaribouBlue/top-spot/internal/server/middleware"
	"github.com/CaribouBlue/top-spot/internal/server/utils"
	"github.com/CaribouBlue/top-spot/internal/spotify"
	"github.com/CaribouBlue/top-spot/internal/templates"
)

type AuthMux struct {
	*http.ServeMux
	Opts       AuthMuxOpts
	Services   AuthMuxServices
	Middleware []middleware.Middleware
}

type AuthMuxOpts struct {
	PathPrefix       string
	LoginSuccessPath string
}

type AuthMuxServices struct {
	UserService *core.UserService
}

func NewAuthMux(opts AuthMuxOpts, services AuthMuxServices, middleware []middleware.Middleware) *AuthMux {
	mux := &AuthMux{
		http.NewServeMux(),
		opts,
		services,
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
	confirmPassword := r.FormValue("confirm-password")
	accessCode := r.FormValue("access-code")

	_, err := mux.Services.UserService.SignUpNewUser(username, password, confirmPassword, accessCode)
	if err != nil {
		userSignUpFormOpts := templates.UserSignUpFormOpts{
			Username:        username,
			Password:        password,
			ConfirmPassword: confirmPassword,
			AccessCode:      accessCode,
		}

		if err == core.ErrIncorrectAccessCode {
			userSignUpFormOpts.AccessCodeError = "Invalid access code"
			utils.HandleHtmlResponse(r, w, templates.UserSignUpForm(userSignUpFormOpts))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		} else if err == core.ErrUsernameAlreadyExists {
			userSignUpFormOpts.UsernameError = "Username already exists"
			utils.HandleHtmlResponse(r, w, templates.UserSignUpForm(userSignUpFormOpts))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		} else if err == core.ErrPasswordsDoNotMatch {
			userSignUpFormOpts.ConfirmPasswordError = "Passwords do not match"
			utils.HandleHtmlResponse(r, w, templates.UserSignUpForm(userSignUpFormOpts))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		} else {
			http.Error(w, "Failed to sign up user", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Add("HX-Redirect", mux.Opts.PathPrefix+"/user/login")
	w.WriteHeader(http.StatusCreated)
}

func (mux *AuthMux) handleUserLoginSubmit(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	u, err := mux.Services.UserService.LoginUser(username, password)
	if err == core.ErrUserNotFound || err == core.ErrIncorrectPassword {
		http.Error(w, "Invalid login", http.StatusUnprocessableEntity)
		return
	} else if err != nil {
		http.Error(w, "Failed to log in user", http.StatusInternalServerError)
		return
	}

	err = utils.SetAuthCookie(w, u)
	if err != nil {
		log.Default().Print(err)
		http.Error(w, "Failed to set auth cookie", http.StatusInternalServerError)
		return
	}

	spotify, ok := r.Context().Value(utils.SpotifyClientCtxKey).(*spotify.Client)
	if !ok || spotify == nil {
		http.Error(w, "Failed to get Spotify client", http.StatusInternalServerError)
		return
	}

	_, err = spotify.Reauthenticate(u.SpotifyToken)
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
	u := r.Context().Value(utils.UserCtxKey).(*core.UserEntity)

	if u.Id == 0 {
		utils.HandleRedirect(w, r, mux.Opts.PathPrefix+"/user/login")
		return
	}

	spotify, ok := r.Context().Value(utils.SpotifyClientCtxKey).(*spotify.Client)
	if !ok || spotify == nil {
		http.Error(w, "Failed to get Spotify client", http.StatusInternalServerError)
		return
	}

	log.Default().Println("Reauthenticating Spotify client for user:", u.Username, u.SpotifyToken)

	_, err := spotify.Reauthenticate(u.SpotifyToken)
	if err != nil {
		log.Default().Println("Failed to reauthenticate Spotify client:", err)
		utils.HandleRedirect(w, r, mux.Opts.PathPrefix+"/spotify")
		return
	} else {
		utils.HandleRedirect(w, r, mux.Opts.LoginSuccessPath)
		return
	}
}

func (mux *AuthMux) handleLogout(w http.ResponseWriter, r *http.Request) {
	utils.ClearAuthCookie(w)
	utils.HandleRedirect(w, r, mux.Opts.PathPrefix+"/user/login")
}

func (mux *AuthMux) handleSpotifyAuth(w http.ResponseWriter, r *http.Request) {
	spotify, ok := r.Context().Value(utils.SpotifyClientCtxKey).(*spotify.Client)
	if !ok || spotify == nil {
		http.Error(w, "Failed to get Spotify client", http.StatusInternalServerError)
		return
	}

	userAuthUrl, err := spotify.GetUserAuthUrl()
	if err != nil {
		http.Error(w, "Failed to get user auth url", http.StatusInternalServerError)
		return
	}

	utils.HandleRedirect(w, r, userAuthUrl)
}

func (mux *AuthMux) handleSpotifyAuthRedirect(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(utils.UserCtxKey).(*core.UserEntity)

	u, err := mux.Services.UserService.GetUserById(u.Id)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	spotify, ok := r.Context().Value(utils.SpotifyClientCtxKey).(*spotify.Client)
	if !ok || spotify == nil {
		http.Error(w, "Failed to get Spotify client", http.StatusInternalServerError)
		return
	}

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

	token, err := spotify.Authenticate(code)
	if err != nil {
		http.Error(w, "Failed to get new access token", http.StatusBadRequest)
		return
	}

	u.SpotifyToken = token.RefreshToken
	_, err = mux.Services.UserService.AuthenticateSpotify(u.Id, u.SpotifyToken)
	if err != nil {
		log.Default().Println("Failed to authenticate Spotify user:", err)
		http.Error(w, "Failed to authenticate Spotify", http.StatusInternalServerError)
		return
	}

	log.Default().Println("Authenticated Spotify user:", u.Username, u.SpotifyToken)

	utils.HandleRedirect(w, r, mux.Opts.PathPrefix+"/login")
}
