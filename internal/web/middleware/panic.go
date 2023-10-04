package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/rocketb/asperitas/internal/web/metrics"
	"github.com/rocketb/asperitas/pkg/web"
)

// Panic recovers from panic and convert panic to an error.
func Panic() web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
			defer func() {
				if r := recover(); r != nil {
					trace := debug.Stack()
					err = fmt.Errorf("PANIC [%v], TRACE[%s]", r, string(trace))

					metrics.AddPanics(ctx)
				}
			}()
			return handler(ctx, w, r)
		}
		return h
	}
	return m
}
