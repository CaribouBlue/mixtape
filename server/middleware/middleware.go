package middleware

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"slices"

	"github.com/CaribouBlue/top-spot/db"
	"github.com/CaribouBlue/top-spot/model"
)

type middleware func(http.Handler) http.Handler

func Apply(handler http.Handler, middlewares ...middleware) http.Handler {
	slices.Reverse(middlewares)
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

func HandlerFunc(handler http.HandlerFunc, middlewares ...middleware) http.Handler {
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

func WithRequestLogging() middleware {
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

type WithUserOpts struct {
	DefaultUserId int64
}

func WithUser(opts WithUserOpts) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			_, ok := ctx.Value(UserCtxKey).(*model.UserModel)
			if ok {
				next.ServeHTTP(w, r)
				return
			}

			db := db.Global()

			user := model.NewUserModel(db, model.WithId(opts.DefaultUserId))
			err := user.Read()
			if err == sql.ErrNoRows {
				// if user does not exist, continue with empty user data model
			} else if err != nil {
				http.Error(w, "Failed to get user", http.StatusInternalServerError)
				return
			}

			ctx = context.WithValue(ctx, UserCtxKey, user)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type WithEnforcedAuthenticationOpts struct {
	UnauthenticatedRedirectPath string
}

func WithEnforcedAuthentication(opts WithEnforcedAuthenticationOpts) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			user := ctx.Value(UserCtxKey).(*model.UserModel)
			if user == nil {
				http.Error(w, "User not found in context, may need to apply WithUser middleware", http.StatusInternalServerError)
				return
			}

			isAuthenticated, err := user.IsAuthenticated()
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
