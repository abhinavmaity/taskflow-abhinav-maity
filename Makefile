.PHONY: backend-dev backend-test backend-migrate backend-seed frontend-dev frontend-build

backend-dev:
	cd backend && go run ./cmd/api

backend-test:
	cd backend && go test ./...

backend-migrate:
	cd backend && go run ./cmd/migrate

backend-seed:
	cd backend && go run ./cmd/seed

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build
