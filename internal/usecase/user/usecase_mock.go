package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type UsecaseMock struct {
	mock.Mock
}

func NewUsecaseMock() *UsecaseMock {
	return &UsecaseMock{}
}

func (r *UsecaseMock) GetAll(ctx context.Context) ([]User, error) {
	args := r.Called(ctx)
	if args.Get(1) != nil {
		return []User{}, args.Error(1)
	}
	return args.Get(0).([]User), args.Error(1)
}

func (r *UsecaseMock) GetByID(ctx context.Context, userID uuid.UUID) (User, error) {
	args := r.Called(ctx, userID)
	if args.Get(1) != nil {
		return User{}, args.Error(1)
	}
	return args.Get(0).(User), args.Error(1)
}

func (r *UsecaseMock) GetByIDs(ctx context.Context, userIDs []uuid.UUID) ([]User, error) {
	args := r.Called(ctx, userIDs)
	if args.Get(1) != nil {
		return []User{}, args.Error(1)
	}

	return args.Get(0).([]User), args.Error(1)
}

func (r *UsecaseMock) GetByUsername(ctx context.Context, username string) (User, error) {
	args := r.Called(ctx, username)
	if args.Get(1) != nil {
		return User{}, args.Error(1)
	}
	return args.Get(0).(User), args.Error(1)
}

func (r *UsecaseMock) Add(ctx context.Context, nu NewUser, now time.Time) (User, error) {
	args := r.Called(ctx, nu, now)
	if args.Get(1) != nil {
		return User{}, args.Error(1)
	}
	return args.Get(0).(User), args.Error(1)
}

func (r *UsecaseMock) Count(ctx context.Context) (int, error) {
	args := r.Called(ctx)
	if args.Get(1) != nil {
		return 0, args.Error(1)
	}

	return args.Get(0).(int), args.Error(1)
}

func (r *UsecaseMock) Authenticate(ctx context.Context, name, password string) (User, error) {
	args := r.Called(ctx, name, password)
	if args.Get(1) != nil {
		return User{}, args.Error(1)
	}
	return args.Get(0).(User), args.Error(1)
}
