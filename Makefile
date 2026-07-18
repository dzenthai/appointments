# Make sure that direnv is allowed
.PHONY: api/run
api/run:
	go run ./cmd/api

.PHONY: migrations/create
migrations/create:
	migrate create -seq -ext .sql -dir ./migrations ${name}

MIGRATION := migrate -path ./migrations -database ${DB_DSN}

.PHONY: migrations/up
migrations/up:
	$(MIGRATION) up

.PHONY: migrations/down
migrations/down:
	$(MIGRATION) down

.PHONY: docker/run/db
docker/run/db:
	docker-compose up -d --build appointments-db