-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS list (
	id SERIAL PRIMARY KEY,
	title VARCHAR(255) NOT NULL,
  content TEXT,

	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  created_by INTEGER NOT NULL,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS list;
-- +goose StatementEnd
