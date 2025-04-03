-- name: UpsertNFT :one
INSERT INTO "NFT" ("id", name, "createdAt", "updatedAt", status, "tokenUri", "txCreationHash",
                   "creatorId", "collectionId", image, description, "animationUrl",
                   "nameSlug", source, "ownerId", "slug")
VALUES ($1, $2, $3, $4, $5, $6, $7,
        $8, $9, $10, $11, $12, $13,
        $14, $15, $16)
ON CONFLICT ("id", "collectionId")
    DO UPDATE SET name             = EXCLUDED.name,
                  "updatedAt"      = CURRENT_TIMESTAMP,
                  status           = EXCLUDED.status,
                  "tokenUri"       = EXCLUDED."tokenUri",
                  image            = EXCLUDED.image,
                  description      = EXCLUDED.description,
                  "animationUrl"   = EXCLUDED."animationUrl",
                  "ownerId"        = EXCLUDED."ownerId"
RETURNING *;
