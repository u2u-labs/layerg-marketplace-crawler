-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    chains (
        id INT PRIMARY KEY,
        chain VARCHAR NOT NULL,
        name VARCHAR NOT NULL,
        rpc_url VARCHAR NOT NULL,
        chain_id BIGINT NOT NULL,
        explorer VARCHAR NOT NULL,
        latest_block BIGINT NOT NULL,
        block_time INT NOT NULL
    );

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS chains;

-- +goose StatementEnd