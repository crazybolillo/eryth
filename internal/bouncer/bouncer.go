package bouncer

import (
	"context"
	"github.com/crazybolillo/eryth/internal/db"
	"github.com/crazybolillo/eryth/internal/sqlc"
	"github.com/jackc/pgx/v5"
	"log/slog"
)

type Response struct {
	Allow       bool   `json:"allow"`
	Destination string `json:"destination"`
}

type Bouncer struct {
	*pgx.Conn
}

func (b *Bouncer) Check(ctx context.Context, endpoint, dialed string) Response {
	result := Response{
		Allow:       false,
		Destination: "",
	}

	tx, err := b.Begin(ctx)
	if err != nil {
		slog.Error("Unable to start transaction", slog.String("reason", err.Error()))
		return result
	}

	queries := sqlc.New(tx)
	destination, err := queries.GetEndpointByExtension(ctx, db.Text(dialed))
	if err != nil {
		slog.Error("Failed to retrieve endpoint", slog.String("dialed", dialed), slog.String("reason", err.Error()))
		return result
	}

	return Response{
		Allow:       true,
		Destination: destination,
	}
}
