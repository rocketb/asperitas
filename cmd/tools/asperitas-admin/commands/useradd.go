package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/rocketb/asperitas/internal/usecase/user"
	"github.com/rocketb/asperitas/internal/usecase/user/repo"
	database "github.com/rocketb/asperitas/pkg/database/pgx"
	"github.com/rocketb/asperitas/pkg/logger"
)

func UserAdd(log *logger.Logger, cfg database.Config, name, password string) error {
	if name == "" || password == "" {
		fmt.Println("help: useradd <name> <password>")
		return ErrHelp
	}

	db, err := database.Open(cfg)
	if err != nil {
		return fmt.Errorf("opening db: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userCase := user.NewCore(repo.NewPostgres(db, log))

	nu := user.NewUser{
		Name:     name,
		Password: password,
	}

	usr, err := userCase.Add(ctx, nu, time.Now())
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}

	fmt.Println("user id: ", usr.ID)
	return nil
}
