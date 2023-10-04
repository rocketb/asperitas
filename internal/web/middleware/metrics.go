package middleware

import (
	"context"
	"net/http"

	"github.com/rocketb/asperitas/internal/web/metrics"
	"github.com/rocketb/asperitas/pkg/web"
)

// Metrics update app metrics.
func Metrics() web.Middleware {
	m := func(hanler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx = metrics.Set(ctx)

			err := hanler(ctx, w, r)

			metrics.AddGoroutines(ctx)
			metrics.AddRequest(ctx)

			if err != nil {
				metrics.AddErrors(ctx)
			}

			return err
		}
		return h
	}
	return m
}
