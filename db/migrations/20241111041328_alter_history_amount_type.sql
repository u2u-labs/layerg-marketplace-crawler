-- +goose Up
-- +goose no transaction
SET enable_experimental_alter_column_type_general = true;

-- +goose StatementBegin
ALTER TABLE onchain_histories ALTER column amount TYPE DECIMAL(60,18);
-- +goose StatementEnd

SET enable_experimental_alter_column_type_general = false;


-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd