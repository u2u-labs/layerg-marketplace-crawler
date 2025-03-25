-- +goose Up
-- +goose StatementBegin

-- Flexible approach to handle different database constraint behaviors
-- First, drop the existing constraint/index

-- PostgreSQL and CockroachDB specific handling
WITH constraint_info AS (
    SELECT
        tc.constraint_name,
        tc.table_name,
        kcu.column_name
    FROM
        information_schema.table_constraints tc
            JOIN
        information_schema.key_column_usage kcu
        ON tc.constraint_name = kcu.constraint_name
    WHERE
        tc.table_name = 'erc_1155_collection_assets'
      AND tc.constraint_type = 'UNIQUE'
      AND tc.constraint_name = 'uc_erc1155'
)
SELECT
    CASE
        WHEN constraint_name IS NOT NULL
            THEN 'ALTER TABLE erc_1155_collection_assets DROP CONSTRAINT ' || constraint_name || ';'
        ELSE 'SELECT 1;'
        END AS drop_statement
FROM
    constraint_info;

-- Add the new unique constraint
ALTER TABLE erc_1155_collection_assets
    ADD CONSTRAINT UC_ERC1155_OWNER UNIQUE (asset_id, chain_id, token_id, owner);

-- Create the new view
CREATE OR REPLACE VIEW erc_1155_total_supply AS
SELECT
    asset_id,
    token_id,
    attributes,
    SUM(balance) AS total_supply
FROM erc_1155_collection_assets
GROUP BY asset_id, token_id, attributes;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS erc_1155_total_supply CASCADE;
ALTER TABLE erc_1155_collection_assets DROP CONSTRAINT IF EXISTS UC_ERC1155_OWNER;
-- +goose StatementEnd
