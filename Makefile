.PHONY: api-test api-run migrate-up web-install web-build test

api-test:
	cd apps/api && go test ./...

api-run:
	cd apps/api && go run ./cmd/server

migrate-up:
	cd apps/api && go run ./cmd/migrate up

web-install:
	cd apps/web && npm install

web-build:
	cd apps/web && npm run build

test: api-test web-build
