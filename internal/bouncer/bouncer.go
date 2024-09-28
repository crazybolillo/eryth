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
	CallerID    string `json:"callerid"`
}

type Bouncer struct {
	*pgx.Conn
}

func (b *Bouncer) Check(ctx context.Context, endpoint, dialed string) Response {
	result := Response{
		Allow: false,
	}

	tx, err := b.Begin(ctx)
	if err != nil {
		slog.Error("Unable to start transaction", slog.String("reason", err.Error()))
		return result
	}

	queries := sqlc.New(tx)
	row, err := queries.GetEndpointByExtension(ctx, sqlc.GetEndpointByExtensionParams{
		ID:        endpoint,
		Extension: db.Text(dialed),
	})
	if err != nil {
		slog.Error(
			"Failed to retrieve call target",
			slog.String("from", endpoint),
			slog.String("dialed", dialed),
			slog.String("reason", err.Error()),
		)
		return result
	}

	return Response{
		Allow:       true,
		Destination: row.ID,
		CallerID:    row.Callerid.String,
	}
}
