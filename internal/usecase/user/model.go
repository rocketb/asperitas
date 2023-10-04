package user

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User represents core user.
type User struct {
	ID           uuid.UUID
	Name         string
	PasswordHash []byte
	Roles        []Role
	DateCreated  time.Time
}

// NewUser is what we require to add User.
type NewUser struct {
	Name     string
	Password string
	Roles    []Role
}

// Repo represents user storage interface
type Repo interface {
	Add(ctx context.Context, nu User) error
	Count(ctx context.Context) (int, error)
	GetAll(ctx context.Context) ([]User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (User, error)
	GetByIDs(ctx context.Context, userIDs []uuid.UUID) ([]User, error)
	GetByUsername(ctx context.Context, username string) (User, error)
}

// Usecase represents user use cases.
type Usecase interface {
	Add(ctx context.Context, nu NewUser, now time.Time) (User, error)
	Count(ctx context.Context) (int, error)
	GetAll(ctx context.Context) ([]User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (User, error)
	GetByIDs(ctx context.Context, userIDs []uuid.UUID) ([]User, error)
	GetByUsername(cxt context.Context, username string) (User, error)
	Authenticate(ctx context.Context, name, password string) (User, error)
}
