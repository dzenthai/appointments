.PHONY: api/run
api/run:
	go run ./cmd/api

.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

MIGRATION := migrate -path ./migrations -database ${DB_DSN}

check-dsn:
ifndef DB_DSN
	$(error DB_DSN is not set; run 'direnv allow')
endif

.PHONY: migrations/create
migrations/create:
	migrate create -seq -ext .sql -dir ./migrations ${name}

.PHONY: migrations/up
migrations/up: check-dsn
	$(MIGRATION) up

.PHONY: migrations/down
migrations/down: check-dsn
	$(MIGRATION) down

.PHONY: docker/run/db
docker/run/db:
	docker compose up -d appointments-db