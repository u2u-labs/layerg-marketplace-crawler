-- name: GetCrawlingBackfillCrawler :many
SELECT 
    bc.*, 
    a.type, 
    a.initial_block,
    c.latest_block AS current_latest_block
FROM 
    backfill_crawlers AS bc
JOIN 
    assets AS a 
    ON a.chain_id = bc.chain_id 
    AND a.collection_address = bc.collection_address
JOIN
    chains AS c
    ON c.id = a.chain_id
WHERE 
    bc.status = 'CRAWLING';

-- name: UpdateCrawlingBackfill :exec
UPDATE backfill_crawlers
SET 
    status = COALESCE($3, status),            
    current_block = COALESCE($4, current_block)  
WHERE chain_id = $1
AND collection_address = $2;

-- name: AddBackfillCrawler :exec
INSERT INTO backfill_crawlers (
    chain_id, collection_address, current_block, block_scan_interval
)
VALUES (
    $1, $2, $3, $4
) ON CONFLICT ON CONSTRAINT BACKFILL_CRAWLERS_PKEY DO UPDATE SET
    current_block = EXCLUDED.current_block,
    status = 'CRAWLING'
RETURNING *;

-- name: GetCrawlingBackfillCrawlerById :one
SELECT
    bc.*,
    a.type,
    a.initial_block,
    c.latest_block AS current_latest_block
FROM
    backfill_crawlers AS bc
        JOIN
    assets AS a
    ON a.chain_id = bc.chain_id
        AND a.collection_address = bc.collection_address
JOIN
        chains AS c
ON c.id = a.chain_id
WHERE
    bc.chain_id = $1 AND bc.collection_address = $2;
