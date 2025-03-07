-- +goose Up
-- +goose StatementBegin
CREATE TYPE asset_type AS ENUM ('ERC721', 'ERC1155', 'ERC20');

CREATE TABLE
    assets (
        id VARCHAR PRIMARY KEY,
        chain_id INT NOT NULL,
        collection_address VARCHAR(42) NOT NULL,
        type asset_type NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        decimal_data SMALLINT,
        initial_block BIGINT,
        last_updated TIMESTAMP,
        FOREIGN KEY (chain_id) REFERENCES chains (id),
        CONSTRAINT UC_ASSET_COLLECTION UNIQUE (chain_id, collection_address)
    );

CREATE INDEX assets_chain_id_collection_address_idx ON assets (chain_id);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX assets_chain_id_collection_address_idx;

DROP TABLE assets;

DROP TYPE asset_type;

-- +goose StatementEnd