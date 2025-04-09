-- name: GetAAWalletByAddress :one
SELECT *
FROM "AAWallet"
WHERE "aaAddress" = $1;
