.PHONY: all build run clean test templ help install-templ dev

# Переменные
APP_NAME=textproof
BUILD_DIR=build
DATA_DIR=data
PORT=8080
DIFFICULTY=4

# Цвета для вывода
CYAN=\033[0;36m
GREEN=\033[0;32m
RED=\033[0;31m
NC=\033[0m # No Color

all: clean templ build ## Полная сборка проекта

help: ## Показать эту справку
	@echo "$(CYAN)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'

install-deps: ## Установить зависимости
	@echo "$(CYAN)Установка зависимостей...$(NC)"
	go mod download
	go mod verify

install-templ: ## Установить Templ
	@echo "$(CYAN)Установка Templ...$(NC)"
	go install github.com/a-h/templ/cmd/templ@latest

templ: ## Генерация темплейтов
	@echo "$(CYAN)Генерация темплейтов...$(NC)"
	@if ! command -v templ &> /dev/null; then \
		echo "$(RED)Templ не установлен. Запустите: make install-templ$(NC)"; \
		exit 1; \
	fi
	templ generate

build: templ ## Собрать проект
	@echo "$(CYAN)Сборка проекта...$(NC)"
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server/main.go
	@echo "$(GREEN)✓ Сборка завершена: $(BUILD_DIR)/$(APP_NAME)$(NC)"

run: build ## Запустить сервер
	@echo "$(CYAN)Запуск сервера на порту $(PORT)...$(NC)"
	./$(BUILD_DIR)/$(APP_NAME) -port $(PORT) -difficulty $(DIFFICULTY)

dev: templ ## Запуск в режиме разработки (с авто-пересборкой темплейтов)
	@echo "$(CYAN)Режим разработки...$(NC)"
	@echo "$(CYAN)Запуск templ в watch режиме...$(NC)"
	@(templ generate --watch &)
	@sleep 2
	@echo "$(CYAN)Запуск air для hot-reload...$(NC)"
	@if command -v air &> /dev/null; then \
		air; \
	else \
		echo "$(RED)Air не установлен. Установите: go install github.com/cosmtrek/air@latest$(NC)"; \
		echo "$(CYAN)Запуск без hot-reload...$(NC)"; \
		make run; \
	fi

test: ## Запустить тесты
	@echo "$(CYAN)Запуск тестов...$(NC)"
	go test -v ./...

test-cover: ## Тесты с покрытием
	@echo "$(CYAN)Запуск тестов с покрытием...$(NC)"
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Отчёт о покрытии: coverage.html$(NC)"

clean: ## Очистить собранные файлы
	@echo "$(CYAN)Очистка...$(NC)"
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	find . -name "*_templ.go" -type f -delete
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

fmt: ## Форматировать код
	@echo "$(CYAN)Форматирование кода...$(NC)"
	go fmt ./...
	@if command -v templ &> /dev/null; then \
		templ fmt .; \
	fi
	@echo "$(GREEN)✓ Форматирование завершено$(NC)"

lint: ## Проверить код линтером
	@echo "$(CYAN)Проверка кода...$(NC)"
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "$(RED)golangci-lint не установлен$(NC)"; \
		echo "Установите: https://golangci-lint.run/usage/install/"; \
	fi

install-tools: install-templ ## Установить все инструменты разработки
	@echo "$(CYAN)Установка инструментов разработки...$(NC)"
	go install github.com/cosmtrek/air@latest
	@echo "$(GREEN)✓ Инструменты установлены$(NC)"

docker-build: ## Собрать Docker образ
	@echo "$(CYAN)Сборка Docker образа...$(NC)"
	docker build -t $(APP_NAME):latest .
	@echo "$(GREEN)✓ Docker образ собран$(NC)"

docker-run: ## Запустить в Docker
	@echo "$(CYAN)Запуск в Docker...$(NC)"
	docker run -p $(PORT):$(PORT) -v $(PWD)/$(DATA_DIR):/app/$(DATA_DIR) $(APP_NAME):latest

backup: ## Создать бэкап данных
	@echo "$(CYAN)Создание бэкапа...$(NC)"
	@mkdir -p backups
	tar -czf backups/backup-$$(date +%Y%m%d-%H%M%S).tar.gz $(DATA_DIR)
	@echo "$(GREEN)✓ Бэкап создан в backups/$(NC)"

stats: ## Показать статистику проекта
	@echo "$(CYAN)Статистика проекта:$(NC)"
	@echo "  Всего строк Go кода:"
	@find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1
	@echo "  Всего файлов Templ:"
	@find . -name "*.templ" | wc -l
	@echo "  Размер данных:"
	@if [ -d "$(DATA_DIR)" ]; then du -sh $(DATA_DIR); else echo "  Нет данных"; fi

logs: ## Показать последние логи
	@if [ -f "$(APP_NAME).log" ]; then \
		tail -f $(APP_NAME).log; \
	else \
		echo "$(RED)Файл логов не найден$(NC)"; \
	fi

.DEFAULT_GOAL := help