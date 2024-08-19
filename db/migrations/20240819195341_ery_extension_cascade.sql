-- migrate:up
ALTER TABLE ery_extension DROP CONSTRAINT ery_extension_endpoint_id_fkey;

ALTER TABLE
    ery_extension
ADD CONSTRAINT
    ry_extension_endpoint_id_fkey
FOREIGN KEY
    (endpoint_id)
REFERENCES
    ps_endpoints(sid)
ON DELETE CASCADE;

-- migrate:down
