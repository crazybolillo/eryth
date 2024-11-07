package main

import (
	"context"
	"fmt"
	"github.com/crazybolillo/eryth/internal/handler"
	"github.com/crazybolillo/eryth/internal/metric"
	"github.com/crazybolillo/eryth/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"
)

// @title Asterisk Administration API
// @version 1.0
// @description API to perform configuration management on Asterisk servers.
// @host localhost:8080
func main() {
	os.Exit(run(context.Background()))
}

func run(ctx context.Context) int {
	err := serve(ctx)
	if err != nil {
		return 1
	}

	return 0
}

func serve(ctx context.Context) error {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		slog.Error("Missing Database URL")
		return fmt.Errorf("missing Database URL")
	}

	u, err := url.Parse(dbUrl)
	if err != nil {
		slog.Error("Invalid database URL", "reason", err.Error())
		return err
	}
	slog.Info(
		"Connecting to database",
		slog.String("host", u.Host),
		slog.String("port", u.Port()),
		slog.String("user", u.User.Username()),
		slog.String("database", u.Path[1:]),
	)

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("failed to establish database connection", slog.String("reason", err.Error()))
		return err
	}

	metricsCtx, metricsCancel := context.WithCancel(ctx)
	metric.WatchDbPool(metricsCtx, pool, time.Second*10)

	defer metricsCancel()
	defer pool.Close()

	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(httplog.NewLogger("eryth", httplog.Options{
		TimeFieldFormat: "2006/01/02 15:04:05",
		LogLevel:        slog.LevelInfo,
		Concise:         true,
	})))
	r.Use(middleware.AllowContentEncoding("application/json"))

	endpoint := handler.Endpoint{Service: &service.EndpointService{Cursor: pool}}
	r.Mount("/endpoints", endpoint.Router())

	authorization := handler.Authorization{Bouncer: &service.Bouncer{Cursor: pool}}
	r.Mount("/bouncer", authorization.Router())

	phonebook := handler.Contact{Service: &service.Contact{Cursor: pool}}
	r.Mount("/contacts", phonebook.Router())

	cdr := handler.Cdr{Service: &service.Cdr{Cursor: pool}}
	r.Mount("/cdr", cdr.Router())

	r.Mount("/metrics", promhttp.Handler())

	listen := os.Getenv("LISTEN_ADDR")
	if listen == "" {
		listen = ":8080"
	}

	slog.Info("Starting server", slog.String("addr", listen))
	err = http.ListenAndServe(listen, r)
	if err != nil {
		slog.Error("Failed to start server", "reason", err.Error())
	}

	return err
}
