.PHONY: all build run dev clean test swagger templ help install-tools fmt lint check docker-build docker-run backup setup

# =====================================================================
# ПЕРЕМЕННЫЕ
# =====================================================================

APP_NAME := textproof
BUILD_DIR := build
DATA_DIR := data
PORT := 8080
DIFFICULTY := 4
MAIN_FILE := cmd/server/main.go

# Цвета для вывода
CYAN := \033[0;36m
GREEN := \033[0;32m
RED := \033[0;31m
YELLOW := \033[0;33m
NC := \033[0m # No Color

# =====================================================================
# ОСНОВНЫЕ КОМАНДЫ
# =====================================================================

.DEFAULT_GOAL := help

help: ## Показать эту справку
	@echo "$(CYAN)TextProof - Доступные команды:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

all: clean generate build ## Полная сборка проекта

# =====================================================================
# УСТАНОВКА И НАСТРОЙКА
# =====================================================================

install-tools: ## Установить необходимые инструменты (templ, swag, air)
	@echo "$(CYAN)Установка инструментов...$(NC)"
	@go install github.com/a-h/templ/cmd/templ@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/air-verse/air@latest
	@echo "$(GREEN)✓ Инструменты установлены$(NC)"

deps: ## Установить зависимости проекта
	@echo "$(CYAN)Установка зависимостей...$(NC)"
	@go mod download
	@go mod verify
	@echo "$(GREEN)✓ Зависимости установлены$(NC)"

setup: install-tools deps generate ## Полная настройка проекта
	@echo "$(GREEN)✓ Проект готов к работе!$(NC)"
	@echo "$(CYAN)Запустите 'make run' для старта сервера$(NC)"
	@echo "$(CYAN)Swagger UI: http://localhost:$(PORT)/swagger/index.html$(NC)"

# =====================================================================
# ГЕНЕРАЦИЯ
# =====================================================================

templ: ## Сгенерировать Templ шаблоны
	@echo "$(CYAN)Генерация Templ шаблонов...$(NC)"
	@if ! command -v templ &> /dev/null; then \
		echo "$(RED)Templ не установлен. Запустите: make install-tools$(NC)"; \
		exit 1; \
	fi
	@templ generate
	@echo "$(GREEN)✓ Templ шаблоны сгенерированы$(NC)"

swagger: ## Сгенерировать Swagger документацию
	@echo "$(CYAN)Генерация Swagger документации...$(NC)"
	@if ! command -v swag &> /dev/null; then \
		echo "$(RED)Swag не установлен. Запустите: make install-tools$(NC)"; \
		exit 1; \
	fi
	@swag init -g $(MAIN_FILE) -o docs --parseDependency --parseInternal --exclude web/templates
	@$(MAKE) fix-docs
	@echo "$(GREEN)✓ Swagger документация сгенерирована$(NC)"

fix-docs: ## Исправить известную ошибку в docs/docs.go
	@if [ -f scripts/fixdocs.go ]; then \
		go run scripts/fixdocs.go; \
	fi

generate: templ swagger ## Сгенерировать всё (templ + swagger)
	@echo "$(GREEN)✓ Генерация завершена$(NC)"

# =====================================================================
# СБОРКА И ЗАПУСК
# =====================================================================

build: templ ## Собрать приложение
	@echo "$(CYAN)Сборка приложения...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "$(GREEN)✓ Сборка завершена: $(BUILD_DIR)/$(APP_NAME)$(NC)"

run: generate ## Запустить сервер
	@echo "$(CYAN)Запуск сервера на порту $(PORT)...$(NC)"
	@go run $(MAIN_FILE) -port $(PORT) -difficulty $(DIFFICULTY)

dev: templ ## Запустить в режиме разработки с hot reload (air)
	@echo "$(CYAN)Режим разработки (hot reload)...$(NC)"
	@if command -v air &> /dev/null; then \
		air; \
	else \
		echo "$(RED)Air не установлен. Установите: make install-tools$(NC)"; \
		echo "$(CYAN)Запуск без hot-reload...$(NC)"; \
		$(MAKE) run; \
	fi

dev-modd: ## Запустить в режиме разработки с modd
	@echo "$(CYAN)Режим разработки (modd)...$(NC)"
	@if command -v modd &> /dev/null; then \
		modd; \
	else \
		echo "$(RED)Modd не установлен. Установите: go install github.com/cortesi/modd/cmd/modd@latest$(NC)"; \
		$(MAKE) run; \
	fi

# =====================================================================
# ТЕСТИРОВАНИЕ
# =====================================================================

test: ## Запустить тесты
	@echo "$(CYAN)Запуск тестов...$(NC)"
	@go test -v ./...

test-cover: ## Запустить тесты с покрытием
	@echo "$(CYAN)Запуск тестов с покрытием...$(NC)"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Отчёт о покрытии: coverage.html$(NC)"

test-race: ## Запустить тесты с детектором гонок
	@echo "$(CYAN)Запуск тестов с race detector...$(NC)"
	@go test -race -v ./...

bench: ## Запустить бенчмарки
	@echo "$(CYAN)Запуск бенчмарков...$(NC)"
	@go test -bench=. -benchmem ./...

# =====================================================================
# КАЧЕСТВО КОДА
# =====================================================================

fmt: ## Форматировать код
	@echo "$(CYAN)Форматирование кода...$(NC)"
	@go fmt ./...
	@if command -v templ &> /dev/null; then \
		templ fmt .; \
	fi
	@echo "$(GREEN)✓ Код отформатирован$(NC)"

lint: ## Проверить код линтером
	@echo "$(CYAN)Проверка кода...$(NC)"
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "$(RED)golangci-lint не установлен$(NC)"; \
		echo "Установите: https://golangci-lint.run/usage/install/"; \
	fi

vet: ## Запустить go vet
	@echo "$(CYAN)Запуск go vet...$(NC)"
	@go vet ./...

check: fmt vet test ## Полная проверка (fmt + vet + test)
	@echo "$(GREEN)✓ Все проверки пройдены$(NC)"

# =====================================================================
# ОЧИСТКА
# =====================================================================

clean: ## Очистить собранные файлы
	@echo "$(CYAN)Очистка...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -rf docs
	@rm -f coverage.out coverage.html
	@find . -name "*_templ.go" -type f -delete
	@echo "$(GREEN)✓ Очистка завершена$(NC)"

clean-data: ## Удалить данные блокчейна (ОСТОРОЖНО!)
	@echo "$(RED)ВНИМАНИЕ: Это удалит все данные блокчейна!$(NC)"
	@read -p "Вы уверены? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		rm -rf $(DATA_DIR); \
		echo "$(GREEN)✓ Данные удалены$(NC)"; \
	else \
		echo "$(CYAN)Отменено$(NC)"; \
	fi

# =====================================================================
# DOCKER
# =====================================================================

docker-build: ## Собрать Docker образ
	@echo "$(CYAN)Сборка Docker образа...$(NC)"
	@docker build -t $(APP_NAME):latest .
	@echo "$(GREEN)✓ Docker образ собран$(NC)"

docker-run: ## Запустить в Docker
	@echo "$(CYAN)Запуск в Docker...$(NC)"
	@docker run -p $(PORT):$(PORT) -v $$(pwd)/$(DATA_DIR):/app/$(DATA_DIR) $(APP_NAME):latest

docker-clean: ## Очистить Docker артефакты
	@docker system prune -f
	@echo "$(GREEN)✓ Docker очищен$(NC)"

# =====================================================================
# УТИЛИТЫ
# =====================================================================

backup: ## Создать бэкап данных
	@echo "$(CYAN)Создание бэкапа...$(NC)"
	@mkdir -p backups
	@tar -czf backups/backup-$$(date +%Y%m%d-%H%M%S).tar.gz $(DATA_DIR)
	@echo "$(GREEN)✓ Бэкап создан в backups/$(NC)"

stats: ## Показать статистику проекта
	@echo "$(CYAN)=== Статистика проекта ===$(NC)"
	@echo "Строк кода (Go):"
	@find . -name "*.go" -not -path "./vendor/*" -not -name "*_templ.go" | xargs wc -l | tail -1
	@echo "Строк кода (Templ):"
	@find . -name "*.templ" | xargs wc -l | tail -1
	@echo "Файлов Go:"
	@find . -name "*.go" -not -path "./vendor/*" -not -name "*_templ.go" | wc -l
	@echo "Файлов Templ:"
	@find . -name "*.templ" | wc -l

# =====================================================================
# SWAGGER
# =====================================================================

swagger-serve: swagger run ## Сгенерировать swagger и запустить сервер

# =====================================================================
# PRODUCTION
# =====================================================================

prod-build: generate ## Собрать для production (Linux, Windows, macOS)
	@echo "$(CYAN)Сборка для production...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@echo "$(CYAN)  - Linux (amd64)...$(NC)"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_FILE)
	@echo "$(CYAN)  - Windows (amd64)...$(NC)"
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(MAIN_FILE)
	@echo "$(CYAN)  - macOS (amd64)...$(NC)"
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_FILE)
	@echo "$(CYAN)  - macOS (arm64)...$(NC)"
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(MAIN_FILE)
	@echo "$(GREEN)✓ Production builds готовы в $(BUILD_DIR)/$(NC)"

# =====================================================================
# GIT
# =====================================================================

pre-commit: fmt vet test ## Проверки перед коммитом
	@echo "$(GREEN)✓ Готово к коммиту$(NC)"