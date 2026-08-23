.PHONY: build test race vet fmt-check migrate seed web-install web-test web-typecheck web-build verify

build:
	go build ./...

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

fmt-check:
	@test -z "$$(gofmt -l .)"

migrate:
	go run ./cmd/migrate up

seed:
	go run ./cmd/migrate seed

web-install:
	cd web && npm ci

web-test:
	cd web && npm run test:unit -- --run

web-typecheck:
	cd web && npm run check:types

web-build:
	cd web && npm run desk:bundle

verify: fmt-check build test race vet web-test web-typecheck web-build
