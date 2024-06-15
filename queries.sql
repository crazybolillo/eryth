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

-- name: NewEndpoint :exec
INSERT INTO ps_endpoints
    (id, transport, aors, auth, context, disallow, allow)
VALUES
    ($1, $2, $1, $1, $3, 'all', $4);
