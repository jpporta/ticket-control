-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS link (
	id SERIAL PRIMARY KEY,
	url TEXT NOT NULL,
	title VARCHAR(255) NOT NULL,
		
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  created_by INTEGER NOT NULL,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (created_by) REFERENCES public.user(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS link;
-- +goose StatementEnd
