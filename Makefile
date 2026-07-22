.PHONY: run/api
run/api: check-dsn
	go run ./cmd/api

.PHONY: tidy
tidy:
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Tidying module dependencies...'
	go mod tidy -v

.PHONY: audit
audit:
	@echo 'Checking module dependencies...'
	go mod tidy -diff
	go mod verify
	@echo 'Checking formatting...'
	test -z "$(shell gofmt -l .)"
	@echo 'Vetting code...'
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...
	@echo 'Running tests...'
	go test -race -vet=off -count=1 ./...
	@echo 'Building application...'
	go build ./...

MIGRATION := migrate -path ./migrations -database ${DB_DSN}

.PHONY: check-dsn
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

COMPOSE_UP := docker compose up -d

.PHONY: docker/compose
docker/compose:
	$(COMPOSE_UP)

.PHONY: docker/compose/db
docker/compose/db:
	$(COMPOSE_UP) appointments-db

.PHONY: docker/compose/migrate
docker/compose/db:
	$(COMPOSE_UP) appointments-migrate

.PHONY: docker/compose/api
docker/compose/db:
	$(COMPOSE_UP) appointments-api