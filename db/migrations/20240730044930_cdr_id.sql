-- migrate:up
ALTER TABLE CDR ADD COLUMN id BIGSERIAL PRIMARY KEY;

-- migrate:down

