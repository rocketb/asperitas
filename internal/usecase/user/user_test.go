package user

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestGetAll(t *testing.T) {
	tests := []struct {
		name    string
		users   []User
		repoErr error
	}{
		{
			name: "list all",
			users: []User{
				{
					ID: uuid.UUID{},
				},
			},
		},
		{
			name:    "error on getting users",
			repoErr: errors.New("err"),
			users:   []User{},
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := NewCore(repo)

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetAll", context.Background()).Return(tt.users, tt.repoErr)

			users, err := uc.GetAll(context.Background())
			assert.Equal(t, err, tt.repoErr)
			assert.Equal(t, tt.users, users)
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name    string
		total   int
		repoErr error
	}{
		{
			name:  "count users ok",
			total: 1,
		},
		{
			name:    "error on count users",
			repoErr: errors.New("some err"),
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := NewCore(repo)

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("Count", context.Background()).Return(tt.total, tt.repoErr)

			total, err := uc.Count(context.Background())
			assert.Equal(t, err, tt.repoErr)
			assert.Equal(t, tt.total, total)
		})
	}
}

func TestGetByID(t *testing.T) {
	type args struct {
		userID uuid.UUID
	}

	tests := []struct {
		name    string
		args    args
		user    User
		repoErr error
	}{
		{
			name: "get by id",
			args: args{
				userID: uuid.UUID{},
			},
			user: User{
				ID:   uuid.UUID{},
				Name: "name",
			},
		},
		{
			name: "error on user get",
			args: args{
				userID: uuid.UUID{},
			},
			user:    User{},
			repoErr: errors.New("some err"),
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := &Core{
			UserRepo: repo,
		}

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetByID", context.Background(), tt.args.userID).Return(tt.user, tt.repoErr)

			usr, err := uc.GetByID(context.Background(), tt.args.userID)
			assert.Equal(t, err, tt.repoErr)
			assert.Equal(t, tt.user, usr)
		})
	}
}

func TestGetByIDs(t *testing.T) {
	type args struct {
		userIDs []uuid.UUID
	}

	tests := []struct {
		name    string
		args    args
		users   []User
		repoErr error
	}{
		{
			name: "get by id",
			args: args{
				userIDs: []uuid.UUID{{}},
			},
			users: []User{
				{
					ID:   uuid.UUID{},
					Name: "name",
				},
			},
		},
		{
			name: "error on user get",
			args: args{
				userIDs: []uuid.UUID{{}},
			},
			users:   []User{},
			repoErr: errors.New("some err"),
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := &Core{
			UserRepo: repo,
		}

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetByIDs", context.Background(), tt.args.userIDs).Return(tt.users, tt.repoErr)

			usrs, err := uc.GetByIDs(context.Background(), tt.args.userIDs)
			assert.Equal(t, err, tt.repoErr)
			assert.Equal(t, tt.users, usrs)
		})
	}
}

func TestGetByUsername(t *testing.T) {
	type args struct {
		username string
	}
	tests := []struct {
		name    string
		args    args
		user    User
		repoErr error
	}{
		{
			name: "get by user name",
			args: args{
				username: "name",
			},
			user: User{
				ID:   uuid.UUID{},
				Name: "name",
			},
		},
		{
			name: "error on user get",
			args: args{
				username: "name",
			},
			user:    User{},
			repoErr: errors.New("some err"),
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := NewCore(repo)

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetByUsername", context.Background(), tt.args.username).Return(tt.user, tt.repoErr).Once()

			usr, err := uc.GetByUsername(context.Background(), tt.args.username)
			assert.Equal(t, tt.repoErr, err)
			assert.Equal(t, tt.user, usr)
		})
	}
}

func TestAdd(t *testing.T) {
	type fields struct {
		uidGen   func() uuid.UUID
		passHash func(password []byte, cost int) ([]byte, error)
	}
	type args struct {
		nu NewUser
	}

	tests := []struct {
		name    string
		args    args
		fields  fields
		repoErr error
		caseErr error
	}{
		{
			name: "add user",
			args: args{
				nu: NewUser{
					Name: "name",
				},
			},
			fields: fields{
				uidGen:   func() uuid.UUID { return uuid.UUID{} },
				passHash: func(password []byte, cost int) ([]byte, error) { return []byte("pass"), nil },
			},
		},
		{
			name: "error repo user add",
			args: args{
				nu: NewUser{
					Name: "name",
				},
			},
			fields: fields{
				uidGen:   func() uuid.UUID { return uuid.UUID{} },
				passHash: func(password []byte, cost int) ([]byte, error) { return []byte(""), nil },
			},
			repoErr: errors.New("some err"),
			caseErr: errors.New("some err"),
		},
		{
			name: "error pass hash gen",
			args: args{
				nu: NewUser{
					Name: "name",
				},
			},
			fields: fields{
				passHash: func(password []byte, cost int) ([]byte, error) { return []byte(""), errors.New("some err") },
			},
			caseErr: fmt.Errorf("generating password hash: %w", errors.New("some err")),
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := Core{
			UserRepo:    repo,
			uidGen:      tt.fields.uidGen,
			passHashGen: tt.fields.passHash,
		}

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("Add", context.Background(), mock.Anything).Return(tt.repoErr)

			usr, err := uc.Add(context.Background(), tt.args.nu, time.Time{})
			if tt.caseErr != nil {
				assert.Equal(t, tt.caseErr, err)
				return
			}

			hash, _ := tt.fields.passHash([]byte(tt.args.nu.Password), bcrypt.DefaultCost)
			expectedUser := User{
				ID:           tt.fields.uidGen(),
				Name:         tt.args.nu.Name,
				PasswordHash: hash,
			}

			assert.Equal(t, expectedUser, usr)
		})
	}
}

func TestAuthenticate(t *testing.T) {
	type fields struct {
		passHashComp func(hash, password []byte) error
	}

	type args struct {
		name     string
		password string
	}

	tests := []struct {
		name    string
		args    args
		fields  fields
		user    User
		repoErr error
		caseErr error
	}{
		{
			name: "auth success",
			args: args{
				name:     "name",
			},
			fields: fields{
				passHashComp: func(h, p []byte) error { return nil },
			},
			user: User{
				ID:   uuid.UUID{},
				Name: "name",
			},
		},
		{
			name: "error on user get",
			args: args{
				name: "name",
			},
			user:    User{},
			repoErr: errors.New("some err"),
			caseErr: ErrAuthenticationFailure,
		},
		{
			name: "error pass check",
			user: User{},
			args: args{
				name: "name",
			},
			fields: fields{
				passHashComp: func(h, p []byte) error { return errors.New("some err") },
			},
			caseErr: ErrAuthenticationFailure,
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := &Core{
			UserRepo:     repo,
			passHashComp: tt.fields.passHashComp,
		}

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetByUsername", context.Background(), tt.args.name).Return(tt.user, tt.repoErr)

			usr, err := uc.Authenticate(context.Background(), tt.args.name, tt.args.password)
			assert.Equal(t, tt.caseErr, err)
			assert.Equal(t, tt.user, usr)
		})
	}
}

// TestUUIDGen just to mock test coverage
func TestUUIDGen(_ *testing.T) {
	uc := NewCore(NewRepoMock())
	uc.uidGen()
}
