package service

import (
	"context"
	"errors"
	"github.com/crazybolillo/eryth/internal/db"
	"github.com/crazybolillo/eryth/internal/sqlc"
	"github.com/crazybolillo/eryth/pkg/model"
	"github.com/jackc/pgx/v5"
)

type Contact struct {
	Cursor
}

func (c *Contact) Paginate(ctx context.Context, filter model.ContactPageFilter, page, size int) (model.ContactPage, error) {
	queries := sqlc.New(c.Cursor)

	rows, err := queries.ListContacts(ctx, sqlc.ListContactsParams{
		Name:   db.Text(filter.Name),
		Phone:  db.Text(filter.Phone),
		Limit:  int32(size),
		Offset: int32(page),
		Op:     db.Text(filter.Operator),
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return model.ContactPage{}, err
	}

	count, err := queries.CountContacts(ctx, sqlc.CountContactsParams{
		Name:  db.Text(filter.Name),
		Phone: db.Text(filter.Phone),
		Op:    db.Text(filter.Operator),
	})
	if err != nil {
		return model.ContactPage{}, err
	}

	contacts := make([]model.Contact, len(rows))
	for idx, row := range rows {
		contacts[idx] = model.Contact{
			ID:    row.ID,
			Name:  displayNameFromClid(row.Callerid.String),
			Phone: row.Extension.String,
		}
	}
	res := model.ContactPage{
		Total:     count,
		Retrieved: len(rows),
		Contacts:  contacts,
	}

	return res, nil
}
