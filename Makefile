.PHONY: dev setup build-api seed test vet dashboard-typecheck dashboard-build check clean

GO_PACKAGES := ./cmd/... ./internal/... ./pkg/...
GOCACHE ?= $(CURDIR)/.gocache
export GOCACHE

dev:
	docker-compose up --build -d

setup:
	docker compose up --build -d postgres redis minio migrate core-api admin-api admin-dashboard
	docker compose --profile tools run --rm --build seed
	docker compose --profile tools run --rm --build tenant-migrate
	docker compose up -d --force-recreate core-api admin-api admin-dashboard

build-api:
	go build -o bin/api cmd/api/main.go

seed:
	docker compose --profile tools run --rm --build seed
	docker compose --profile tools run --rm --build tenant-migrate

test:
	go test $(GO_PACKAGES)

vet:
	go vet $(GO_PACKAGES)

dashboard-typecheck:
	cd admin-dashboard && npm run typecheck

dashboard-build:
	cd admin-dashboard && npm run build

check: test vet dashboard-typecheck dashboard-build

clean:
	rm -rf bin/
