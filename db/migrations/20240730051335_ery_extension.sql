-- migrate:up
ALTER TABLE ps_endpoints ADD COLUMN sid SERIAL PRIMARY KEY;

CREATE TABLE ery_extension (
    id SERIAL PRIMARY KEY,
    endpoint_id SERIAL NOT NULL,
    extension varchar UNIQUE,
    FOREIGN KEY (endpoint_id) REFERENCES ps_endpoints(sid)
)

-- migrate:down

