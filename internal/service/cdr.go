package service

import (
	"context"
	"errors"
	"github.com/crazybolillo/eryth/internal/sqlc"
	"github.com/crazybolillo/eryth/pkg/model"
	"github.com/jackc/pgx/v5"
	"time"
)

type Cdr struct {
	Cursor Cursor
}

func (c *Cdr) Paginate(ctx context.Context, page, size int) (model.CallRecordPage, error) {
	queries := sqlc.New(c.Cursor)

	rows, err := queries.ListCallRecords(ctx, sqlc.ListCallRecordsParams{
		Limit:  int32(size),
		Offset: int32(page),
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return model.CallRecordPage{}, err
	}

	count, err := queries.CountCallRecords(ctx)
	if err != nil {
		return model.CallRecordPage{}, err
	}

	records := make([]model.CallRecord, len(rows))
	for idx, row := range rows {
		var answer *time.Time
		answerVal, ok := row.Answer.(time.Time)
		if !ok {
			answer = nil
		} else {
			answer = &answerVal
		}

		records[idx] = model.CallRecord{
			ID:          row.ID,
			From:        row.Origin.String,
			To:          row.Destination.String,
			Context:     row.Context.String,
			Start:       row.CallStart.Time,
			Answer:      answer,
			End:         row.CallEnd.Time,
			Duration:    row.Duration.Int32,
			BillSeconds: row.Billsec.Int32,
		}
	}

	res := model.CallRecordPage{
		Records:   records,
		Total:     count,
		Retrieved: len(records),
	}

	return res, nil
}
