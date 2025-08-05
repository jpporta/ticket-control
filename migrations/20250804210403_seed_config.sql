-- +goose Up
-- +goose StatementBegin
INSERT INTO config (key, value) VALUES
('thermal-printer', '{"ip": "192.168.3.225", "port": 9100, "enabled": true}');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM config WHERE key = 'thermal-printer';
-- +goose StatementEnd
