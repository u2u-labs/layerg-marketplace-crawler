-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS nft
(
    uid              UUID PRIMARY KEY   DEFAULT gen_random_uuid(),
    "tokenId"        TEXT      NOT NULL,
    name             TEXT      NOT NULL,
    "createdAt"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status           TEXT      NOT NULL,
    "tokenUri"       TEXT      NOT NULL,
    "txCreationHash" TEXT      NOT NULL,
    "creatorId"      TEXT,
    "collectionId"   TEXT      NOT NULL,
    "chainId"        BIGINT    NOT NULL,
    image            TEXT,
    description      TEXT,
    "animationUrl"   TEXT,
    "nameSlug"       TEXT,
    "metricPoint"    BIGINT    NOT NULL,
    "metricDetail"   JSONB     NOT NULL DEFAULT '{
      "VolumeIndividual": 0,
      "UserMetric": 0
    }'::jsonb,
    source           TEXT,
    UNIQUE ("tokenId", "collectionId", "chainId")
);

CREATE INDEX IF NOT EXISTS nft_collection_id ON nft ("collectionId");

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS nft_collection_id;

DROP TABLE IF EXISTS nft;

-- +goose StatementEnd
