-include .env
export

export PROJECT_ROOT=${shell pwd}

env-up:
	@docker compose up -d sportapp-postgres

env-down:
	@docker compose down sportapp-postgres

env-cleanup:
	@read -p "Очистить все volume файлы окружения? Опасность утери данных. [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down sportapp-postgres && \
		rm -rf out/pgdata && \
		echo "Файлы окружения очищены"; \
	else \
		echo "Очистка окружения отменена"; \
	fi

env-port-forward:
	@docker compose up -d port-forwarder

env-port-close:
	@docker compose down port-forwarder

migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "Отсутствует необходимый параметр seq. Пример: make migrate-create seq=init"; \
		exit 1; \
	fi; \
	docker compose run --rm sportapp-postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Отсутствует необходимый параметр action. Пример: make migrate-action action=up"; \
		exit 1; \
	fi; \
	docker compose run --rm sportapp-postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@sportapp-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		"$(action)"

sportapp-run:
	@go run cmd/sport_app/main.go

sportapp-deploy:
	@docker compose up -d --build backend

ps:
	@docker compose ps

# Development and Quality Assurance targets
test:
	@go test -v ./...

test-cover:
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

fmt:
	@gofmt -s -w .

fmt-check:
	@files=$$(gofmt -l .); if [ -n "$$files" ]; then echo "Unformatted files found:"; echo "$$files"; exit 1; else echo "All files properly formatted."; fi

vet:
	@go vet ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed locally, running go vet"; \
		go vet ./...; \
	fi

check: fmt-check vet test

build:
	@go build -v -o out/sport_app cmd/sport_app/main.go