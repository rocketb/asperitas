package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rocketb/asperitas/internal/data/dbmigrate"
	database "github.com/rocketb/asperitas/pkg/database/pgx"
)

// ErrHelp provides context that help was given.
var ErrHelp = errors.New("provide help")

// Migrate migrates the database schema using the provided configuration.
func Migrate(cfg database.Config) error {
	db, err := database.Open(cfg)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}

	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dbmigrate.Migrate(ctx, db); err != nil {
		return fmt.Errorf("migrating database: %w", err)
	}

	fmt.Println("migration complete")
	return nil
}
