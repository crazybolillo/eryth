// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: queries.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const countEndpoints = `-- name: CountEndpoints :one
SELECT COUNT(*) FROM ps_endpoints
`

func (q *Queries) CountEndpoints(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, countEndpoints)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const deleteAOR = `-- name: DeleteAOR :exec
DELETE FROM ps_aors WHERE id = $1
`

func (q *Queries) DeleteAOR(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, deleteAOR, id)
	return err
}

const deleteAuth = `-- name: DeleteAuth :exec
DELETE FROM ps_auths WHERE id = $1
`

func (q *Queries) DeleteAuth(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, deleteAuth, id)
	return err
}

const deleteEndpoint = `-- name: DeleteEndpoint :one
DELETE FROM ps_endpoints WHERE sid = $1 RETURNING id
`

func (q *Queries) DeleteEndpoint(ctx context.Context, sid int32) (string, error) {
	row := q.db.QueryRow(ctx, deleteEndpoint, sid)
	var id string
	err := row.Scan(&id)
	return id, err
}

const getEndpointByExtension = `-- name: GetEndpointByExtension :one
SELECT
    dest.id, src.callerid
FROM
    ps_endpoints dest
INNER JOIN
    ery_extension ee ON dest.sid = ee.endpoint_id
INNER JOIN
    ps_endpoints src ON src.id = $1
WHERE
    ee.extension = $2
`

type GetEndpointByExtensionParams struct {
	ID        string      `json:"id"`
	Extension pgtype.Text `json:"extension"`
}

type GetEndpointByExtensionRow struct {
	ID       string      `json:"id"`
	Callerid pgtype.Text `json:"callerid"`
}

func (q *Queries) GetEndpointByExtension(ctx context.Context, arg GetEndpointByExtensionParams) (GetEndpointByExtensionRow, error) {
	row := q.db.QueryRow(ctx, getEndpointByExtension, arg.ID, arg.Extension)
	var i GetEndpointByExtensionRow
	err := row.Scan(&i.ID, &i.Callerid)
	return i, err
}

const getEndpointByID = `-- name: GetEndpointByID :one
SELECT
    pe.id, pe.callerid, pe.context, ee.extension, pe.transport, aor.max_contacts, pe.allow
FROM
    ps_endpoints pe
INNER JOIN
    ery_extension ee ON ee.endpoint_id = pe.sid
INNER JOIN
    ps_aors aor ON aor.id = pe.id
WHERE
    pe.sid = $1
`

type GetEndpointByIDRow struct {
	ID          string      `json:"id"`
	Callerid    pgtype.Text `json:"callerid"`
	Context     pgtype.Text `json:"context"`
	Extension   pgtype.Text `json:"extension"`
	Transport   pgtype.Text `json:"transport"`
	MaxContacts pgtype.Int4 `json:"max_contacts"`
	Allow       pgtype.Text `json:"allow"`
}

func (q *Queries) GetEndpointByID(ctx context.Context, sid int32) (GetEndpointByIDRow, error) {
	row := q.db.QueryRow(ctx, getEndpointByID, sid)
	var i GetEndpointByIDRow
	err := row.Scan(
		&i.ID,
		&i.Callerid,
		&i.Context,
		&i.Extension,
		&i.Transport,
		&i.MaxContacts,
		&i.Allow,
	)
	return i, err
}

const listEndpoints = `-- name: ListEndpoints :many
SELECT
    pe.sid, pe.id, pe.callerid, pe.context, ee.extension
FROM
    ps_endpoints pe
LEFT JOIN
    ery_extension ee
ON ee.endpoint_id = pe.sid
LIMIT $1 OFFSET $2
`

type ListEndpointsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListEndpointsRow struct {
	Sid       int32       `json:"sid"`
	ID        string      `json:"id"`
	Callerid  pgtype.Text `json:"callerid"`
	Context   pgtype.Text `json:"context"`
	Extension pgtype.Text `json:"extension"`
}

func (q *Queries) ListEndpoints(ctx context.Context, arg ListEndpointsParams) ([]ListEndpointsRow, error) {
	rows, err := q.db.Query(ctx, listEndpoints, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListEndpointsRow
	for rows.Next() {
		var i ListEndpointsRow
		if err := rows.Scan(
			&i.Sid,
			&i.ID,
			&i.Callerid,
			&i.Context,
			&i.Extension,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const newAOR = `-- name: NewAOR :exec
INSERT INTO ps_aors
    (id, max_contacts)
VALUES
    ($1, $2)
`

type NewAORParams struct {
	ID          string      `json:"id"`
	MaxContacts pgtype.Int4 `json:"max_contacts"`
}

func (q *Queries) NewAOR(ctx context.Context, arg NewAORParams) error {
	_, err := q.db.Exec(ctx, newAOR, arg.ID, arg.MaxContacts)
	return err
}

const newEndpoint = `-- name: NewEndpoint :one
INSERT INTO ps_endpoints
    (id, transport, aors, auth, context, disallow, allow, callerid)
VALUES
    ($1, $2, $1, $1, $3, 'all', $4, $5)
RETURNING sid
`

type NewEndpointParams struct {
	ID        string      `json:"id"`
	Transport pgtype.Text `json:"transport"`
	Context   pgtype.Text `json:"context"`
	Allow     pgtype.Text `json:"allow"`
	Callerid  pgtype.Text `json:"callerid"`
}

func (q *Queries) NewEndpoint(ctx context.Context, arg NewEndpointParams) (int32, error) {
	row := q.db.QueryRow(ctx, newEndpoint,
		arg.ID,
		arg.Transport,
		arg.Context,
		arg.Allow,
		arg.Callerid,
	)
	var sid int32
	err := row.Scan(&sid)
	return sid, err
}

const newExtension = `-- name: NewExtension :exec
INSERT INTO ery_extension
    (endpoint_id, extension)
VALUES
    ($1, $2)
`

type NewExtensionParams struct {
	EndpointID int32       `json:"endpoint_id"`
	Extension  pgtype.Text `json:"extension"`
}

func (q *Queries) NewExtension(ctx context.Context, arg NewExtensionParams) error {
	_, err := q.db.Exec(ctx, newExtension, arg.EndpointID, arg.Extension)
	return err
}

const newMD5Auth = `-- name: NewMD5Auth :exec
INSERT INTO ps_auths
    (id, auth_type, username, realm, md5_cred)
VALUES
    ($1, 'md5', $2, $3, $4)
`

type NewMD5AuthParams struct {
	ID       string      `json:"id"`
	Username pgtype.Text `json:"username"`
	Realm    pgtype.Text `json:"realm"`
	Md5Cred  pgtype.Text `json:"md5_cred"`
}

func (q *Queries) NewMD5Auth(ctx context.Context, arg NewMD5AuthParams) error {
	_, err := q.db.Exec(ctx, newMD5Auth,
		arg.ID,
		arg.Username,
		arg.Realm,
		arg.Md5Cred,
	)
	return err
}
