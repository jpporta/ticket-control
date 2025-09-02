-- +goose Up
-- +goose StatementBegin
ALTER TABLE access ADD COLUMN path VARCHAR(255) NOT NULL DEFAULT '/';
ALTER TABLE access ADD COLUMN method VARCHAR(255) NOT NULL DEFAULT 'GET';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE access DROP COLUMN path;
ALTER TABLE access DROP COLUMN method;
-- +goose StatementEnd
