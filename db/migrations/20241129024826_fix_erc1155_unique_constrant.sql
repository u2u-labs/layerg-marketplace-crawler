-- +goose Up
-- +goose StatementBegin

DROP INDEX IF EXISTS UC_ERC1155 CASCADE;

ALTER TABLE erc_1155_collection_assets 
ADD CONSTRAINT UC_ERC1155_OWNER UNIQUE (asset_id, chain_id, token_id, owner);


-- Update total supply for all assets
CREATE VIEW erc_1155_total_supply AS (
  SELECT 
    asset_id,
    token_id,
    attributes,
    SUM(balance) AS total_supply
  FROM 
    erc_1155_collection_assets
  GROUP BY 
    asset_id, token_id, attributes
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP VIEW IF EXISTS erc_1155_total_supply CASCADE;
DROP INDEX IF EXISTS UC_ERC1155_OWNER CASCADE;
-- +goose StatementEnd
