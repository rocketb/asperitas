package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/rocketb/asperitas/cmd/tools/asperitas-admin/commands"
	database "github.com/rocketb/asperitas/pkg/database/pgx"
	"github.com/rocketb/asperitas/pkg/logger"
	"github.com/rocketb/asperitas/pkg/vault"

	"github.com/ardanlabs/conf/v3"
)

var build = "develop"

type config struct {
	conf.Version
	Args conf.Args
	DB   struct {
		User         string `conf:"default:postgres"`
		Password     string `conf:"default:postgres,mask"`
		Host         string `conf:"default:db"`
		Name         string `conf:"defult:postgres"`
		MaxIdleConns int    `conf:"default:2"`
		MaxOpenConns int    `conf:"default:0"`
		DisableTLS   bool   `conf:"default:true"`
	}
	Vault struct {
		Address    string `conf:"default:http://vault:8200"`
		MountPath  string `conf:"default:secret"`
		Token      string `conf:"default:token,mask"`
		KeysFolder string `conf:"default:/deploy/keys/"`
	}
}

func main() {
	log := logger.New(os.Stdout, logger.LevelInfo, "asperitas-tools", func(context.Context) string { return "00000000-0000-0000-0000-000000000000" })

	if err := run(log); err != nil {
		if !errors.Is(err, commands.ErrHelp) {
			fmt.Println("ERROR", err)
		}
		os.Exit(1)
	}
}

func run(log *logger.Logger) error {
	cfg := config{
		Version: conf.Version{
			Build: build,
			Desc:  "copyright information",
		},
	}

	const prefix = "ASPERITAS"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}

		out, err := conf.String(&cfg)
		if err != nil {
			return fmt.Errorf("generating config output: %w", err)
		}
		log.Info(context.Background(), "startup", "config", out)

		return fmt.Errorf("parsing config: %w", err)
	}

	return executeCommand(cfg.Args, log, cfg)
}

// executeCommand executes one of the given args command
func executeCommand(args conf.Args, log *logger.Logger, cfg config) error {
	dbConf := database.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	}

	vaultConfig := vault.Config{
		Address:   cfg.Vault.Address,
		Token:     cfg.Vault.Token,
		MountPath: cfg.Vault.MountPath,
	}

	switch args.Num(0) {
	case "migrate":
		if err := commands.Migrate(dbConf); err != nil {
			return fmt.Errorf("migrating db: %w", err)
		}
	case "seed":
		if err := commands.Seed(dbConf); err != nil {
			return fmt.Errorf("seeding database: %w", err)
		}
	case "useradd":
		name := args.Num(1)
		email := args.Num(2)
		if err := commands.UserAdd(log, dbConf, name, email); err != nil {
			return fmt.Errorf("adding user: %w", err)
		}
	case "genkey":
		if err := commands.GenKey(); err != nil {
			return fmt.Errorf("key generation: %w", err)
		}
	case "vault":
		if err := commands.Vault(vaultConfig, cfg.Vault.KeysFolder); err != nil {
			return fmt.Errorf("setting private key: %w", err)
		}
	case "vault-init":
		if err := commands.VaultInit(vaultConfig); err != nil {
			return fmt.Errorf("vault initialization: %w", err)
		}
	default:
		fmt.Println("migrate:    create the schema in the database")
		fmt.Println("seed:       add data to the database")
		fmt.Println("useradd:    add a new user to the database")
		fmt.Println("genkey:     generate a set of private/public key files")
		fmt.Println("vault:      load app private key into vault")
		fmt.Println("vault-init  initialize new vault instance")
		fmt.Println("provide a command to get more help.")
		return commands.ErrHelp
	}

	return nil
}
