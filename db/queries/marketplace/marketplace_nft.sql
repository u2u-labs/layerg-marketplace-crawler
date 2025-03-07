-- name: UpsertNFT :one
INSERT INTO nft (uid, "tokenId", name, "createdAt", "updatedAt", status, "tokenUri", "txCreationHash",
                 "creatorId", "collectionId", "chainId", image, description, "animationUrl",
                 "nameSlug", "metricPoint", "metricDetail", source)
VALUES (COALESCE($1, gen_random_uuid()), $2, $3, $4, $5, $6, $7, $8,
        $9, $10, $11, $12, $13, $14,
        $15, $16, $17, $18)
ON CONFLICT ("tokenId", "collectionId", "chainId")
    DO UPDATE SET name             = EXCLUDED.name,
                  "updatedAt"      = CURRENT_TIMESTAMP,
                  status           = EXCLUDED.status,
                  "tokenUri"       = EXCLUDED."tokenUri",
                  "txCreationHash" = EXCLUDED."txCreationHash",
                  "creatorId"      = EXCLUDED."creatorId",
                  image            = EXCLUDED.image,
                  description      = EXCLUDED.description,
                  "animationUrl"   = EXCLUDED."animationUrl",
                  "nameSlug"       = EXCLUDED."nameSlug",
                  "metricPoint"    = EXCLUDED."metricPoint",
                  source           = EXCLUDED.source
RETURNING *;
