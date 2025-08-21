-- +goose Up
-- +goose StatementBegin
ALTER TABLE schedule ADD COLUMN check_function VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE schedule DROP COLUMN check_function;
-- +goose StatementEnd
