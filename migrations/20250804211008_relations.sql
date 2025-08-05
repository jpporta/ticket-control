-- +goose Up
-- +goose StatementBegin
ALTER TABLE task
ADD CONSTRAINT fk_task_created_by
FOREIGN KEY (created_by) REFERENCES public.user(id);

ALTER TABLE list
ADD CONSTRAINT fk_list_created_by
FOREIGN KEY (created_by) REFERENCES public.user(id);

ALTER TABLE access
ADD CONSTRAINT fk_ip_control_user
FOREIGN KEY (user_id) REFERENCES public.user(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
