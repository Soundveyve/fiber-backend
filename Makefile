.PHONY: help run build test clean migrate-up migrate-down migrate-create sqlc docker-up docker-down

# Цвета для вывода
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)

# Переменные
APP_NAME=fiber-backend
MIGRATION_DIR=migrations
DB_URL=postgresql://postgres:postgres@localhost:5432/fiber_db?sslmode=disable

## help: Показать справку по командам
help:
	@echo ''
	@echo 'Использование:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Доступные команды:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${WHITE}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

## Разработка:

## run: Запустить приложение
run:
	@echo "${GREEN}Запуск приложения...${RESET}"
	go run cmd/api/main.go

## build: Собрать бинарный файл
build:
	@echo "${GREEN}Сборка приложения...${RESET}"
	go build -o bin/$(APP_NAME) cmd/api/main.go

## test: Запустить тесты
test:
	@echo "${GREEN}Запуск тестов...${RESET}"
	go test -v ./...

## clean: Удалить собранные файлы
clean:
	@echo "${GREEN}Очистка...${RESET}"
	rm -rf bin/

## База данных:

## migrate-up: Применить все миграции
migrate-up:
	@echo "${GREEN}Применение миграций...${RESET}"
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" -verbose up

## migrate-down: Откатить последнюю миграцию
migrate-down:
	@echo "${YELLOW}Откат миграции...${RESET}"
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" -verbose down 1

## migrate-force: Принудительно установить версию миграции (использовать: make migrate-force VERSION=1)
migrate-force:
	@echo "${YELLOW}Принудительная установка версии миграции...${RESET}"
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" force $(VERSION)

## migrate-create: Создать новую миграцию (использовать: make migrate-create NAME=название_миграции)
migrate-create:
	@echo "${GREEN}Создание миграции $(NAME)...${RESET}"
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME)

## sqlc: Сгенерировать код из SQL запросов
sqlc:
	@echo "${GREEN}Генерация кода sqlc...${RESET}"
	sqlc generate

## Docker:

## docker-up: Запустить Docker контейнеры
docker-up:
	@echo "${GREEN}Запуск Docker контейнеров...${RESET}"
	docker-compose up -d

## docker-down: Остановить Docker контейнеры
docker-down:
	@echo "${YELLOW}Остановка Docker контейнеров...${RESET}"
	docker-compose down

## docker-logs: Показать логи контейнеров
docker-logs:
	docker-compose logs -f

## docker-build: Собрать Docker образ приложения
docker-build:
	@echo "${GREEN}Сборка Docker образа...${RESET}"
	docker build -t $(APP_NAME):latest .

## Установка зависимостей:

## deps: Установить все зависимости
deps:
	@echo "${GREEN}Установка зависимостей...${RESET}"
	go mod download
	go mod tidy

## install-tools: Установить необходимые инструменты
install-tools:
	@echo "${GREEN}Установка инструментов...${RESET}"
	@echo "Установка golang-migrate..."
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Установка sqlc..."
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@echo "${GREEN}✅ Все инструменты установлены${RESET}"

## Полезные команды:

## dev: Запустить БД и применить миграции
dev: docker-up
	@echo "${YELLOW}Ожидание запуска PostgreSQL...${RESET}"
	@sleep 3
	@$(MAKE) migrate-up
	@echo "${GREEN}✅ Окружение разработки готово${RESET}"

## reset-db: Полностью пересоздать БД
reset-db:
	@echo "${YELLOW}Пересоздание БД...${RESET}"
	@$(MAKE) docker-down
	docker volume rm fiber-backend_postgres_data 2>/dev/null || true
	@$(MAKE) docker-up
	@sleep 3
	@$(MAKE) migrate-up
	@echo "${GREEN}✅ БД пересоздана${RESET}"
