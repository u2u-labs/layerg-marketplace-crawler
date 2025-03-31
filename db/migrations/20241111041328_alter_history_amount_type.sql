-- +goose Up
-- +goose no transaction
-- SET enable_experimental_alter_column_type_general = true;

-- +goose StatementBegin
-- Add a new column with the desired type
ALTER TABLE onchain_histories ADD COLUMN amount_new DECIMAL(60,18) NOT NULL DEFAULT 0;

-- +goose StatementEnd

-- SET enable_experimental_alter_column_type_general = false;


-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd