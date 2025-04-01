# Load environment variables from .env file
-include .env
export

GOOSE_DRIVER?='postgres'
GOOSE_DBSTRING?='postgres://root:admin@localhost:26257/marketplace_crawler?sslmode=disable'
GOOSE_MIGRATION_DIR?='./db/migrations'
GOOSE_DEFAULT_DBSTRING?=postgres://root:admin@localhost:26257/defaultdb?sslmode=disable

GOOSE_MKP_DEFAULT_DBSTRING?=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable
GOOSE_MKP_DBSTRING?='postgres://postgres:postgres@localhost:5432/marketplace?sslmode=disable'
GOOSE_MKP_MIGRATION_DIR?='./db/migrations/marketplace'

build:
	go build -ldflags -w
all:
	go build -ldflags -w
	chmod +x layerg-crawler
	./layerg-crawler --config .layerg-crawler.yaml

api:
	go build -ldflags -w
	chmod +x layerg-crawler
	./layerg-crawler api --config .layerg-crawler.yaml

worker:
	go build -ldflags -w
	chmod +x layerg-crawler
	./layerg-crawler worker --config .layerg-crawler.yaml

#crawler db
create-db:
	@echo "Creating database marketplace_crawler..."
	@psql "$(GOOSE_DEFAULT_DBSTRING)" -c "SELECT 1 FROM pg_database WHERE datname = 'marketplace_crawler'" | grep -q 1 || psql "$(GOOSE_DEFAULT_DBSTRING)" -c "CREATE DATABASE marketplace_crawler"

create-db-marketplace:
	@echo "Creating database marketplace..."
	@psql "$(GOOSE_MKP_DEFAULT_DBSTRING)" -c "SELECT 1 FROM pg_database WHERE datname = 'marketplace'" | grep -q 1 || psql "$(GOOSE_MKP_DEFAULT_DBSTRING)" -c "CREATE DATABASE marketplace"

migrate-up: create-db migrate-up-marketplace
	@echo "migrate up crawler $(GOOSE_DBSTRING) $(GOOSE_MIGRATION_DIR)"
	@GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) goose -dir $(GOOSE_MIGRATION_DIR) up-to 20250319042110
migrate-down:
	@GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) goose -dir $(GOOSE_MIGRATION_DIR) down
migrate-reset:
	@GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_DBSTRING) goose -dir $(GOOSE_MIGRATION_DIR) reset	

migrate-up-marketplace: create-db-marketplace
	@echo "Checking and upserting default record in goose_db_version..."
	@echo "migrate up marketplace $(GOOSE_MKP_DBSTRING) $(GOOSE_MKP_MIGRATION_DIR)"
	@psql $(GOOSE_MKP_DBSTRING) -c "INSERT INTO \"goose_db_version\" (id, version_id, is_applied, tstamp) VALUES (0, 0, true, now()) ON CONFLICT (id) DO NOTHING;" || true
	@GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_MKP_DBSTRING) goose -dir $(GOOSE_MKP_MIGRATION_DIR) up-to 20250305083624
migrate-down-marketplace:
	@GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_MKP_DBSTRING) goose -dir $(GOOSE_MKP_MIGRATION_DIR) down
migrate-reset-marketplace:
	@GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING=$(GOOSE_MKP_DBSTRING) goose -dir $(GOOSE_MKP_MIGRATION_DIR) reset

swag:
	swag init -g cmd/api_cmd.go -o ./docs
