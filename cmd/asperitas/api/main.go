package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rocketb/asperitas/internal/handlers"
	"github.com/rocketb/asperitas/internal/web/auth"
	"github.com/rocketb/asperitas/internal/web/debug"
	db "github.com/rocketb/asperitas/pkg/database/pgx"
	"github.com/rocketb/asperitas/pkg/logger"
	"github.com/rocketb/asperitas/pkg/vault"
	"github.com/rocketb/asperitas/pkg/web"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

const (
	// The name of our config file, without the file extension because viper supports many different config file languages.
	defaultConfigFilename = "asperitas_conf"

	// The environment variable prefix of all environment variables bound to our command line flags.
	envPrefix = "ASPERITAS"

	// Replace hyphenated flag names with camelCase in the config file
	replaceHyphenWithCamelCase = false
)

var build = "develop"

// Config represents app configuration.
type Config struct {
	Auth struct {
		KeyStoreFolder string
		ActiveKID      string
	}
	Web struct {
		Address         string
		ReadTimeout     time.Duration
		WriteTimeout    time.Duration
		ShutdownTimeout time.Duration
		IdleTimeout     time.Duration
		DebugAddress    string
	}
	Vault struct {
		Address   string
		Token     string
		MountPath string
	}
	Tempo struct {
		ReporterURI string
		ServiceName string
		Probability float64
	}
	DB struct {
		User         string
		Password     string
		Host         string
		Name         string
		Schema       string
		MaxIdleConns int
		MaxOpenCons  int
		DisableTLS   bool
	}
}

func main() {
	if err := command().Execute(); err != nil {
		os.Exit(1)
	}
}

// command represents app run commands.
func command() *cobra.Command {
	var config Config

	cmd := &cobra.Command{
		Use:   "asperitas",
		Short: "Asperitas web app",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeConfig(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			run(config)
		},
	}
	cmd.Flags().StringVar(&config.Web.Address, "listen", "0.0.0.0:8080", "Network address to accept connections.")
	cmd.Flags().DurationVar(&config.Web.ShutdownTimeout, "shutdown-timeout", 20*time.Second, "Graceful shutdown timeout.")
	cmd.Flags().DurationVar(&config.Web.ReadTimeout, "read-timeout", 5*time.Second, "Read timeout.")
	cmd.Flags().DurationVar(&config.Web.WriteTimeout, "write-timeout", 10*time.Second, "Write timeout")
	cmd.Flags().StringVar(&config.Web.DebugAddress, "debug-listen", "0.0.0.0:4000", "Debug address to listen.")
	cmd.Flags().StringVar(&config.Auth.KeyStoreFolder, "key-store-folder", "deploy/keys/", "Key store folder.")
	cmd.Flags().DurationVar(&config.Web.IdleTimeout, "idle-timeout", 120*time.Second, "Write timeout")
	cmd.Flags().StringVar(&config.Auth.ActiveKID, "active-kid", "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1", "Active kid.")
	cmd.Flags().StringVar(&config.DB.User, "db-user", "postgres", "DB user name.")
	cmd.Flags().StringVar(&config.DB.Password, "db-password", "postgres", "DB password.")
	cmd.Flags().StringVar(&config.DB.Host, "db-host", "localhost", "DB host.")
	cmd.Flags().StringVar(&config.DB.Name, "db-name", "postgres", "DB name.")
	cmd.Flags().IntVar(&config.DB.MaxIdleConns, "db-max-idle-conns", 2, "DB max idle connections.")
	cmd.Flags().IntVar(&config.DB.MaxOpenCons, "db-max-open-conns", 0, "DB max open connections, 0 is unlimited.")
	cmd.Flags().BoolVar(&config.DB.DisableTLS, "db-disable-tls", true, "DB disable tls connection.")
	cmd.Flags().StringVar(&config.Tempo.ServiceName, "tempo-service-name", "asperitas-api", "Tempo service name.")
	cmd.Flags().StringVar(&config.Tempo.ReporterURI, "tempo-reporter-uri", "tempo:4317", "Tempo reporter URI.")
	cmd.Flags().Float64Var(&config.Tempo.Probability, "tempo-probability", 1, "Tempo Probability.")
	cmd.Flags().StringVar(&config.Vault.Address, "vault-addr", "http://vault:8200", "Vault address.")
	cmd.Flags().StringVar(&config.Vault.Token, "vault-token", "token", "Vault token.")
	cmd.Flags().StringVar(&config.Vault.MountPath, "vault-mount-path", "secret", "Vault mount path.")
	return cmd
}

