# Environment Variables
GOOSE_MIGRATION_DIR=./sql/schema
GOOSE_DRIVER=postgres
GOOSE_DBSTRING=postgres://postgres:postgres@localhost:5432/gator

# Export environment variables
export GOOSE_MIGRATION_DIR
export GOOSE_DRIVER
export GOOSE_DBSTRING

# Commands
.PHONY: up down gen

up:
	GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) \
	GOOSE_DRIVER=$(GOOSE_DRIVER) \
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) \
	goose up

down:
	GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) \
	GOOSE_DRIVER=$(GOOSE_DRIVER) \
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) \
	goose down

gen:
	sqlc generate
