-- name: UpsertAnalysisCollection :exec
INSERT INTO "AnalysisCollection" (id, "collectionId", "keyTime", address, type, volume, vol, "volumeWei", "floorPrice",
                                  floor, "floorWei", items, owner, "createdAt")
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
ON CONFLICT ("id")
    DO UPDATE
    SET volume       = EXCLUDED.volume,
        vol          = EXCLUDED.vol,
        "volumeWei"  = EXCLUDED."volumeWei",
        "floorPrice" = EXCLUDED."floorPrice",
        floor        = EXCLUDED.floor,
        "floorWei"   = EXCLUDED."floorWei",
        items        = EXCLUDED.items,
        owner        = EXCLUDED.owner;
