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
    "ownerId"        TEXT      NOT NULL,
    UNIQUE ("id", "collectionId")
);

--CREATE INDEX IF NOT EXISTS nft_collection_id ON "NFT" ("collectionId");

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
    "status"         TEXT             NOT NULL,
    "type"           TEXT             NOT NULL,
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
    "gameLayergId"   TEXT,
    "source"         TEXT,
    "categoryG"      JSONB,
    "vol"            DOUBLE PRECISION NOT NULL DEFAULT 0,
    "volumeWei"      TEXT             NOT NULL DEFAULT '0',

    CONSTRAINT "Collection_pkey" PRIMARY KEY ("id")
);

CREATE TABLE IF NOT EXISTS "AnalysisCollection"
(
    "id"           TEXT             NOT NULL PRIMARY KEY,
    "collectionId" UUID             NOT NULL REFERENCES "Collection" ON UPDATE CASCADE ON DELETE RESTRICT,
    "keyTime"      TEXT             NOT NULL,
    address        TEXT             NOT NULL,
    type           TEXT             NOT NULL,
    volume         NUMERIC(78)      NOT NULL DEFAULT 0,
    vol            DOUBLE PRECISION NOT NULL DEFAULT 0,
    "volumeWei"    TEXT             NOT NULL DEFAULT '0',
    "floorPrice"   BIGINT           NOT NULL DEFAULT 0,
    floor          DOUBLE PRECISION NOT NULL DEFAULT 0,
    "floorWei"     TEXT             NOT NULL DEFAULT '0',
    items          BIGINT           NOT NULL DEFAULT 0,
    owner          BIGINT           NOT NULL DEFAULT 0,
    "createdAt"    TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE "Collection"
    ADD COLUMN IF NOT EXISTS "chainId" BIGINT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS "Order"
(
    "index"            INTEGER          NOT NULL DEFAULT 1,
    "sig"              TEXT             NOT NULL,
    "makerId"          UUID             NOT NULL REFERENCES "User" ON UPDATE CASCADE ON DELETE RESTRICT,
    "makeAssetType"    INTEGER          NOT NULL,
    "makeAssetAddress" TEXT             NOT NULL,
    "makeAssetValue"   TEXT             NOT NULL,
    "makeAssetId"      TEXT             NOT NULL,
    "takerId"          UUID             REFERENCES "User" ON UPDATE CASCADE ON DELETE SET NULL,
    "takeAssetType"    INTEGER          NOT NULL,
    "takeAssetAddress" TEXT             NOT NULL,
    "takeAssetValue"   TEXT             NOT NULL,
    "takeAssetId"      TEXT             NOT NULL,
    "salt"             TEXT             NOT NULL,
    "start"            INTEGER          NOT NULL DEFAULT 0,
    "end"              INTEGER          NOT NULL DEFAULT 0,
    "orderStatus"      TEXT             NOT NULL DEFAULT 'OPEN'::"ORDERSTATUS",
    "orderType"        TEXT             NOT NULL,
    "root"             TEXT             NOT NULL DEFAULT '0x0000000000000000000000000000000000000000000000000000000000000000',
    "proof"            TEXT[]                    DEFAULT ARRAY []::TEXT[],
    "tokenId"          VARCHAR(255)     NOT NULL,
    "collectionId"     UUID             NOT NULL,
    "quantity"         INTEGER          NOT NULL DEFAULT 1,
    "price"            TEXT             NOT NULL DEFAULT '0',
    "priceNum"         DOUBLE PRECISION NOT NULL DEFAULT 0,
    "netPrice"         TEXT             NOT NULL DEFAULT '0',
    "netPriceNum"      DOUBLE PRECISION NOT NULL DEFAULT 0,
    "createdAt"        TIMESTAMP(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt"        TIMESTAMP(3),
    "quoteToken"       TEXT             NOT NULL,
    "filledQty"        INTEGER          NOT NULL DEFAULT 0,
    PRIMARY KEY ("sig", "index"),
    CONSTRAINT "order_by_id_fk"
        FOREIGN KEY ("tokenId", "collectionId") REFERENCES "NFT"
            ON UPDATE CASCADE ON DELETE RESTRICT
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
--DROP INDEX IF EXISTS nft_collection_id;

--DROP TABLE IF EXISTS "NFT";

-- +goose StatementEnd
