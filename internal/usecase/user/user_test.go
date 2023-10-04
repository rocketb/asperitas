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

var (
	ErrDefault = errors.New("default err")
	uuuid      = uuid.New()
)

func TestGetAll(t *testing.T) {
	tests := []struct {
		name        string
		users       []User
		wantErr     assert.ErrorAssertionFunc
		userRepoErr error
	}{
		{
			name: "list all",
			users: []User{
				{
					ID: uuid.New(),
				},
			},
			wantErr: assert.NoError,
		},
		{
			name:        "error on getting users",
			wantErr:     assert.Error,
			userRepoErr: errors.New("err"),
			users:       []User{},
		},
	}

	for _, tt := range tests {
		userRepo := NewRepoMock()
		uc := NewCore(userRepo)

		t.Run(tt.name, func(t *testing.T) {
			userRepo.Mock.On("GetAll", context.Background()).Return(tt.users, tt.userRepoErr)

			users, err := uc.GetAll(context.Background())
			if tt.wantErr(t, err) {
				assert.Equal(t, err, tt.userRepoErr)
			}

			assert.Equal(t, tt.users, users)
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name        string
		total       int
		wantErr     assert.ErrorAssertionFunc
		userRepoErr error
	}{
		{
			name:    "count users ok",
			wantErr: assert.NoError,
			total:   1,
		},
		{
			name:        "error on count users",
			wantErr:     assert.Error,
			userRepoErr: errors.New("some err"),
		},
	}

	for _, tt := range tests {
		usersRepo := NewRepoMock()
		uc := NewCore(usersRepo)

		t.Run(tt.name, func(t *testing.T) {
			usersRepo.Mock.On("Count", context.Background()).Return(tt.total, tt.userRepoErr)

			total, err := uc.Count(context.Background())
			if tt.wantErr(t, err) {
				assert.Equal(t, err, tt.userRepoErr)
			}

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
		wantErr assert.ErrorAssertionFunc
		repoErr error
		caseErr error
	}{
		{
			name: "get by id",
			args: args{
				userID: uuuid,
			},
			user: User{
				ID:   uuuid,
				Name: "name",
			},
			wantErr: assert.NoError,
		},
		{
			name: "error on user get",
			args: args{
				userID: uuuid,
			},
			user:    User{},
			wantErr: assert.Error,
			repoErr: ErrDefault,
			caseErr: ErrDefault,
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := &Core{
			UserRepo: repo,
		}

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetByID", context.Background(), tt.args.userID).Return(tt.user, tt.repoErr).Once()
			usr, err := uc.GetByID(context.Background(), tt.args.userID)
			if tt.wantErr(t, err) {
				if tt.repoErr != nil {
					assert.Equal(t, err, tt.caseErr)
				}
			}

			assert.Equal(t, tt.user, usr)
		})
	}
}

func TestGetByIDs(t *testing.T) {
	type args struct {
		userID uuid.UUID
	}
	tests := []struct {
		name    string
		args    args
		users   []User
		wantErr assert.ErrorAssertionFunc
		repoErr error
		caseErr error
	}{
		{
			name: "get by id",
			args: args{
				userID: uuuid,
			},
			users: []User{
				{
					ID:   uuuid,
					Name: "name",
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "error on user get",
			args: args{
				userID: uuuid,
			},
			users:   []User{},
			wantErr: assert.Error,
			repoErr: ErrDefault,
			caseErr: ErrDefault,
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := &Core{
			UserRepo: repo,
		}

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetByIDs", context.Background(), []uuid.UUID{tt.args.userID}).Return(tt.users, tt.repoErr).Once()
			usrs, err := uc.GetByIDs(context.Background(), []uuid.UUID{tt.args.userID})
			if tt.wantErr(t, err) {
				if tt.repoErr != nil {
					assert.Equal(t, err, tt.caseErr)
				}
			}

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
		wantErr assert.ErrorAssertionFunc
		repoErr error
		caseErr error
	}{
		{
			name: "get by user name",
			args: args{
				username: "name",
			},
			user: User{
				ID:   uuuid,
				Name: "name",
			},
			wantErr: assert.NoError,
		},
		{
			name: "error on user get",
			args: args{
				username: "name",
			},
			user:    User{},
			wantErr: assert.Error,
			repoErr: ErrDefault,
			caseErr: ErrDefault,
		},
	}

	for _, tt := range tests {
		repo := NewRepoMock()
		uc := NewCore(repo)

		t.Run(tt.name, func(t *testing.T) {
			repo.Mock.On("GetByUsername", context.Background(), tt.args.username).Return(tt.user, tt.repoErr).Once()
			usr, err := uc.GetByUsername(context.Background(), tt.args.username)
			if tt.wantErr(t, err) {
				if tt.caseErr != nil {
					assert.Equal(t, err, tt.caseErr)
				}
			}

			assert.Equal(t, tt.user, usr)
		})
	}
}

func TestAdd(t *testing.T) {
	curTime := time.Now()
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
		wantErr assert.ErrorAssertionFunc
		repoErr error
		caseErr error
	}{
		{
			name: "add user",
			args: args{
				nu: NewUser{
					Name:     "name",
					Password: "password",
					Roles:    []Role{RoleUser},
				},
			},
			fields: fields{
				uidGen:   func() uuid.UUID { return uuuid },
				passHash: func(password []byte, cost int) ([]byte, error) { return []byte("pass"), nil },
			},
			wantErr: assert.NoError,
		},
		{
			name: "error repo user add",
			args: args{
				nu: NewUser{
					Name:     "name",
					Password: "password",
				},
			},
			fields: fields{
				uidGen:   func() uuid.UUID { return uuuid },
				passHash: func(password []byte, cost int) ([]byte, error) { return []byte(""), nil },
			},
			wantErr: assert.Error,
			repoErr: ErrDefault,
			caseErr: ErrDefault,
		},
		{
			name: "error pass hash gen",
			args: args{
				nu: NewUser{
					Name:     "name",
					Password: "password",
				},
			},
			fields: fields{
				passHash: func(password []byte, cost int) ([]byte, error) { return []byte(""), ErrDefault },
			},
			wantErr: assert.Error,
			caseErr: fmt.Errorf("generating password hash: %w", ErrDefault),
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
			repo.Mock.On("Add", context.Background(), mock.Anything).Return(tt.repoErr).Once()
			usr, err := uc.Add(context.Background(), tt.args.nu, curTime)
			if tt.wantErr(t, err) {
				assert.Equal(t, tt.caseErr, err)
				return
			}

			hash, _ := tt.fields.passHash([]byte(tt.args.nu.Password), bcrypt.DefaultCost)
			expectedUser := User{
				ID:           tt.fields.uidGen(),
				Name:         tt.args.nu.Name,
				PasswordHash: hash,
				DateCreated:  curTime,
				Roles:        []Role{RoleUser},
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
		wantErr assert.ErrorAssertionFunc
		repoErr error
		caseErr error
	}{
		{
			name: "auth success",
			args: args{
				name:     "name",
				password: "password",
			},
			fields: fields{
				passHashComp: func(h, p []byte) error { return nil },
			},
			user: User{
				ID:   uuuid,
				Name: "name",
			},
			wantErr: assert.NoError,
		},
		{
			name: "error on user get",
			args: args{
				name: "name",
			},
			user:    User{},
			wantErr: assert.Error,
			repoErr: ErrDefault,
			caseErr: ErrAuthenticationFailure,
		},
		{
			name: "error pass check",
			user: User{},
			args: args{
				name: "name",
			},
			fields: fields{
				passHashComp: func(h, p []byte) error { return ErrDefault },
			},
			wantErr: assert.Error,
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
			repo.Mock.On("GetByUsername", context.Background(), tt.args.name).Return(tt.user, tt.repoErr).Once()
			usr, err := uc.Authenticate(context.Background(), tt.args.name, tt.args.password)
			if tt.wantErr(t, err) {
				assert.Equal(t, tt.caseErr, err)
			}

			assert.Equal(t, tt.user, usr)
		})
	}
}

// TestUUIDGen just to mock test coverage
func TestUUIDGen(_ *testing.T) {
	uc := NewCore(NewRepoMock())
	uc.uidGen()
}
