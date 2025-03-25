-- +goose Up
-- +goose no transaction
-- SET enable_experimental_alter_column_type_general = true;

-- +goose StatementBegin

-- Drop the old column
ALTER TABLE onchain_histories DROP COLUMN amount;
-- +goose StatementEnd

-- SET enable_experimental_alter_column_type_general = false;


-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd