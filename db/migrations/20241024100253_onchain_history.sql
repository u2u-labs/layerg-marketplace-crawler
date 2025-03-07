-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    onchain_histories (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        "from" VARCHAR(42) NOT NULL,
        "to" VARCHAR(42) NOT NULL,
        asset_id VARCHAR NOT NULL,
        token_id DECIMAL(78, 0) NOT NULL,
        amount FLOAT NOT NULL,
        tx_hash VARCHAR(66) NOT NULL,
        timestamp TIMESTAMP NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

CREATE INDEX onchain_history_tx_hash_idx ON onchain_histories (tx_hash);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX onchain_history_tx_hash_idx;

DROP TABLE IF EXISTS onchain_histories;

-- +goose StatementEnd