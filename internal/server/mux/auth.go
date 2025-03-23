package mux

import (
	"net/http"

	"github.com/CaribouBlue/mixtape/internal/core"
	"github.com/CaribouBlue/mixtape/internal/log/rlog"
	"github.com/CaribouBlue/mixtape/internal/server/middleware"
	"github.com/CaribouBlue/mixtape/internal/server/utils"
	"github.com/CaribouBlue/mixtape/internal/templates"
)

type AuthMux struct {
	Mux[AuthMuxOpts, AuthMuxServices]
}

func (mux *AuthMux) Opts() MuxOpts {
	return mux.opts.MuxOpts
}

type AuthMuxOpts struct {
	MuxOpts
	LoginSuccessPath string
}

type AuthMuxServices struct {
	MuxServices
	UserService *core.UserService
}

func NewAuthMux(opts AuthMuxOpts, services AuthMuxServices, middleware []middleware.Middleware, children []ChildMux) *AuthMux {
	mux := &AuthMux{
		*NewMux(
			opts,
			services,
			children,
			middleware,
		),
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
			rlog.Logger(r).Error().Err(err).Msg("Failed to sign up user")
			http.Error(w, "Failed to sign up user", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Add("HX-Redirect", mux.opts.PathPrefix+"/user/login")
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
		rlog.Logger(r).Error().Err(err).Msg("Failed to log in user")
		http.Error(w, "Failed to log in user", http.StatusInternalServerError)
		return
	}

	err = utils.SetAuthCookie(w, u)
	if err != nil {
		rlog.Logger(r).Error().Err(err).Msg("Failed to set auth cookie")
		http.Error(w, "Failed to set auth cookie", http.StatusInternalServerError)
		return
	}

	utils.HandleRedirect(w, r, mux.opts.PathPrefix+"/login")
}

func (mux *AuthMux) handleLogin(w http.ResponseWriter, r *http.Request) {
	u, err := utils.ContextValue(r.Context(), utils.UserCtxKey)
	if err != nil {
		rlog.Logger(r).Error().Err(err).Msg("Failed to get user from context")
		http.Error(w, "Could not get user data", http.StatusUnauthorized)
		return
	}

	if u.Id == 0 {
		utils.HandleRedirect(w, r, mux.opts.PathPrefix+"/user/login")
		return
	}

	spotify, err := utils.ContextValue(r.Context(), utils.SpotifyClientCtxKey)
	if err != nil || spotify == nil {
		http.Error(w, "Failed to get Spotify client", http.StatusInternalServerError)
		return
	}

	if u.IsAuthenticatedWithSpotify() {
		_, err := spotify.Reauthenticate(u.SpotifyToken)
		if err != nil {
			rlog.Logger(r).Error().Err(err).Msg("Failed to reauthenticate Spotify client")
			http.Error(w, "Failed to authenticate", http.StatusInternalServerError)
			return
		}

		utils.HandleRedirect(w, r, mux.opts.LoginSuccessPath)
		return
	} else {
		utils.HandleRedirect(w, r, mux.opts.PathPrefix+"/spotify")
		return
	}
}

func (mux *AuthMux) handleLogout(w http.ResponseWriter, r *http.Request) {
	err := utils.DeleteCookie(w, r, utils.CookieAuthorization)
	if err != nil {
		rlog.Logger(r).Error().Err(err).Msg("Failed to delete auth cookie")
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}
	utils.HandleRedirect(w, r, mux.opts.PathPrefix+"/user/login")
}

func (mux *AuthMux) handleSpotifyAuth(w http.ResponseWriter, r *http.Request) {
	spotify, err := utils.ContextValue(r.Context(), utils.SpotifyClientCtxKey)
	if err != nil || spotify == nil {
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
	u, err := utils.ContextValue(r.Context(), utils.UserCtxKey)
	if err != nil {
		rlog.Logger(r).Error().Err(err).Msg("Failed to get user from context")
		http.Error(w, "Could not get user data", http.StatusUnauthorized)
		return
	}

	u, err = mux.Services.UserService.GetUserById(u.Id)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	spotify, err := utils.ContextValue(r.Context(), utils.SpotifyClientCtxKey)
	if err != nil || spotify == nil {
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
		rlog.Logger(r).Error().Err(err).Msg("Failed to authenticate Spotify user")
		http.Error(w, "Failed to authenticate Spotify", http.StatusInternalServerError)
		return
	}

	utils.HandleRedirect(w, r, mux.opts.PathPrefix+"/login")
}
