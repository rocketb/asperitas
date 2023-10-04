package repo

import (
	"fmt"
	"time"

	"github.com/rocketb/asperitas/internal/usecase/user"
	"github.com/rocketb/asperitas/pkg/database/pgx/dbarray"

	"github.com/google/uuid"
)

// dbUser represents User in the app strorage.
type dbUser struct {
	ID           uuid.UUID      `db:"user_id"`
	Name         string         `db:"name"`
	Roles        dbarray.String `db:"roles"`
	PasswordHash []byte         `db:"password_hash"`
	DateCreated  time.Time      `db:"date_created"`
}

func toDBUser(usr user.User) dbUser {
	roles := make([]string, len(usr.Roles))
	for i, role := range usr.Roles {
		roles[i] = role.Name()
	}

	return dbUser{
		ID:           usr.ID,
		Name:         usr.Name,
		PasswordHash: usr.PasswordHash,
		DateCreated:  usr.DateCreated,
		Roles:        roles,
	}
}

func toUser(dbUser dbUser) (user.User, error) {
	roles := make([]user.Role, len(dbUser.Roles))
	for i, roleName := range dbUser.Roles {
		role, err := user.ParseRole(roleName)
		if err != nil {
			return user.User{}, fmt.Errorf("parse role: %s", roleName)
		}
		roles[i] = role
	}

	return user.User{
		ID:           dbUser.ID,
		Name:         dbUser.Name,
		PasswordHash: dbUser.PasswordHash,
		DateCreated:  dbUser.DateCreated,
		Roles:        roles,
	}, nil
}

func toUsers(dbUsers []dbUser) ([]user.User, error) {
	usrs := make([]user.User, len(dbUsers))
	for i, dbUsr := range dbUsers {
		usr, err := toUser(dbUsr)
		if err != nil {
			return []user.User{}, err
		}
		usrs[i] = usr
	}

	return usrs, nil
}
