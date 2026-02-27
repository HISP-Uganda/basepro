.PHONY: backend-run backend-test desktop-dev desktop-build desktop-test deps migrate-up migrate-down migrate-create

backend-run:
	cd backend && GOCACHE=/tmp/go-build go run ./cmd/api

backend-test:
	cd backend && GOCACHE=/tmp/go-build go test ./...

desktop-dev:
	cd desktop && GOROOT=/usr/local/go PATH=/usr/local/go/bin:$$PATH wails dev -compiler /usr/local/go/bin/go

desktop-build:
	cd desktop && GOROOT=/usr/local/go PATH=/usr/local/go/bin:$$PATH wails build -compiler /usr/local/go/bin/go -skipbindings

desktop-test:
	cd desktop/frontend && npm test

migrate-up:
	cd backend && GOCACHE=/tmp/go-build go run ./cmd/migrate up

migrate-down:
	cd backend && GOCACHE=/tmp/go-build go run ./cmd/migrate down

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "usage: make migrate-create name=<migration_name>"; \
		exit 1; \
	fi
	cd backend && GOCACHE=/tmp/go-build go run ./cmd/migrate create -name $(name)

deps:
	cd backend && GOCACHE=/tmp/go-build go mod tidy
	cd desktop/frontend && npm install
