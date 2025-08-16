-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS schedule (
	id SERIAL PRIMARY KEY,
	name VARCHAR(255) NOT NULL,
	title VARCHAR(255) NOT NULL,
	description TEXT,
	cron_expression VARCHAR(255) NOT NULL,
	enabled BOOLEAN NOT NULL DEFAULT TRUE,

  created_by INTEGER NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

  FOREIGN KEY (created_by) REFERENCES public.user(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS schedule;
-- +goose StatementEnd
