package middleware

import (
	"context"
	"log"
	"net/http"
	"slices"

	"github.com/CaribouBlue/top-spot/internal/entities/user"
	"github.com/CaribouBlue/top-spot/internal/server/utils"
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

			path := r.URL.Path
			method := r.Method

			next.ServeHTTP(wrappedWriter, r)

			log.Println(wrappedWriter.statusCode, method, path)
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

			metadata, ok := ctx.Value(utils.RequestMetaDataCtxKey).(utils.RequestMetadata)
			if !ok {
				metadata = utils.NewRequestMetadata(r)
			}

			ctx = context.WithValue(ctx, utils.RequestMetaDataCtxKey, metadata)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type WithUserOpts struct {
	DefaultUserId int64
	UserService   user.UserService
}

func WithUser(opts WithUserOpts) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			_, ok := ctx.Value(utils.UserCtxKey).(*user.User)
			if ok {
				next.ServeHTTP(w, r)
				return
			}

			userCtx := &user.User{}

			userCtx, err := utils.ParseAuthCookie(w, r)
			if err == nil {
				storedUser, err := opts.UserService.Get(userCtx.Id)
				if err == user.ErrNoUserFound {
					userCtx = &user.User{}
				} else if err == nil {
					userCtx = storedUser
				} else {
					http.Error(w, "Failed to get user", http.StatusInternalServerError)
					return
				}
			} else {
				userCtx = &user.User{}
			}

			ctx = context.WithValue(ctx, utils.UserCtxKey, userCtx)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type WithEnforcedAuthenticationOpts struct {
	UnauthenticatedRedirectPath string
	UserService                 user.UserService
}

func WithEnforcedAuthentication(opts WithEnforcedAuthenticationOpts) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			u, ok := ctx.Value(utils.UserCtxKey).(*user.User)
			if !ok || u == nil {
				http.Error(w, "User not found in context, may need to apply WithUser middleware", http.StatusInternalServerError)
				return
			}

			isAuthenticated, err := opts.UserService.IsAuthenticated(u)
			if err != nil {
				http.Error(w, "Failed to check authentication", http.StatusInternalServerError)
				return
			}

			if !isAuthenticated {
				utils.HandleRedirect(w, r, opts.UnauthenticatedRedirectPath)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
