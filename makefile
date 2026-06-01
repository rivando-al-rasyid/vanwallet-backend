include ./.env

MIGRATION_PATH=database/migrations
SEED_FILE=database/seed.sql

# Connection string used INSIDE the Docker network
DATABASE_URL_DOCKER=postgresql://$(DB_USER):$(DB_PASS)@postgres:5432/$(DB_NAME)?sslmode=disable

# Create a new migration file (runs on your host)
migrate-create:
	@migrate create -ext sql -dir $(MIGRATION_PATH) -seq create_$(NAME)_table

# Apply all pending migrations
migrate-up:
	docker compose run --rm \
		-v $$(pwd)/$(MIGRATION_PATH):/migrations \
		migrate/migrate \
		-database "$(DATABASE_URL_DOCKER)" \
		-path /migrations up

# Roll back the last migration
migrate-down:
	docker compose run --rm \
		-v $$(pwd)/$(MIGRATION_PATH):/migrations \
		migrate/migrate \
		-database "$(DATABASE_URL_DOCKER)" \
		-path /migrations down

# Force a specific version
migrate-force:
	docker compose run --rm \
		-v $$(pwd)/$(MIGRATION_PATH):/migrations \
		migrate/migrate \
		-database "$(DATABASE_URL_DOCKER)" \
		-path /migrations force $(VERSION)

# Seed the database (requires a running postgres container)
seed:
	docker compose exec -T postgres \
		psql -U $(DB_USER) -d $(DB_NAME) -f - < $(SEED_FILE)

# Truncate all tables and re-seed
seed-reset:
	docker compose exec -T postgres \
		psql -U $(DB_USER) -d $(DB_NAME) -c \
		"TRUNCATE TABLE withdrawals, expenses, transfers, transactions, topups, wallets, user_pins, favorites, profiles, users RESTART IDENTITY CASCADE;"
	@$(MAKE) seed

# Print the database URL used inside Docker
print-db-url:
	@echo $(DATABASE_URL_DOCKER)