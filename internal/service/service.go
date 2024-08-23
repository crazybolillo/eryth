package service

import (
	"context"
	"github.com/crazybolillo/eryth/internal/sqlc"
	"github.com/jackc/pgx/v5"
)

type Cursor interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	sqlc.DBTX
}
