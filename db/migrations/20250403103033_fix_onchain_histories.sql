-- +goose Up
-- +goose StatementBegin
ALTER TABLE "onchain_histories" ADD CONSTRAINT onchain_history_uq_asset_id_tx_hash UNIQUE (asset_id, tx_hash);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "onchain_histories" DROP CONSTRAINT onchain_history_uq_asset_id_tx_hash;
-- +goose StatementEnd
