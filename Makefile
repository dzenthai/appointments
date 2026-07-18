.PHONY: api/run
api/run:
	go run ./cmd/api

.PHONY: migrations/create
migrations/create:
	migrate create -seq -ext .sql -dir ./migrations ${name}

MIGRATION := migrate -path ./migrations -database ${DB_DSN}

check-dsn:
ifndef DB_DSN
	$(error DB_DSN is not set; run 'direnv allow')
endif

.PHONY: migrations/up
migrations/up: check-dsn
	$(MIGRATION) up

.PHONY: migrations/down
migrations/down: check-dsn
	$(MIGRATION) down

.PHONY: docker/run/db
docker/run/db:
	docker-compose up -d --build appointments-db