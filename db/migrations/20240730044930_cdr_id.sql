-- migrate:up
ALTER TABLE cdr ADD COLUMN id BIGSERIAL PRIMARY KEY;

-- migrate:down

