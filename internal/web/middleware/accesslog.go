package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/rocketb/asperitas/pkg/logger"
	"github.com/rocketb/asperitas/pkg/web"
)

// AccessLog writes information about the request to the logs.
func AccessLog(log *logger.Logger) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			v := web.GetValues(ctx)

			log.Info(ctx, "request started", "trace_id", r.Method, "path", r.URL.Path,
				"remote_addr", r.RemoteAddr)

			err := handler(ctx, w, r)

			log.Info(ctx, "request completed", "trace_id", r.Method, "path", r.URL.Path,
				"remote_addr", r.RemoteAddr, "status_code", v.StatusCode, "since", time.Since(v.Now))

			return err
		}
		return h
	}
	return m
}
