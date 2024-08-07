package main

import (
	"context"
	"fmt"
	"github.com/crazybolillo/eryth/internal/bouncer"
	"github.com/crazybolillo/eryth/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
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

	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("failed to establish database connection")
		return err
	}
	defer conn.Close(ctx)

	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(httplog.NewLogger("eryth", httplog.Options{
		TimeFieldFormat: "2006/01/02 15:04:05",
		LogLevel:        slog.LevelInfo,
		Concise:         true,
	})))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ","),
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))
	r.Use(middleware.AllowContentEncoding("application/json"))

	endpoint := handler.Endpoint{Conn: conn}
	r.Mount("/endpoint", endpoint.Router())

	checker := &bouncer.Bouncer{Conn: conn}
	authorization := handler.Authorization{Bouncer: checker}
	r.Mount("/bouncer", authorization.Router())

	slog.Info("Listening on :8080")
	return http.ListenAndServe(":8080", r)
}
