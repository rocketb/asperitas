package repo

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rocketb/asperitas/internal/usecase/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var uuuid = uuid.New()

func TestMemory_Add(t *testing.T) {
	curDate := time.Now()

	type fields struct {
		data      map[uuid.UUID]*user.User
		idGenFunc func() uuid.UUID
	}
	type args struct {
		user *user.User
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      string
		wantUsers map[string]*user.User
		wantErr   error
	}{
		{
			name: "ok",
			fields: fields{
				data: map[uuid.UUID]*user.User{},
				idGenFunc: func() uuid.UUID {
					return uuuid
				},
			},
			args: args{
				user: &user.User{
					Name:         "name",
					PasswordHash: []byte("pass"),
					DateCreated:  curDate,
				},
			},
			want: "1",
			wantUsers: map[string]*user.User{
				"1": {
					ID:           uuuid,
					Name:         "name",
					PasswordHash: []byte("pass"),
					DateCreated:  curDate,
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
			if err != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			got, _ := r.GetByID(context.Background(), tt.args.user.ID)
			if got.ID != tt.args.user.ID {
				t.Errorf("Add() got = %v, want %v", got.ID, tt.want)
			}
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
					uuuid: {
						ID: uuuid,
					},
				},
			},
			args: args{userID: uuuid},
			want: &user.User{ID: uuuid},
		},
		{
			name:    "err get non existing user",
			fields:  fields{data: map[uuid.UUID]*user.User{}},
			args:    args{userID: uuuid},
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
			if err != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			assert.Equalf(t, tt.want, got, "GetByID(ctx, %v)", tt.args.userID)
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
					uuuid: {
						ID:   uuuid,
						Name: "username",
					},
				},
			},
			args: args{username: "username"},
			want: &user.User{ID: uuuid, Name: "username"},
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
			if err != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			assert.Equalf(t, tt.want, got, "GetByUsername(ctx, %v)", tt.args.username)
		})
	}
}

func TestNewMemory(t *testing.T) {
	r := NewMemory()
	assert.NotNil(t, r)
}
