-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS task (
	id SERIAL PRIMARY KEY,
	title VARCHAR(255) NOT NULL,
	description TEXT,
  priority INTEGER CHECK (priority >= 1 AND priority <= 5),

	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  created_by INTEGER NOT NULL,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS task;
-- +goose StatementEnd
