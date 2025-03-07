-- +goose Up
-- +goose StatementBegin
CREATE TYPE order_status AS ENUM ('CANCELED', 'FILLED', 'TRANSFER', 'TRANSFERFILLED');
ALTER TYPE asset_type ADD VALUE 'ORDER';

CREATE TABLE evm_account
(
    id            UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    address       TEXT UNIQUE NOT NULL,
    on_sale_count BIGINT      NOT NULL,
    holding_count BIGINT      NOT NULL,
    created_at    TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE
    order_asset
(
    id         UUID PRIMARY KEY        DEFAULT gen_random_uuid(),
    maker      TEXT           NOT NULL,
    taker      TEXT,
    sig        TEXT           NOT NULL,
    index      INT            NOT NULL,
    status     order_status   NOT NULL,
    take_qty   DECIMAL(78, 0) NOT NULL,
    filled_qty DECIMAL(78, 0) NOT NULL,
    nonce      DECIMAL(78, 0) NOT NULL,
    timestamp  TIMESTAMP      NOT NULL,
    remaining  DECIMAL(78, 0) NOT NULL,
    asset_id   TEXT           NOT NULL,
    chain_id   INT            NOT NULL,
    tx_hash    TEXT           NOT NULL,
    created_at TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (asset_id) REFERENCES assets (id),
    FOREIGN KEY (maker) REFERENCES evm_account (address)
);

CREATE INDEX order_assets_chain_id_idx ON order_asset (chain_id);
CREATE INDEX order_assets_maker_idx ON order_asset (maker);
CREATE INDEX order_assets_taker_idx ON order_asset (taker);
CREATE INDEX order_assets_tx_hash_idx ON order_asset (tx_hash);
CREATE INDEX order_assets_created_at_idx ON order_asset (created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS order_assets_created_at_idx;
DROP INDEX IF EXISTS order_assets_tx_hash_idx;
DROP INDEX IF EXISTS order_assets_taker_idx;
DROP INDEX IF EXISTS order_assets_maker_idx;
DROP INDEX IF EXISTS order_assets_chain_id_idx;

DROP TABLE IF EXISTS order_asset;
DROP TABLE IF EXISTS evm_account;

DROP TYPE IF EXISTS order_status;

-- +goose StatementEnd
