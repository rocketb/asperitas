package middleware

import (
	"context"
	"net/http"

	"github.com/rocketb/asperitas/pkg/web"
)

func Cors(origin string) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Conrol-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Conrol-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")

			return handler(ctx, w, r)
		}
		return h
	}
	return m
}
