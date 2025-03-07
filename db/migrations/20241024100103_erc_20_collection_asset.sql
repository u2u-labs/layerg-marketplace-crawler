-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    erc_20_collection_assets (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        chain_id INT NOT NULL,
        asset_id VARCHAR NOT NULL,
        owner VARCHAR(42) NOT NULL UNIQUE,
        balance DECIMAL(78, 0) NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (asset_id) REFERENCES assets (id)
    );

CREATE INDEX erc_20_collection_assets_asset_id_idx ON erc_20_collection_assets (asset_id, owner);

CREATE INDEX erc_20_collection_assets_owner_idx ON erc_20_collection_assets (chain_id, owner);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX erc_20_collection_assets_asset_id_idx;

DROP INDEX erc_20_collection_assets_owner_idx;

DROP TABLE IF EXISTS erc_20_collection_assets;

-- +goose StatementEnd