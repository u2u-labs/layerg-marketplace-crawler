-- +goose Up
-- +goose no transaction
-- SET enable_experimental_alter_column_type_general = true;

-- +goose StatementBegin
-- Copy data with explicit casting
UPDATE onchain_histories SET amount_new = amount::DECIMAL(60,18);

-- +goose StatementEnd

-- SET enable_experimental_alter_column_type_general = false;


-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd