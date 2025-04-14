-- +goose Up
-- +goose StatementBegin
ALTER TABLE "NFT" ADD COLUMN  IF NOT EXISTS "totalSupply" INTEGER DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
