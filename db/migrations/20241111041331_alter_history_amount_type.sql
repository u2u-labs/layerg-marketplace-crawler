-- +goose Up
-- +goose no transaction
-- SET enable_experimental_alter_column_type_general = true;

-- +goose StatementBegin

-- Rename the new column to the original name
ALTER TABLE onchain_histories RENAME COLUMN amount_new TO amount;
-- +goose StatementEnd

-- SET enable_experimental_alter_column_type_general = false;


-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd