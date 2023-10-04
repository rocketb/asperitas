package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/rocketb/asperitas/cmd/asperitas/metrics/collector"
	"github.com/rocketb/asperitas/cmd/asperitas/metrics/publisher"
	expvarsrv "github.com/rocketb/asperitas/cmd/asperitas/metrics/publisher/expvar"
	prometheussrv "github.com/rocketb/asperitas/cmd/asperitas/metrics/publisher/prometheus"
	"github.com/rocketb/asperitas/pkg/logger"

	"github.com/ardanlabs/conf/v3"
)

var build = "develop"

func main() {
	traceIDFunc := func(ctx context.Context) string {
		return "00000000-0000-0000-0000-000000000000"
	}

	log := logger.New(os.Stdout, logger.LevelInfo, "METRICS", traceIDFunc)

	ctx := context.Background()

	if err := run(ctx, log); err != nil {
		log.Error(ctx, "sturtup", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logger.Logger) error {
	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// ====================================
	// Configuration

	cfg := struct {
		conf.Version
		Web struct {
			DebugHost string `conf:"default:0.0.0.0:4001"`
		}
		Prometheus struct {
			Addr            string        `conf:"default:0.0.0.0:3002"`
			Route           string        `conf:"default:/metrics"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		Expvar struct {
			Addr            string        `conf:"default:0.0.0.0:3001"`
			Route           string        `conf:"default:/metrics"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		Collect struct {
			From string `conf:"default:http://app:4000/debug/vars"`
		}
		Publish struct {
			To       string        `conf:"default:console"`
			Interval time.Duration `conf:"default:5s"`
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "metrics",
		},
	}

	const prefix = "METRICS"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// ====================================
	// App starting

	log.Info(ctx, "starting service", "version", build)
	defer log.Info(ctx, "shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Info(ctx, "startup", "config", out)

	// ====================================
	// Start DEBUG service

	log.Info(ctx, "startup", "status", "debug router started", "host", cfg.Web.DebugHost)

	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())

	go func() {
		if err := http.ListenAndServe(cfg.Web.DebugHost, mux); err != nil {
			log.Error(ctx, "shutdown", "status", "debug router closed", "host", cfg.Web.DebugHost, "msg", err)
		}
	}()

	// ====================================
	// Start Prometheus sercice

	prom := prometheussrv.New(log, cfg.Prometheus.Addr, cfg.Prometheus.Route, cfg.Prometheus.ReadTimeout, cfg.Prometheus.WriteTimeout, cfg.Prometheus.IdleTimeout)
	defer prom.Stop(cfg.Prometheus.ShutdownTimeout)

	// ====================================
	// Start expvar Service

	exp := expvarsrv.New(log, cfg.Expvar.Addr, cfg.Expvar.Route, cfg.Expvar.ReadTimeout, cfg.Expvar.WriteTimeout, cfg.Expvar.IdleTimeout)
	defer exp.Stop(cfg.Expvar.ShutdownTimeout)

	// ====================================
	// Start collectors and publishers

	collector, err := collector.New(cfg.Collect.From)
	if err != nil {
		return fmt.Errorf("starting collector: %w", err)
	}

	stdout := publisher.NewStdout(log)

	publish, err := publisher.New(log, collector, cfg.Publish.Interval, prom.Publish, exp.Publish, stdout.Publish)
	if err != nil {
		return fmt.Errorf("starting publisher: %w", err)
	}
	defer publish.Stop()

	// ====================================
	// Shutdown

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	log.Info(ctx, "shutdown", "status", "shutdown started")
	defer log.Info(ctx, "shutdown", "status", "shutdown complete")

	return nil
}
