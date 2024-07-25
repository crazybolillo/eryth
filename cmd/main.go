package main

import (
	"context"
	"github.com/crazybolillo/eryth/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"net/http"
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
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("failed to establish database connection")
		return err
	}
	defer conn.Close(ctx)

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ","),
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))
	r.Use(middleware.AllowContentEncoding("application/json"))

	endpoint := handler.Endpoint{Conn: conn}
	r.Mount("/endpoint", endpoint.Router())

	return http.ListenAndServe(":8080", r)
}
