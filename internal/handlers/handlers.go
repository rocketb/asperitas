package handlers

import (
	"net/http"
	"os"

	v1 "github.com/rocketb/asperitas/internal/handlers/v1"
	"github.com/rocketb/asperitas/internal/web/auth"
	"github.com/rocketb/asperitas/internal/web/middleware"
	"github.com/rocketb/asperitas/pkg/logger"
	"github.com/rocketb/asperitas/pkg/web"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"
)

// Options represents optional parameters
type Options struct {
	corsOrigin string
}

// WithCORS provides configuration options for CORS.
func WithCORS(origin string) func(opts *Options) {
	return func(opts *Options) {
		opts.corsOrigin = origin
	}
}

// APIMuxConfig contains all configuration required by handlers.
type APIMuxConfig struct {
	Build    string
	Shutdown chan os.Signal
	Log      *logger.Logger
	Auth     auth.Auth
	DB       *sqlx.DB
	Tracer   trace.Tracer
}

// APIMux constructs http handler with all application routes defined.
func APIMux(cfg APIMuxConfig, options ...func(opts *Options)) http.Handler {
	var opts Options
	for _, option := range options {
		option(&opts)
	}

	app := web.NewApp(
		cfg.Shutdown,
		cfg.Tracer,
		cfg.Log,
		middleware.AccessLog(cfg.Log),
		middleware.Errors(cfg.Log),
		middleware.Metrics(),
		middleware.Panic(),
	)

	if opts.corsOrigin != "" {
		app.EnableCORS(middleware.Cors(opts.corsOrigin))
	}

	v1.Routes(app, v1.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		Auth:  cfg.Auth,
		DB:    cfg.DB,
	})

	return app
}
