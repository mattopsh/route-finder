.PHONY: build test vet run check-deps

build:
	go build -o bin/api ./cmd/api

test:
	go test ./...

vet:
	go vet ./...

run:
	OSRM_BASE_URL=https://router.project-osrm.org OSRM_TIMEOUT_MS=3000 OSRM_MAX_CONCURRNCY=5 APP_PORT=8080 go run ./cmd/api