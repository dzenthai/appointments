# Make sure that direnv is allowed
.PHONY: api/run
api/run:
	go run ./cmd/api

.PHONY: migrations/create
migrations/create:
	migrate create -seq -ext .sql -dir ./migrations ${name}

.PHONY: migrations/up
migrations/up:
	migrate -path ./migrations -database ${DSN} up