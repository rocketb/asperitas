package repo

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/rocketb/asperitas/internal/usecase/user"
	db "github.com/rocketb/asperitas/pkg/database/pgx"
	"github.com/rocketb/asperitas/pkg/database/pgx/dbarray"
	"github.com/rocketb/asperitas/pkg/logger"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Postgres represents postgres storage for users data.
type Postgres struct {
	log *logger.Logger
	db  sqlx.ExtContext
}

func NewPostgres(db *sqlx.DB, log *logger.Logger) *Postgres {
	return &Postgres{
		log: log,
		db:  db,
	}
}

// GetAll returns a list of users.
func (r *Postgres) GetAll(ctx context.Context) ([]user.User, error) {
	const q = `
	SELECT
		user_id, name, password_hash, roles, date_created
	FROM
		users
	`

	var dbUsers []dbUser
	if err := db.QuerySlice(ctx, r.log, r.db, q, &dbUsers); err != nil {
		return []user.User{}, fmt.Errorf("selecting all users: %w", err)
	}

	usrs, err := toUsers(dbUsers)
	if err != nil {
		return []user.User{}, err
	}

	return usrs, nil
}

// Count returns total numver of users in the DB.
func (r *Postgres) Count(ctx context.Context) (int, error) {
	const q = `
	SELECT
		count(1)
	FROM
		users
	`

	var count struct {
		Count int `db:"count"`
	}

	if err := db.QueryStruct(ctx, r.log, r.db, q, &count); err != nil {
		return 0, fmt.Errorf("quering total users count: %w", err)
	}

	return count.Count, nil
}

// GetByID finds user by user ID in the app storage.
func (r *Postgres) GetByID(ctx context.Context, userID uuid.UUID) (user.User, error) {
	data := struct {
		ID string `db:"user_id"`
	}{
		ID: userID.String(),
	}

	const q = `
	SELECT
		user_id, name, roles, password_hash, date_created
	FROM
		users
	WHERE
		user_id = :user_id
	`

	var dbUser dbUser
	if err := db.NamedQueryStruct(ctx, r.log, r.db, q, data, &dbUser); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return user.User{}, user.ErrNotFound
		}
		return user.User{}, fmt.Errorf("selecting userID(%q): %w", userID, err)
	}

	usr, err := toUser(dbUser)
	if err != nil {
		return user.User{}, err
	}

	return usr, nil
}

// GetByIDs finds users by user IDs in the app storage.
func (r *Postgres) GetByIDs(ctx context.Context, userIDs []uuid.UUID) ([]user.User, error) {
	ids := make([]string, len(userIDs))
	for i, uid := range userIDs {
		ids[i] = uid.String()
	}

	data := struct {
		UserID interface {
			driver.Valuer
			sql.Scanner
		} `db:"user_id"`
	}{
		UserID: dbarray.Array(ids),
	}

	const q = `
	SELECT
		user_id, name, roles, password_hash, date_created
	FROM
		users
	WHERE
		user_id = ANY(:user_id)
	`

	var dbUsers []dbUser
	if err := db.NamedQuerySlice(ctx, r.log, r.db, q, data, &dbUsers); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("selecting users by ids: %w", err)
	}

	usrs, err := toUsers(dbUsers)
	if err != nil {
		return nil, err
	}

	return usrs, nil
}

// GetByUsername finds user by username in the app storage.
func (r *Postgres) GetByUsername(ctx context.Context, username string) (user.User, error) {
	data := struct {
		Name string `db:"name"`
	}{
		Name: username,
	}

	const q = `
	SELECT
		user_id, name, roles, password_hash, date_created
	FROM
		users
	WHERE
		name = :name
	`

	var dbUser dbUser
	if err := db.NamedQueryStruct(ctx, r.log, r.db, q, data, &dbUser); err != nil {
		if errors.Is(err, db.ErrDBNotFound) {
			return user.User{}, user.ErrNotFound
		}
		return user.User{}, fmt.Errorf("selecting user(%q): %w", username, err)
	}

	usr, err := toUser(dbUser)
	if err != nil {
		return user.User{}, err
	}

	return usr, nil
}

// Add creates user in the app storage and return ID of new user.
func (r *Postgres) Add(ctx context.Context, usr user.User) error {
	const q = `
	INSERT INTO users
		(user_id, name, roles, password_hash, date_created)
	VALUES
		(:user_id, :name, :roles, :password_hash, :date_created)
	`

	if err := db.NamedExecContext(ctx, r.log, r.db, q, toDBUser(usr)); err != nil {
		if errors.Is(err, db.ErrDBDuplicatedEntry) {
			return fmt.Errorf("adding user: %w", user.ErrAlreadyExists)
		}

		return fmt.Errorf("inserting user: %w", err)
	}

	return nil
}
