-- +goose Up
-- +goose StatementBegin
ALTER TABLE task ADD COLUMN completed_at TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE task DROP COLUMN completed_at;
-- +goose StatementEnd
