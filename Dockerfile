FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /api ./cmd/api

FROM alpine:3.24

COPY --from=builder /api /api

EXPOSE 4000

ENTRYPOINT ["/api"]
