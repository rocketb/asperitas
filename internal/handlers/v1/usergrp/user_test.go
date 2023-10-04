package usergrp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rocketb/asperitas/internal/usecase/user"
	"github.com/rocketb/asperitas/internal/web/auth"
	"github.com/rocketb/asperitas/internal/web/paging"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserHandler_Register(t *testing.T) {
	appNewUser := AppNewUser{
		Name:     "name",
		Password: "password",
		Roles:    []string{"USER"},
	}
	tErr := errors.New("some error")

	tests := []struct {
		name        string
		userAddErr  error
		userAuthErr error
		genTokenErr error
		wantErrMsg  string
		newUser     AppNewUser
	}{
		{
			name:    "register success",
			newUser: appNewUser,
		},
		{
			name: "payload decode error",
			newUser: AppNewUser{
				Name:     "u",
				Password: "p",
				Roles:    []string{"USER"},
			},
			wantErrMsg: "unable to decode payload: unable to validate payload: [{\"field\":\"username\",\"error\":\"username must be at least 3 characters in length\"},{\"field\":\"password\",\"error\":\"password must be at least 8 characters in length\"}]",
		},
		{
			name:       "user add error",
			newUser:    appNewUser,
			userAddErr: tErr,
			wantErrMsg: fmt.Errorf("unable to create user: %w", tErr).Error(),
		},
		{
			name:       "user already exists",
			newUser:    appNewUser,
			userAddErr: user.ErrAlreadyExists,
			wantErrMsg: user.ErrAlreadyExists.Error(),
		},
		{
			name:        "token gen error",
			newUser:     appNewUser,
			genTokenErr: tErr,
			wantErrMsg:  fmt.Errorf("generating token: %w", tErr).Error(),
		},
	}

	for _, tt := range tests {
		userUsecase := user.NewUsecaseMock()
		authUsecase := auth.NewMock()

		h := &UserHandler{
			Users: userUsecase,
			Auth:  authUsecase,
		}

		t.Run(tt.name, func(t *testing.T) {
			userUsecase.Mock.On("Add", context.Background(), toCoreNewUser(tt.newUser), mock.Anything).Return(user.User{}, tt.userAddErr).Once()
			userUsecase.Mock.On("Authenticate", context.Background(), tt.newUser.Name, tt.newUser.Password).Return(user.User{}, tt.userAuthErr).Once()
			authUsecase.Mock.On("GenerateToken", context.Background(), mock.Anything).Return("tkn", tt.genTokenErr).Once()

			body, _ := json.Marshal(tt.newUser)
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			w := httptest.NewRecorder()

			err := h.Register(context.Background(), w, r)
			if tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}

			resp := w.Result()
			actualBody, _ := io.ReadAll(resp.Body)
			expectedBody, _ := json.Marshal(struct {
				Token string `json:"token"`
			}{Token: "tkn"})

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, expectedBody, actualBody)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	tUser := AppLoginUser{
		Username: "user",
		Password: "password",
	}

	tests := []struct {
		name        string
		wantErrMsg  string
		userRepoErr error
		genTokenErr error
		usr         AppLoginUser
	}{
		{
			name: "login success",
			usr:  tUser,
		},
		{
			name: "payload decode error",
			usr: AppLoginUser{
				Username: "user",
				Password: "p",
			},
			wantErrMsg: "unable to decode payload",
		},
		{
			name:        "login failure",
			usr:         tUser,
			userRepoErr: user.ErrAuthenticationFailure,
			wantErrMsg:  user.ErrAuthenticationFailure.Error(),
		},
		{
			name:        "user auth error",
			usr:         tUser,
			userRepoErr: errors.New("some error"),
			wantErrMsg:  "unable to authenticate user: some error",
		},
		{
			name:        "token generation error",
			usr:         tUser,
			genTokenErr: errors.New("some error"),
			wantErrMsg:  "generating token: some error",
		},
	}

	for _, tt := range tests {
		userUsecase := user.NewUsecaseMock()
		authUsecase := auth.NewMock()

		h := &UserHandler{
			Users: userUsecase,
			Auth:  authUsecase,
		}

		t.Run(tt.name, func(t *testing.T) {
			userUsecase.Mock.On("Authenticate", context.Background(), tt.usr.Username, tt.usr.Password).Return(user.User{ID: uuid.New(), Name: "name"}, tt.userRepoErr)
			authUsecase.Mock.On("GenerateToken", mock.Anything, mock.Anything).Return("tkn", tt.genTokenErr)

			body, _ := json.Marshal(tt.usr)
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
			w := httptest.NewRecorder()

			err := h.Login(context.Background(), w, r)
			if tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}

			resp := w.Result()
			actualBody, _ := io.ReadAll(resp.Body)
			expectedBody, _ := json.Marshal(struct {
				Token string `json:"token"`
			}{Token: "tkn"})

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, expectedBody, actualBody)
		})
	}
}

