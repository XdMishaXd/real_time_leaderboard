-- +goose Up
-- +goose StatementBegin
INSERT INTO apps (id, name, secret)
VALUES (1, 'app_name', 'super-secret-key')
ON CONFLICT (name) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS apps;
-- +goose StatementEnd
