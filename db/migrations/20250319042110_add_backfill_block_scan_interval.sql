-- +goose Up
-- +goose StatementBegin
ALTER TABLE backfill_crawlers ADD COLUMN IF NOT EXISTS block_scan_interval BIGINT DEFAULT 100;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE backfill_crawlers DROP COLUMN IF EXISTS block_scan_interval;
-- +goose StatementEnd