func TestUserHandler_List(t *testing.T) {
	tUsers := []user.User{
		{
			Name:  "uname",
			Roles: []user.Role{user.RoleUser},
		},
	}
	tErr := errors.New("some error")

	tests := []struct {
		name             string
		users            []user.User
		qparams          string
		wantResponse     paging.Response[AppUser]
		wantErrMsg       string
		userRepoErr      error
		userCountRepoErr error
	}{
		{
			name:         "list users",
			users:        tUsers,
			wantResponse: paging.NewResponse(toAppUsers(tUsers), 1, 1, 10),
		},
		{
			name:       "pagging error",
			qparams:    "page=x",
			wantErrMsg: "[{\"field\":\"page\",\"error\":\"strconv.Atoi: parsing \\\"x\\\": invalid syntax\"}]",
		},
		{
			name:        "get user repo error",
			userRepoErr: tErr,
			wantErrMsg:  fmt.Errorf("collecting users: %w", tErr).Error(),
		},
		{
			name:             "count users repo error",
			userCountRepoErr: tErr,
			wantErrMsg:       fmt.Errorf("counting users: %w", tErr).Error(),
		},
	}

	for _, tt := range tests {
		userUsecase := user.NewUsecaseMock()

		h := &UserHandler{
			Users: userUsecase,
		}

		t.Run(tt.name, func(t *testing.T) {
			userUsecase.Mock.On("GetAll", context.Background()).Return(tt.users, tt.userRepoErr)
			userUsecase.Mock.On("Count", context.Background()).Return(1, tt.userCountRepoErr)

			r := httptest.NewRequest(http.MethodGet, "/?"+tt.qparams, nil)
			w := httptest.NewRecorder()

			err := h.List(context.Background(), w, r)
			if tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}

			resp := w.Result()
			actualBody, _ := io.ReadAll(resp.Body)
			expectedBody, _ := json.Marshal(tt.wantResponse)

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, expectedBody, actualBody)
		})
	}
}

func TestUserHandler_GetByID(t *testing.T) {
	uid := uuid.New()
	tErr := errors.New("some error")
	dnow := time.Now()

	tests := []struct {
		name         string
		user         user.User
		wantResponse AppUser
		wantErrMsg   string
		userRepoErr  error
	}{
		{
			name: "list users",
			user: user.User{
				ID:          uid,
				Roles:       []user.Role{user.RoleUser},
				DateCreated: dnow,
			},
			wantResponse: AppUser{
				ID:          uid.String(),
				Roles:       []string{"USER"},
				DateCreated: dnow.Format(time.RFC3339),
			},
		},
		{
			name:        "get user repo error",
			userRepoErr: tErr,
			wantErrMsg:  fmt.Errorf("getting user: %w", tErr).Error(),
		},
		{
			name:        "user not found error",
			userRepoErr: user.ErrNotFound,
			wantErrMsg:  user.ErrNotFound.Error(),
		},
	}

	for _, tt := range tests {
		userUsecase := user.NewUsecaseMock()

		h := &UserHandler{
			Users: userUsecase,
		}

		t.Run(tt.name, func(t *testing.T) {
			ctx := auth.SetUserID(context.Background(), tt.user.ID)
			userUsecase.Mock.On("GetByID", ctx, tt.user.ID).Return(tt.user, tt.userRepoErr)

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			err := h.GetByID(ctx, w, r)
			if tt.wantErrMsg != "" {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}

			resp := w.Result()
			actualBody, _ := io.ReadAll(resp.Body)
			expectedBody, _ := json.Marshal(tt.wantResponse)

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, expectedBody, actualBody)
		})
	}
}
