-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "NFT"
(
    "id"             TEXT      NOT NULL,
    name             TEXT      NOT NULL,
    "createdAt"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status           TEXT      NOT NULL,
    "tokenUri"       TEXT      NOT NULL,
    "txCreationHash" TEXT      NOT NULL,
    "creatorId"      UUID,
    "collectionId"   UUID      NOT NULL,
    image            TEXT,
    "isActive"       BOOLEAN   NOT NULL DEFAULT TRUE,
    description      TEXT,
    "animationUrl"   TEXT,
    "nameSlug"       TEXT,
    "metricPoint"    BIGINT    NOT NULL,
    "metricDetail"   JSONB     NOT NULL DEFAULT '{
      "VolumeIndividual": 0,
      "UserMetric": 0
    }'::jsonb,
    source           TEXT,
    UNIQUE ("id", "collectionId")
);

ALTER TABLE "NFT"
    ADD COLUMN IF NOT EXISTS "chainId" BIGINT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS nft_collection_id ON "NFT" ("collectionId");

CREATE TABLE IF NOT EXISTS "Collection"
(
    "id"             UUID             NOT NULL,
    "txCreationHash" TEXT             NOT NULL,
    "name"           TEXT             NOT NULL,
    "nameSlug"       TEXT,
    "symbol"         TEXT             NOT NULL,
    "description"    TEXT,
    "address"        TEXT,
    "shortUrl"       TEXT,
    "metadata"       TEXT,
    "isU2U"          BOOLEAN          NOT NULL DEFAULT true,
    "status"         "TX_STATUS"      NOT NULL,
    "type"           "CONTRACT_TYPE"  NOT NULL,
    "categoryId"     INTEGER,
    "createdAt"      TIMESTAMP(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"      TIMESTAMP(3)     NOT NULL,
    "coverImage"     TEXT,
    "avatar"         TEXT,
    "projectId"      UUID,
    "isVerified"     BOOLEAN          NOT NULL DEFAULT false,
    "floorPrice"     BIGINT           NOT NULL DEFAULT 0,
    "floor"          DOUBLE PRECISION NOT NULL DEFAULT 0,
    "floorWei"       TEXT             NOT NULL DEFAULT '0',
    "isActive"       BOOLEAN          NOT NULL DEFAULT true,
    "flagExtend"     BOOLEAN                   DEFAULT false,
    "isSync"         BOOLEAN                   DEFAULT true,
    "subgraphUrl"    TEXT,
    "lastTimeSync"   INTEGER                   DEFAULT 0,
    "metricPoint"    BIGINT                    DEFAULT 0,
    "metricDetail"   JSONB                     DEFAULT '{
      "Verified": 0,
      "Volume": {
        "key": "volume_lv0",
        "value": "0",
        "point": 0,
        "total": 0
      },
      "TotalUniqueOwner": {
        "key": "owner_lv0",
        "value": "0",
        "point": 0,
        "total": 0
      },
      "TotalItems": {
        "key": "item_lv0",
        "value": 0,
        "point": 0,
        "total": 0
      },
      "Followers": {
        "key": "follower_lv0",
        "value": 0,
        "point": 0,
        "total": 0
      }
    }',
    "metadataJson"   JSONB,
    "gameId"         TEXT,
    "source"         TEXT,
    "categoryG"      JSONB,
    "vol"            DOUBLE PRECISION NOT NULL DEFAULT 0,
    "volumeWei"      TEXT             NOT NULL DEFAULT '0',

    CONSTRAINT "Collection_pkey" PRIMARY KEY ("id")
);

ALTER TABLE "Collection"
    ADD COLUMN IF NOT EXISTS "chainId" BIGINT NOT NULL DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS nft_collection_id;

DROP TABLE IF EXISTS "NFT";

-- +goose StatementEnd
