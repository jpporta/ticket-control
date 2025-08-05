-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS config (
	id SERIAL PRIMARY KEY,
	key VARCHAR(255) NOT NULL UNIQUE,
	value JSONB NOT NULL,

	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS config;
-- +goose StatementEnd
