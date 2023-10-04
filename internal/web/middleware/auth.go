package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/rocketb/asperitas/internal/web/auth"
	"github.com/rocketb/asperitas/internal/web/request"
	"github.com/rocketb/asperitas/pkg/web"

	"github.com/google/uuid"
)

// Authenticate validates a JWT from the `Authoriztion` header.
func Authenticate(a auth.Auth) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			claims, err := a.Authenticate(ctx, r.Header.Get("authorization"))
			if err != nil {
				return auth.NewError("authenticate: failed: %s", err)
			}

			ctx = auth.SetClaims(ctx, claims)

			return handler(ctx, w, r)
		}
		return h
	}
	return m
}

// Authorize validates that an authenticated user has at least one role from specified list.
func Authorize(a auth.Auth, rule string) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			claims := auth.GetClaims(ctx)
			if claims.Subject == "" {
				return auth.NewError("authorize: you are not authorized for that action, no claims")
			}
			var uid uuid.UUID
			id := web.Param(r, "user_id")
			if id != "" {
				var err error
				uid, err = uuid.Parse(id)
				if err != nil {
					return request.NewError(errors.New("invalid ID"), http.StatusBadRequest)
				}
				ctx = auth.SetUserID(ctx, uid)
			}

			if err := a.Authorize(ctx, claims, uid, rule); err != nil {
				return auth.NewError("authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, rule, err)
			}
			return handler(ctx, w, r)
		}
		return h
	}
	return m
}
