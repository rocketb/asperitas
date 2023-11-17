package repo

import (
	"context"
	"sync"
	"testing"

	"github.com/rocketb/asperitas/internal/usecase/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMemory_Add(t *testing.T) {
	type fields struct {
		data      map[uuid.UUID]*user.User
	}
	type args struct {
		user *user.User
	}

	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   error
		wantUsers map[uuid.UUID]*user.User
	}{
		{
			name: "ok",
			fields: fields{
				data: map[uuid.UUID]*user.User{},
			},
			args: args{
				user: &user.User{
					Name:         "name",
				},
			},
			wantUsers: map[uuid.UUID]*user.User{
				{}: {
					Name:         "name",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:   &sync.RWMutex{},
				data: tt.fields.data,
			}

			err := r.Add(context.Background(), tt.args.user)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantUsers, r.data)
		})
	}
}

func TestMemory_GetByID(t *testing.T) {
	type fields struct {
		data map[uuid.UUID]*user.User
	}
	type args struct {
		userID uuid.UUID
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *user.User
		wantErr error
	}{
		{
			name: "get existing user",
			fields: fields{
				data: map[uuid.UUID]*user.User{
					{}: {
						ID: uuid.UUID{},
					},
				},
			},
			args: args{userID: uuid.UUID{}},
			want: &user.User{ID: uuid.UUID{}},
		},
		{
			name:    "err get non existing user",
			fields:  fields{data: map[uuid.UUID]*user.User{}},
			args:    args{userID: uuid.UUID{}},
			wantErr: user.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:   &sync.RWMutex{},
				data: tt.fields.data,
			}

			got, err := r.GetByID(context.Background(), tt.args.userID)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemory_GetByUsername(t *testing.T) {
	type fields struct {
		data map[uuid.UUID]*user.User
	}
	type args struct {
		username string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *user.User
		wantErr error
	}{
		{
			name: "get existing user",
			fields: fields{
				data: map[uuid.UUID]*user.User{
					{}: {
						ID:   uuid.UUID{},
						Name: "username",
					},
				},
			},
			args: args{username: "username"},
			want: &user.User{ID: uuid.UUID{}, Name: "username"},
		},
		{
			name:    "err get non existing user",
			fields:  fields{data: map[uuid.UUID]*user.User{}},
			args:    args{username: "username"},
			wantErr: user.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:   &sync.RWMutex{},
				data: tt.fields.data,
			}

			got, err := r.GetByUsername(context.Background(), tt.args.username)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewMemory(t *testing.T) {
	r := NewMemory()
	assert.NotNil(t, r)
}
