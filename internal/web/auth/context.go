package auth

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey int

const claimKey ctxKey = 1

const userKey ctxKey = 2

// SetClaims stores the claims in the context.
func SetClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, claimKey, claims)
}

// GetClaims returns the claims from the context.
func GetClaims(ctx context.Context) Claims {
	v, ok := ctx.Value(claimKey).(Claims)
	if !ok {
		return Claims{}
	}

	return v
}

// SetUserID strores user ID from the request in the context.
func SetUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userKey, userID)
}

// GetUserID returns user ID from the context.
func GetUserID(ctx context.Context) uuid.UUID {
	v, ok := ctx.Value(userKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}
	}

	return v
}
