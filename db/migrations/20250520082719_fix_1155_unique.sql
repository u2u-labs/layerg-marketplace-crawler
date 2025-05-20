-- +goose Up
-- +goose StatementBegin
ALTER TABLE erc_1155_collection_assets DROP CONSTRAINT IF EXISTS UC_ERC1155;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE erc_1155_collection_assets ADD CONSTRAINT UC_ERC1155 UNIQUE (asset_id, chain_id, token_id);
-- +goose StatementEnd
