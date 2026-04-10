FROM golang:1.22-alpine AS builder
WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -o /out/migrate ./cmd/migrate && \
    CGO_ENABLED=0 GOOS=linux go build -o /out/seed ./cmd/seed

FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates postgresql-client && update-ca-certificates

COPY --from=builder /out/api /app/api
COPY --from=builder /out/migrate /app/migrate
COPY --from=builder /out/seed /app/seed
COPY backend/migrations /app/migrations
COPY backend/seed /app/seed-data
COPY backend/scripts/start.sh /app/start.sh

RUN chmod +x /app/api /app/migrate /app/seed /app/start.sh

ENV MIGRATIONS_DIR=/app/migrations
ENV SEED_DIR=/app/seed-data

EXPOSE 8080

ENTRYPOINT ["/app/start.sh"]
