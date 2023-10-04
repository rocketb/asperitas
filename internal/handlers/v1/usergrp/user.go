package usergrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rocketb/asperitas/internal/usecase/user"
	"github.com/rocketb/asperitas/internal/web/auth"
	"github.com/rocketb/asperitas/internal/web/paging"
	"github.com/rocketb/asperitas/internal/web/request"
	"github.com/rocketb/asperitas/pkg/logger"
	"github.com/rocketb/asperitas/pkg/web"

	"github.com/golang-jwt/jwt/v4"
)

type UserHandler struct {
	Logger *logger.Logger
	Users  user.Usecase
	Auth   auth.Auth
}

// Register adds new user to the app.
func (h *UserHandler) Register(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var nu AppNewUser
	if err := web.Decode(r, &nu); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	usr, err := h.Users.Add(ctx, toCoreNewUser(nu), time.Now())
	if err != nil {
		if errors.Is(err, user.ErrAlreadyExists) {
			return request.NewError(err, http.StatusUnprocessableEntity)
		}
		return fmt.Errorf("unable to create user: %w", err)
	}

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   usr.ID.String(),
			Issuer:    "asperitas project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		User: auth.User{
			Username: usr.Name,
			ID:       usr.ID,
		},
		Roles: usr.Roles,
	}

	var tkn struct {
		Token string `json:"token"`
	}
	tkn.Token, err = h.Auth.GenerateToken(ctx, claims)
	if err != nil {
		return fmt.Errorf("generating token: %w", err)
	}

	return web.Respond(ctx, w, tkn, http.StatusOK)
}

// Login logins to the app with given credentials and returns JWT token.
func (h *UserHandler) Login(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var u AppLoginUser
	if err := web.Decode(r, &u); err != nil {
		return auth.NewError("unable to decode payload")
	}

	usr, err := h.Users.Authenticate(ctx, u.Username, u.Password)
	if err != nil {
		switch err {
		case user.ErrAuthenticationFailure:
			return request.NewError(err, http.StatusForbidden)
		default:
			return fmt.Errorf("unable to authenticate user: %w", err)
		}
	}
	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   usr.ID.String(),
			Issuer:    "asperitas project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		User: auth.User{
			Username: usr.Name,
			ID:       usr.ID,
		},
		Roles: usr.Roles,
	}

	var tkn struct {
		Token string `json:"token"`
	}
	tkn.Token, err = h.Auth.GenerateToken(ctx, claims)
	if err != nil {
		return fmt.Errorf("generating token: %w", err)
	}

	return web.Respond(ctx, w, tkn, http.StatusOK)
}

// List returns a list of users.
func (h *UserHandler) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page, err := paging.ParseRequest(r)
	if err != nil {
		return err
	}

	users, err := h.Users.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("collecting users: %w", err)
	}

	total, err := h.Users.Count(ctx)
	if err != nil {
		return fmt.Errorf("counting users: %w", err)
	}

	return web.Respond(ctx, w, paging.NewResponse(toAppUsers(users), total, page.Number, page.RowsPerPage), http.StatusOK)
}

// GetByID returns a user bu its ID.
func (h *UserHandler) GetByID(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
	uid := auth.GetUserID(ctx)
	usr, err := h.Users.GetByID(ctx, uid)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNotFound):
			return request.NewError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("getting user: %w", err)
		}
	}

	return web.Respond(ctx, w, toAppUser(usr), http.StatusOK)
}
