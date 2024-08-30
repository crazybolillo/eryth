-- name: NewMD5Auth :exec
INSERT INTO ps_auths
    (id, auth_type, username, realm, md5_cred)
VALUES
    ($1, 'md5', $2, $3, $4);

-- name: NewAOR :exec
INSERT INTO ps_aors
    (id, max_contacts)
VALUES
    ($1, $2);

-- name: NewEndpoint :one
INSERT INTO ps_endpoints
    (id, transport, aors, auth, context, disallow, allow, callerid, accountcode)
VALUES
    ($1, $2, $1, $1, $3, 'all', $4, $5, $6)
RETURNING sid;

-- name: DeleteEndpoint :one
DELETE FROM ps_endpoints WHERE sid = $1 RETURNING id;

-- name: DeleteAOR :exec
DELETE FROM ps_aors WHERE id = $1;

-- name: DeleteAuth :exec
DELETE FROM ps_auths WHERE id = $1;

-- name: ListEndpoints :many
SELECT
    pe.sid, pe.id, pe.callerid, pe.context, ee.extension
FROM
    ps_endpoints pe
LEFT JOIN
    ery_extension ee
ON ee.endpoint_id = pe.sid
LIMIT $1 OFFSET $2;

-- name: NewExtension :exec
INSERT INTO ery_extension
    (endpoint_id, extension)
VALUES
    ($1, $2);

-- name: GetEndpointByExtension :one
SELECT
    dest.id, src.callerid
FROM
    ps_endpoints dest
INNER JOIN
    ery_extension ee ON dest.sid = ee.endpoint_id
INNER JOIN
    ps_endpoints src ON src.id = $1
WHERE
    ee.extension = $2;

-- name: GetEndpointByID :one
SELECT
    pe.id, pe.accountcode, pe.callerid, pe.context, ee.extension, pe.transport, aor.max_contacts, pe.allow
FROM
    ps_endpoints pe
INNER JOIN
    ery_extension ee ON ee.endpoint_id = pe.sid
INNER JOIN
    ps_aors aor ON aor.id = pe.id
WHERE
    pe.sid = $1;

-- name: CountEndpoints :one
SELECT COUNT(*) FROM ps_endpoints;

-- name: UpdateEndpointBySid :exec
UPDATE
    ps_endpoints
SET
    callerid = $1,
    context = $2,
    transport = $3,
    allow = $4
WHERE
    sid = $5;

-- name: UpdateExtensionByEndpointId :exec
UPDATE
    ery_extension
SET
    extension = $1
WHERE
    endpoint_id = $2;

-- name: UpdateAORById :exec
UPDATE
    ps_aors
SET
    max_contacts = $1
WHERE
    id = $2;

-- name: UpdateMD5AuthById :exec
UPDATE
    ps_auths
SET
    md5_cred = $1
WHERE
    id = $2;

