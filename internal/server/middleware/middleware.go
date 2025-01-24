package middleware

import (
	"context"
	"log"
	"net/http"
	"slices"

	"github.com/CaribouBlue/top-spot/internal/entities/user"
	"github.com/google/uuid"
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

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func WithRequestLogging() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wrappedWriter := &wrappedWriter{w, http.StatusOK}

			path := r.URL.Path
			method := r.Method

			next.ServeHTTP(wrappedWriter, r)

			log.Println(wrappedWriter.statusCode, method, path)
		})
	}
}

type RequestMetadata struct {
	RequestId string
}

func NewRequestMetadata() RequestMetadata {
	return RequestMetadata{
		RequestId: uuid.New().String(),
	}
}

func WithRequestMetadata() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			metadata, ok := ctx.Value(RequestMetaDataCtxKey).(RequestMetadata)
			if !ok {
				metadata = NewRequestMetadata()
			}

			ctx = context.WithValue(ctx, RequestMetaDataCtxKey, metadata)

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

			_, ok := ctx.Value(UserCtxKey).(*user.User)
			if ok {
				next.ServeHTTP(w, r)
				return
			}

			u, err := opts.UserService.Get(opts.DefaultUserId)
			if err == user.ErrNoUserFound {
				u = &user.User{}
			} else if err != nil {
				log.Print(err)
				http.Error(w, "Failed to get user", http.StatusInternalServerError)
				return
			}

			ctx = context.WithValue(ctx, UserCtxKey, u)

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

			u, ok := ctx.Value(UserCtxKey).(*user.User)
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
				http.Redirect(w, r, opts.UnauthenticatedRedirectPath, http.StatusFound)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
