-- +goose Up
-- +goose StatementBegin
ALTER TABLE task ADD COLUMN completed_by INTEGER;
ALTER TABLE task ADD CONSTRAINT fk_completed_by FOREIGN KEY (completed_by) REFERENCES public.user(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE task DROP CONSTRAINT fk_completed_by;
ALTER TABLE task DROP COLUMN completed_by;
-- +goose StatementEnd
