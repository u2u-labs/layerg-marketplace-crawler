-- +goose Up
-- +goose StatementBegin
INSERT INTO apps (id, name, secret_key)
VALUES ('f3e3bf76-62dc-42a7-ad0d-ef9033bc13a5', '', 'default');


-- INSERT INTO chains (id, chain, name, rpc_url, chain_id, explorer, latest_block, block_time)
-- VALUES (2, 'U2U', 'Solaris Mainnet', 'https://rpc-mainnet.uniultra.xyz', 39, 'https://u2uscan.xyz/', 20233972, 2000);

-- +goose StatementEnd

-- +goose Down
DELETE FROM apps WHERE id = 'f3e3bf76-62dc-42a7-ad0d-ef9033bc13a5';
-- +goose StatementBegin
-- +goose StatementEnd
