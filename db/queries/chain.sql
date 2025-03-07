-- name: GetChainById :one
SELECT * FROM chains WHERE id = $1;

-- name: GetAllChain :many
SELECT * FROM chains;

-- name: UpdateChainLatestBlock :exec
UPDATE chains
SET
    latest_block = $2
WHERE
    id = $1;

-- name: AddChain :exec
INSERT INTO chains (
    id, chain, name, rpc_url, chain_id,explorer, latest_block, block_time   
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;
