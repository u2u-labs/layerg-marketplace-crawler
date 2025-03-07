-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    apps (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        name VARCHAR NOT NULL,
        secret_key VARCHAR NOT NULL
    );

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS apps;

-- +goose StatementEnd