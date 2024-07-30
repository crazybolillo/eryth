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
    (id, transport, aors, auth, context, disallow, allow)
VALUES
    ($1, $2, $1, $1, $3, 'all', $4)
RETURNING sid;

-- name: DeleteEndpoint :exec
DELETE FROM ps_endpoints WHERE id = $1;

-- name: DeleteAOR :exec
DELETE FROM ps_aors WHERE id = $1;

-- name: DeleteAuth :exec
DELETE FROM ps_auths WHERE id = $1;

-- name: ListEndpoints :many
SELECT
    pe.id, pe.context, ee.extension
FROM
    ps_endpoints pe
LEFT JOIN
    ery_extension ee
ON ee.endpoint_id = pe.sid
LIMIT $1;

-- name: NewExtension :exec
INSERT INTO ery_extension
    (endpoint_id, extension)
VALUES
    ($1, $2);

-- name: GetEndpointByExtension :one
SELECT
    ps_endpoints.id
FROM
    ps_endpoints
INNER JOIN
    ery_extension ee on ps_endpoints.sid = ee.endpoint_id
WHERE
    ee.extension = $1;
