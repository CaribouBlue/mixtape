package middleware

import (
	"net/http"
	"slices"
	"time"

	"github.com/CaribouBlue/mixtape/internal/core"
	"github.com/CaribouBlue/mixtape/internal/log"
	"github.com/CaribouBlue/mixtape/internal/log/rlog"
	"github.com/CaribouBlue/mixtape/internal/server/response"
	"github.com/CaribouBlue/mixtape/internal/server/utils"
	"github.com/CaribouBlue/mixtape/internal/spotify"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Middleware func(http.Handler) http.Handler

func Apply(handler http.Handler, middlewares ...Middleware) http.Handler {
	safeMiddlewares := make([]Middleware, len(middlewares))
	copy(safeMiddlewares, middlewares)
	slices.Reverse(safeMiddlewares)
	for _, middleware := range safeMiddlewares {
		handler = middleware(handler)
	}
	return handler
}

func HandlerFunc(handler http.HandlerFunc, middlewares ...Middleware) http.Handler {
	return Apply(http.HandlerFunc(handler), middlewares...)
}

type wrappedLoggerWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedLoggerWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func WithRequestLogging() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrappedWriter := &wrappedLoggerWriter{w, http.StatusOK}

			next.ServeHTTP(wrappedWriter, r)

			var logger *zerolog.Event
			if wrappedWriter.statusCode >= 500 {
				logger = rlog.Logger(r).Error()
			} else {
				logger = rlog.Logger(r).Info()
			}

			logger.
				Int("status", wrappedWriter.statusCode).
				Msg("Request completed")
		})
	}
}

type wrappedCustomNotFoundWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedCustomNotFoundWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	if w.statusCode == http.StatusNotFound {
		return
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *wrappedCustomNotFoundWriter) Write(b []byte) (int, error) {
	if w.statusCode == http.StatusNotFound {
		return len(b), nil
	}

	return w.ResponseWriter.Write(b)
}

func WithCustomNotFoundHandler(notFoundHandler http.Handler) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrappedWriter := &wrappedCustomNotFoundWriter{w, http.StatusOK}

			next.ServeHTTP(wrappedWriter, r)

			if wrappedWriter.statusCode == http.StatusNotFound {
				notFoundHandler.ServeHTTP(w, r)
			}
		})
	}
}

func WithRequestMetadata() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			sessionIdCookie, err := r.Cookie(utils.CookieNameSessionCorrelationId)
			if err != nil || sessionIdCookie == nil || sessionIdCookie.Value == "" {
				sessionId := uuid.New().String()
				http.SetCookie(w, &http.Cookie{
					Name:     utils.CookieNameSessionCorrelationId,
					Value:    sessionId,
					Path:     "/",
					MaxAge:   int((time.Hour * 24).Seconds()),
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteDefaultMode,
				})
			} else {
				err := utils.RefreshCookie(w, r, utils.CookieNameSessionCorrelationId)
				if err != nil {
					log.Logger().Error().Err(err).Msg("Failed to refresh session correlation ID cookie")
				}
			}

			metadata, err := utils.ContextValue(ctx, utils.RequestMetaDataCtxKey)
			if err != nil || metadata == nil {
				metadata = utils.NewRequestMetadata(r)
			}

			ctx = utils.SetContextValue(ctx, utils.RequestMetaDataCtxKey, metadata)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type WithUserOpts struct {
	UserService *core.UserService
}

func WithUser(opts WithUserOpts) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			user, err := utils.ContextValue(r.Context(), utils.UserCtxKey)
			if user != nil && err == nil {
				next.ServeHTTP(w, r)
				return
			}

			ctxUser := &core.UserEntity{}
			authCookieUser, err := utils.ParseAuthCookie(w, r)
			if err == nil {
				storedUser, err := opts.UserService.GetUserById(authCookieUser.Id)
				if err == nil {
					ctxUser = storedUser
				} else if err != core.ErrUserNotFound {
					log.Logger().Error().Err(err).Msg("Failed to get user by ID")
					http.Error(w, "Failed to get user", http.StatusInternalServerError)
					return
				}
			}

			ctx = utils.SetContextValue(ctx, utils.UserCtxKey, ctxUser)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func WithSpotifyClient() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			spotifyClient := spotify.NewDefaultClient()

			// TODO: handle token updates/invalidation
			user, err := utils.ContextValue(r.Context(), utils.UserCtxKey)
			if err == nil && user != nil && user.SpotifyToken != "" {
				_, err := spotifyClient.Reauthenticate(user.SpotifyToken)
				if err != nil {
					log.Logger().Error().Err(err).Msg("Failed to reauthenticate Spotify client")
				}

			}

			ctx = utils.SetContextValue(ctx, utils.SpotifyClientCtxKey, spotifyClient)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type WithEnforcedAuthenticationOpts struct {
	UnauthenticatedRedirectPath string
	UserService                 *core.UserService
}

func WithEnforcedAuthentication(opts WithEnforcedAuthenticationOpts) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			u, err := utils.ContextValue(ctx, utils.UserCtxKey)
			if err != nil || u == nil {
				http.Error(w, "User not found in context, may need to apply WithUser middleware", http.StatusInternalServerError)
				return
			}

			isAuthenticated, err := opts.UserService.IsAuthenticated(u)
			if err != nil {
				http.Error(w, "Failed to check authentication", http.StatusInternalServerError)
				return
			}

			if !isAuthenticated {
				response.HandleRedirect(w, r, opts.UnauthenticatedRedirectPath)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
