-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    erc_1155_collection_assets (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        chain_id INT NOT NULL,
        asset_id VARCHAR NOT NULL,
        token_id DECIMAL(78, 0) NOT NULL,
        owner VARCHAR(42) NOT NULL,
        balance DECIMAL(78, 0) NOT NULL,
        attributes VARCHAR,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (asset_id) REFERENCES assets (id),
        CONSTRAINT UC_ERC1155 UNIQUE (asset_id, chain_id, token_id)
    );

CREATE INDEX erc_1155_collection_assets_chain_id_idx ON erc_1155_collection_assets (asset_id, token_id);

CREATE INDEX erc_1155_collection_assets_owner_idx ON erc_1155_collection_assets (chain_id, owner);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX erc_1155_collection_assets_chain_id_idx;

DROP INDEX erc_1155_collection_assets_owner_idx;

DROP TABLE IF EXISTS erc_1155_collection_assets;

-- +goose StatementEnd