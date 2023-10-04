package usergrp

import (
	"time"

	"github.com/rocketb/asperitas/internal/usecase/user"
	"github.com/rocketb/asperitas/pkg/validate"
)

// AppUser represents application user.
type AppUser struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	PasswordHash []byte   `json:"-"`
	Roles        []string `json:"roles"`
	DateCreated  string   `json:"date_created"`
}

// AppNewUser what we require from user to add new user.
type AppNewUser struct {
	Name     string   `json:"username" validate:"required,min=3,max=64"`
	Password string   `json:"password" validate:"required,min=8,max=256"`
	Roles    []string `json:"roles" validate:"required,dive,oneof=USER ADMIN"`
}

// Validate checks the data in the model is considered clean.
func (app AppNewUser) Validate() error {
	return validate.Check(app)
}

// AppLoginUser what we require from user to login.
type AppLoginUser struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,len=8"`
}

// Validate checks the data in the model is considered clean.
func (app AppLoginUser) Validate() error {
	return validate.Check(app)
}

func toAppUser(usr user.User) AppUser {
	roles := make([]string, len(usr.Roles))

	for i, role := range usr.Roles {
		roles[i] = role.Name()
	}

	return AppUser{
		ID:          usr.ID.String(),
		Name:        usr.Name,
		Roles:       roles,
		DateCreated: usr.DateCreated.Format(time.RFC3339),
	}
}

func toAppUsers(usrs []user.User) []AppUser {
	users := make([]AppUser, len(usrs))
	for i, u := range usrs {
		users[i] = toAppUser(u)
	}
	return users
}

func toCoreNewUser(nu AppNewUser) user.NewUser {
	roles := make([]user.Role, len(nu.Roles))
	for i, roleName := range nu.Roles {
		role, err := user.ParseRole(roleName)
		if err != nil {
			role = user.RoleUser
		}
		roles[i] = role
	}

	return user.NewUser{
		Name:     nu.Name,
		Password: nu.Password,
		Roles:    roles,
	}
}
