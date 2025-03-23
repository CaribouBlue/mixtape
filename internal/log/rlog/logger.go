package rlog

import (
	"net/http"

	"github.com/rs/zerolog"

	"github.com/CaribouBlue/mixtape/internal/log"
	"github.com/CaribouBlue/mixtape/internal/server/utils"
)

func Logger(r *http.Request) *zerolog.Logger {
	path := r.URL.Path
	method := r.Method

	ctx := log.Logger().With().
		Str("path", path).
		Str("method", method)

	metadata, err := utils.ContextValue(r.Context(), utils.RequestMetaDataCtxKey)
	if err == nil && metadata != nil {
		ctx = ctx.Str("requestCorrelationId", metadata.RequestCorrelationId).Str("sessionCorrelationId", metadata.SessionCorrelationId)
	}

	user, err := utils.ContextValue(r.Context(), utils.UserCtxKey)
	if err == nil && user != nil && user.Id != 0 {
		ctx = ctx.Str("username", user.Username).Bool("isAdmin", user.IsAdmin)
	}

	logger := ctx.Logger()
	return &logger
}
