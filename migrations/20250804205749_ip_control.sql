-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS access (
	id SERIAL PRIMARY KEY,

	user_id INTEGER NOT NULL,
  ip_address VARCHAR(45) NOT NULL,


	accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS access;
-- +goose StatementEnd
