package usergrp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rocketb/asperitas/internal/usecase/user"
)

func Test_toCoreNewUser(t *testing.T) {
	tests := []struct {
		name   string
		nu     AppNewUser
		wantNU user.NewUser
	}{
		{
			name: "role exist",
			nu: AppNewUser{
				Roles: []string{"ADMIN"},
			},
			wantNU: user.NewUser{
				Roles: []user.Role{user.RoleAdmin},
			},
		},
		{
			name: "role not exist",
			nu: AppNewUser{
				Roles: []string{"NOT_EXIST"},
			},
			wantNU: user.NewUser{
				Roles: []user.Role{user.RoleUser},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantNU, toCoreNewUser(tt.nu))
		})
	}

}
