-- name: GetOrCreateEvmAccount :one
INSERT INTO evm_account (address, on_sale_count, holding_count)
SELECT
    $1,
    0,
    0
WHERE NOT EXISTS (
    SELECT 1 FROM evm_account
    WHERE address ILIKE $1
)
RETURNING *;

-- name: GetEvmAccounts :many
SELECT * FROM evm_account
LIMIT $1 OFFSET $2;

-- name: GetEvmAccountByAddress :one
SELECT * FROM evm_account WHERE address ILIKE $1;

-- name: GetTotalEvmAccounts :one
SELECT COUNT(*) FROM evm_account;
