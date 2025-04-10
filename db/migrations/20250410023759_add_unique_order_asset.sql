-- +goose Up
-- +goose StatementBegin
ALTER TABLE "order_asset" ADD CONSTRAINT "unique_order_asset" UNIQUE ("asset_id", tx_hash);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "order_asset" DROP CONSTRAINT "unique_order_asset";
-- +goose StatementEnd
