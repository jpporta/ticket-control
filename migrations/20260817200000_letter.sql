-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS letter (
	id SERIAL PRIMARY KEY,
	title VARCHAR(255) NOT NULL DEFAULT '',
	content TEXT NOT NULL,
	recipient VARCHAR(255) NOT NULL DEFAULT '',
	sender VARCHAR(255) NOT NULL DEFAULT '',

	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	created_by INTEGER NOT NULL,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (created_by) REFERENCES public.user(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS letter;
-- +goose StatementEnd
