package service

import (
	"context"
	"errors"
	"github.com/crazybolillo/eryth/internal/sqlc"
	"github.com/crazybolillo/eryth/pkg/model"
	"github.com/jackc/pgx/v5"
	"net/url"
	"strings"
)

type Location struct {
	Cursor Cursor `json:"cursor"`
}

func locationParseRow(row sqlc.ListLocationsRow) model.Location {
	location := model.Location{
		Endpoint:  row.Endpoint.String,
		UserAgent: row.UserAgent.String,
	}

	id := strings.Split(row.ID, "@")
	if len(id) != 2 {
		location.ID = id[0]
	} else {
		location.ID = id[1]
	}

	var ua string
	rawUA := strings.Replace(row.UserAgent.String, "^", "%", -1)
	ua, err := url.QueryUnescape(rawUA)
	if err != nil {
		ua = rawUA
	}
	location.UserAgent = ua

	var uri string
	rawUri := strings.Replace(row.Uri.String, "^", "%", -1)
	uri, err = url.QueryUnescape(rawUri)
	if err != nil {
		uri = rawUri
	}

	addr := strings.Split(uri, "@")
	if len(addr) != 2 {
		location.Address = addr[0]
	} else {
		location.Address = addr[1]
	}

	return location
}

func (l *Location) Paginate(ctx context.Context, page, size int) (model.LocationPage, error) {
	queries := sqlc.New(l.Cursor)

	rows, err := queries.ListLocations(ctx, sqlc.ListLocationsParams{
		Limit:  int32(size),
		Offset: int32(page),
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return model.LocationPage{}, err
	}

	count, err := queries.CountLocations(ctx)
	if err != nil {
		return model.LocationPage{}, err
	}

	locations := make([]model.Location, len(rows))
	for idx, row := range rows {
		locations[idx] = locationParseRow(row)
	}

	res := model.LocationPage{
		Locations: locations,
		Total:     count,
		Retrieved: len(locations),
	}

	return res, nil
}
