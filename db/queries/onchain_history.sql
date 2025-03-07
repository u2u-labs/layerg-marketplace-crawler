-- name: AddOnchainTransaction :one
INSERT INTO 
    onchain_histories("from","to",asset_id,token_id,amount,tx_hash,timestamp)
VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;


-- name: GetOnchainHistoriesByTxHash :many
SELECT * FROM onchain_histories WHERE tx_hash = $1;