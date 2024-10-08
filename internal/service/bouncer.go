package service

import (
	"context"
	"github.com/crazybolillo/eryth/internal/db"
	"github.com/crazybolillo/eryth/internal/sqlc"
	"github.com/crazybolillo/eryth/pkg/model"
	"log/slog"
)

type Bouncer struct {
	Cursor
}

func (b *Bouncer) Check(ctx context.Context, endpoint, dialed string) model.BouncerResponse {
	result := model.BouncerResponse{
		Allow: false,
	}

	queries := sqlc.New(b.Cursor)
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

	return model.BouncerResponse{
		Allow:       true,
		Destination: row.ID,
		CallerID:    row.Callerid.String,
	}
}
