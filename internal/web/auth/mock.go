package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Mock struct {
	mock.Mock
}

func NewMock() *Mock {
	return &Mock{}
}

func (m *Mock) GenerateToken(ctx context.Context, claims Claims) (string, error) {
	args := m.Called(ctx, claims)
	if args.Get(1) != nil {
		return "", args.Error(1)
	}
	return args.Get(0).(string), args.Error(1)
}

func (m *Mock) Authenticate(ctx context.Context, barerToken string) (Claims, error) {
	args := m.Called(ctx, barerToken)
	if args.Get(1) != nil {
		return Claims{}, args.Error(1)
	}
	return args.Get(0).(Claims), args.Error(1)
}

func (m *Mock) Authorize(ctx context.Context, claims Claims, userID uuid.UUID, rule string) error {
	args := m.Called(ctx, claims, userID, rule)
	return args.Error(1)
}
