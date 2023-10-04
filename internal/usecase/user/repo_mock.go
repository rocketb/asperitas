package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock struct {
	mock.Mock
}

func NewRepoMock() *Mock {
	return &Mock{}
}

func (r *Mock) Add(ctx context.Context, nu User) error {
	args := r.Called(ctx, nu)
	return args.Error(0)
}

func (r *Mock) Count(ctx context.Context) (int, error) {
	args := r.Called(ctx)
	if args.Get(1) != nil {
		return 0, args.Error(1)
	}

	return args.Get(0).(int), args.Error(1)
}

func (r *Mock) GetAll(ctx context.Context) ([]User, error) {
	args := r.Called(ctx)
	if args.Get(1) != nil {
		return []User{}, args.Error(1)
	}
	return args.Get(0).([]User), args.Error(1)
}

func (r *Mock) GetByID(ctx context.Context, userID uuid.UUID) (User, error) {
	args := r.Called(ctx, userID)
	if args.Get(1) != nil {
		return User{}, args.Error(1)
	}
	return args.Get(0).(User), args.Error(1)
}

func (r *Mock) GetByIDs(ctx context.Context, userIDs []uuid.UUID) ([]User, error) {
	args := r.Called(ctx, userIDs)
	if args.Get(1) != nil {
		return []User{}, args.Error(1)
	}
	return args.Get(0).([]User), args.Error(1)
}

func (r *Mock) GetByUsername(ctx context.Context, username string) (User, error) {
	args := r.Called(ctx, username)
	if args.Get(1) != nil {
		return User{}, args.Error(1)
	}
	return args.Get(0).(User), args.Error(1)
}