func run(cfg Config) {
	ctx := context.Background()

	// =============================================================
	// Logger init
	log := logger.New(os.Stdout, logger.LevelInfo, "asperitas", web.GetTraceID)

	log.Info(ctx, "starting service", "version", build)
	defer log.Info(ctx, "shutdown complete")

	// =============================================================
	// Authentication init
	log.Info(ctx, "startup", "status", "initializing authentication support")

	// Simple keystore.
	// ks, err := keystore.NewMemoryFS(os.DirFS(cfg.Auth.KeyStoreFolder))
	// if err != nil {
	// 	log.Error(ctx, "reading keys: %v", err)
	// }

	ks, err := vault.New(vault.Config{
		Address:   cfg.Vault.Address,
		MountPath: cfg.Vault.MountPath,
		Token:     cfg.Vault.Token,
	})
	if err != nil {
		log.Error(ctx, "creating vault ks: %v", err)
	}

	authCfg := auth.Config{
		Log:       log,
		KeyLookup: ks,
		ActiveKID: cfg.Auth.ActiveKID,
	}

	authM := auth.New(authCfg)

	// =============================================================
	// Start Tracing Support

	log.Info(ctx, "startup", "status", "initializing tracing support")

	traceProvider, err := startTracing(
		cfg.Tempo.ServiceName,
		cfg.Tempo.ReporterURI,
		cfg.Tempo.Probability,
	)
	if err != nil {
		log.Error(ctx, "starting tracing: %v", err)
	}
	defer func() {
		if err := traceProvider.Shutdown(context.Background()); err != nil {
			log.Error(ctx, "error on tracing shutdown: %v", err)
		}
	}()

	tracer := traceProvider.Tracer("asperitas")

	// =============================================================
	// Start Debug Service

	go func() {
		log.Info(ctx, "startup", "status", "debug v1 router started", "host", cfg.Web.DebugAddress)

		if err := http.ListenAndServe(cfg.Web.DebugAddress, debug.Mux()); err != nil {
			log.Error(ctx, "shutdown", "status", "debug v1 router closed", "host", cfg.Web.DebugAddress, "msg", err)
		}
	}()

	// =============================================================
	// Start app storage

	log.Info(ctx, "startup", "status", "initializing storage support")

	db, err := db.Open(db.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenCons,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		log.Error(ctx, "opening db: %v", err)
	}
	defer func() {
		log.Info(ctx, "shutdown", "status", "stopping db", "host", cfg.DB.Host)
		db.Close()
	}()

	// =============================================================
	// Start http service

	log.Info(ctx, "startup", "status", "initializing api support")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	apiMux := handlers.APIMux(handlers.APIMuxConfig{
		Build:  build,
		Log:    log,
		Auth:   authM,
		DB:     db,
		Tracer: tracer,
	}, handlers.WithCORS("*"))

	srv := http.Server{
		Addr:         cfg.Web.Address,
		Handler:      apiMux,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
	}

	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Info(ctx, "startup", "status", "service started", "host", srv.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	// Blocking main and waiting for shutdown
	select {
	case err := <-serverErrors:
		log.Error(ctx, "server error: %v", err)
	case sig := <-shutdown:
		log.Info(ctx, "start shutdown", "sig", sig)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error(ctx, "error on server shutdown: %v", err)
			if err := srv.Close(); err != nil {
				log.Error(ctx, "error on server close: %v", err)
			}
		}
	}
}

// initializeConfig inits viper config.
func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	// Set the base name of the config file, without the file extension.
	v.SetConfigName(defaultConfigFilename)

	// Set as many paths as you like where viper should look for the
	// config file. We are only looking in the current working directory.
	v.AddConfigPath(".")

	// Attempt to read the config file, gracefully ignoring errors
	// caused by a config file not being found. Return an error
	// if we cannot parse the config file.
	if err := v.ReadInConfig(); err != nil {
		// It's okay if there isn't a config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --number
	// binds to an environment variable STING_NUMBER. This helps
	// avoid conflicts.
	v.SetEnvPrefix(envPrefix)

	// Environment variables can't have dashes in them, so bind them to their equivalent
	// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --favorite-color which we fix in the bindFlags function
	v.AutomaticEnv()

	// Bind the current command's flags to viper
	bindFlags(cmd, v)

	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Determine the naming convention of the flags when represented in the config file
		configName := f.Name
		// If using camelCase in the config file, replace hyphens with a camelCased string.
		// Since viper does case-insensitive comparisons, we don't need to bother fixing the case, and only need to remove the hyphens.
		if replaceHyphenWithCamelCase {
			configName = strings.ReplaceAll(f.Name, "-", "")
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
				fmt.Printf("ERROR: %s", err)
			}
		}
	})
}

// startTracing initializes app tracing
func startTracing(serviceName string, reporterURI string, probability float64) (*trace.TracerProvider, error) {
	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(), // This should be configurable
			otlptracegrpc.WithEndpoint(reporterURI),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating new exporter: %w", err)
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.TraceIDRatioBased(probability)),
		trace.WithBatcher(exporter,
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
			trace.WithBatchTimeout(trace.DefaultScheduleDelay*time.Microsecond),
		),
		trace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
			)),
	)

	otel.SetTracerProvider(traceProvider)

	return traceProvider, nil
}
