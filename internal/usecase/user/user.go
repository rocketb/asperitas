package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound              = errors.New("not found")
	ErrAuthenticationFailure = errors.New("authentication failure")
	ErrAlreadyExists         = errors.New("already exists")
)

type Core struct {
	UserRepo     Repo
	uidGen       func() uuid.UUID
	passHashGen  func(password []byte, cost int) ([]byte, error)
	passHashComp func(hash, password []byte) error
}

func NewCore(userRepo Repo) *Core {
	return &Core{
		UserRepo:     userRepo,
		uidGen:       uuid.New,
		passHashGen:  bcrypt.GenerateFromPassword,
		passHashComp: bcrypt.CompareHashAndPassword,
	}
}

// GetAll list all app users.
func (u *Core) GetAll(ctx context.Context) ([]User, error) {
	usrs, err := u.UserRepo.GetAll(ctx)
	if err != nil {
		return []User{}, err
	}

	return usrs, nil
}

// Count returns total number of users in the app.
func (u *Core) Count(ctx context.Context) (int, error) {
	total, err := u.UserRepo.Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// GetByID finds user by user id in the app.
func (u *Core) GetByID(ctx context.Context, userID uuid.UUID) (User, error) {
	usr, err := u.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return User{}, err
	}

	return usr, err
}

// GetByIDs finds users by user IDs.
func (u *Core) GetByIDs(ctx context.Context, userIDs []uuid.UUID) ([]User, error) {
	usrs, err := u.UserRepo.GetByIDs(ctx, userIDs)
	if err != nil {
		return []User{}, err
	}

	return usrs, err
}

// GetByUsername finds user by username in the app.
func (u *Core) GetByUsername(ctx context.Context, username string) (User, error) {
	usr, err := u.UserRepo.GetByUsername(ctx, username)
	if err != nil {
		return User{}, err
	}

	return usr, err
}

// Add user to the app with passed credentials.
func (u *Core) Add(ctx context.Context, nu NewUser, now time.Time) (User, error) {
	hash, err := u.passHashGen([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("generating password hash: %w", err)
	}

	usr := User{
		ID:           u.uidGen(),
		Name:         nu.Name,
		PasswordHash: hash,
		DateCreated:  now,
		Roles:        nu.Roles,
	}

	if err := u.UserRepo.Add(ctx, usr); err != nil {
		return User{}, err
	}
	return usr, nil
}

// Authenticate finds a user by their email and verifies their password.
// On success, it returns a Claims User representing this user. The claims can be
// used to generate a token for future authentication.
func (u *Core) Authenticate(ctx context.Context, name, password string) (User, error) {
	usr, err := u.UserRepo.GetByUsername(ctx, name)
	if err != nil {
		return User{}, ErrAuthenticationFailure
	}

	if err := u.passHashComp(usr.PasswordHash, []byte(password)); err != nil {
		return User{}, ErrAuthenticationFailure
	}

	return usr, nil
}
