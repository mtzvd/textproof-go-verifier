# TextProof

## Система доказательства авторства текстов с использованием блокчейн-технологии

TextProof — это веб-приложение для фиксации авторства текстовых документов в блокчейне. Система использует криптографические хеши и Proof-of-Work для создания неизменяемой записи о существовании текста в определённый момент времени.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**Продакшн:** [textproof.ru](https://textproof.ru)

---

## Возможности

- **Депонирование текстов** — зафиксируйте авторство вашего текста в блокчейне
- **Проверка подлинности** — проверьте текст по ID или полному содержимому
- **Блокчейн с Proof-of-Work** — защита от подделки через майнинг блоков
- **Надёжное хранение** — WAL (Write-Ahead Logging) + автоматические бэкапы
- **QR-коды** — для быстрой проверки на мобильных устройствах
- **Встраиваемые бейджи** — HTML-виджеты для внешних сайтов
- **Быстрый поиск** — O(1) поиск дубликатов через индексацию
- **REST API v1** — полноценный JSON API с Swagger-документацией
- **Структурированное логирование** — `log/slog`
- **Безопасность** — rate limiting, security headers, валидация входных данных

---

## Быстрый старт

### Требования

- [Go](https://golang.org/dl/) 1.25 или новее
- [Templ](https://templ.guide/) для генерации шаблонов

### Установка

```bash
# Клонируйте репозиторий
git clone https://github.com/mtzvd/textproof-go-verifier.git
cd textproof-go-verifier

# Установите зависимости
go mod download

# Установите templ (если ещё не установлен)
go install github.com/a-h/templ/cmd/templ@latest

# Сгенерируйте шаблоны
templ generate

# Запустите сервер
go run cmd/server/main.go
```

Приложение будет доступно по адресу: **<http://localhost:8080>**

---

## Использование

### Депонирование текста

1. Перейдите на `/deposit`
2. Заполните форму: имя автора, название произведения, полный текст
3. Нажмите "Зафиксировать в блокчейне"
4. Получите уникальный ID, QR-код и встраиваемый бейдж

### Проверка текста

- **По ID:** `/verify` -> вкладка "По идентификатору" -> введите ID блока
- **По тексту:** `/verify` -> вкладка "По тексту" -> вставьте полный текст
- **Прямая ссылка:** `/verify/{id}` — автоматическая проверка

---

## Архитектура

### Структура проекта

```text
textproof-go-verifier/
├── cmd/server/                  # Точка входа
│   └── main.go
├── internal/
│   ├── api/                     # HTTP handlers, маршруты, middleware
│   │   ├── api.go               # Роутер и маршруты
│   │   ├── middleware.go        # Security headers, rate limiting, body size
│   │   ├── handlers_pages.go    # Страницы (главная, about, privacy, terms)
│   │   ├── handlers_deposit.go  # Депонирование
│   │   ├── handlers_verify.go   # Проверка
│   │   ├── handlers_api.go      # JSON API v1
│   │   ├── handlers_docs.go     # Swagger UI
│   │   ├── helpers.go           # Утилиты рендеринга
│   │   └── flash.go             # Flash messages (cookies)
│   ├── blockchain/              # Логика блокчейна
│   │   ├── block.go             # Структура блока
│   │   ├── blockchain.go        # Основная логика цепи
│   │   ├── storage.go           # Хранение (JSON + WAL + бэкапы)
│   │   ├── errors.go            # Типы ошибок
│   │   └── id_generator.go      # Генерация ID блоков
│   ├── config/                  # Конфигурация и константы
│   └── viewmodels/              # Модели данных для UI
├── web/
│   ├── embed.go                 # embed.FS для статических файлов
│   ├── static/                  # CSS, шрифты, иконки (self-hosted)
│   │   ├── css/
│   │   │   ├── bulma.min.css    # Bulma v1.0.0
│   │   │   ├── fontawesome.min.css
│   │   │   └── styles.css       # Кастомные стили
│   │   └── webfonts/            # Font Awesome woff2
│   └── templates/               # Templ шаблоны
│       ├── base.templ           # Базовый layout
│       ├── components/          # Переиспользуемые компоненты
│       └── ...                  # Страницы
├── docs/                        # Swagger (сгенерированные)
├── scripts/                     # Вспомогательные скрипты
├── data/                        # Данные блокчейна (не в git)
├── Taskfile.yml                 # Task runner
├── modd.conf                    # Hot reload
├── go.mod
└── go.sum
```

### Блокчейн

**Структура блока:**

```go
type Block struct {
    ID        string       // "000-000-001"
    PrevHash  string       // Хеш предыдущего блока
    Timestamp time.Time    // Время создания
    Data      DepositData  // Данные о тексте
    Nonce     int          // Proof-of-Work nonce
    Hash      string       // SHA-256 хеш блока
}

type DepositData struct {
    AuthorName  string  // Имя автора
    Title       string  // Название
    TextStart   string  // Первые 3 слова
    TextEnd     string  // Последние 3 слова
    ContentHash string  // SHA-256 хеш полного текста
    PublicKey   string  // (Опционально) Публичный ключ
}
```

**Proof-of-Work:**

- Конфигурируемая сложность (по умолчанию: 4 нуля)
- Майнинг блока занимает несколько секунд
- Защита от подделки прошлых записей

**Хранение:**

- JSON файлы для простоты
- WAL для защиты от сбоев
- Автоматические бэкапы (хранятся последние 5)
- Atomic write через временные файлы

---

## API

### Web UI (внутренние маршруты)

| Метод | Путь | Описание |
| ----- | ---- | -------- |
| GET | `/` | Главная страница |
| GET | `/deposit` | Форма депонирования |
| POST | `/api/deposit` | Обработка депонирования |
| GET | `/deposit/result/{id}` | Результат депонирования |
| GET | `/verify` | Форма проверки |
| POST | `/api/verify/id` | Проверка по ID (форма) |
| POST | `/api/verify/text` | Проверка по тексту (форма) |
| GET | `/verify/{id}` | Прямая ссылка на проверку |
| GET | `/verify/result/{id}` | Результат проверки |
| GET | `/api/qrcode/{id}` | Генерация QR-кода |
| GET | `/api/badge/{id}` | HTML-бейдж для встраивания |
| GET | `/docs` | Swagger UI |

### Public JSON API v1

| Метод | Путь | Описание |
| ----- | ---- | -------- |
| POST | `/api/v1/deposit` | Депонирование текста |
| POST | `/api/v1/verify/id` | Проверка по ID |
| POST | `/api/v1/verify/text` | Проверка по тексту |
| GET | `/api/v1/stats` | Статистика блокчейна |
| GET | `/api/v1/blockchain` | Информация о блокчейне |
| GET | `/api/v1/blockchain/export` | Экспорт всего блокчейна (JSON) |

Полная документация API: [textproof.ru/docs](https://textproof.ru/docs)

---

## Конфигурация

```bash
go run cmd/server/main.go [опции]

Опции:
  -data-dir string    Директория для хранения данных (default "data")
  -port int           Порт для HTTP сервера (default 8080)
  -difficulty int     Сложность майнинга — количество нулей (default 4)
  -debug              Включить режим отладки
```

---

## Разработка

### Hot Reload с modd

```bash
go install github.com/cortesi/modd/cmd/modd@latest
modd
```

При изменении `.templ` файлов автоматически запустится `templ generate` и сервер перезапустится.

### Тестирование

```bash
go test ./...
go test -cover ./...
```

### Сборка для production

```bash
GOOS=linux GOARCH=amd64 go build -o build/textproof ./cmd/server/
```

---

## Безопасность

- Rate limiting (per-IP) для API-эндпоинтов
- Security headers: X-Content-Type-Options, X-Frame-Options, Referrer-Policy
- HttpOnly cookies для flash messages
- Валидация входных данных и ограничение размера body
- Proof-of-Work как защита от спама
- Content hash index — предотвращение дубликатов
- Atomic writes — защита от повреждения данных
- HTTPS через Caddy reverse proxy

---

## Технологии

- [Go](https://golang.org/) — язык программирования
- [Templ](https://templ.guide/) — type-safe шаблоны
- [Gorilla Mux](https://github.com/gorilla/mux) — HTTP роутер
- [Bulma](https://bulma.io/) v1.0.0 — CSS фреймворк (self-hosted)
- [Font Awesome](https://fontawesome.com/) 6.5.0 — иконки (self-hosted)
- [Alpine.js](https://alpinejs.dev/) — легковесный JS фреймворк
- [Swagger](https://swagger.io/) — документация API
- [go-qrcode](https://github.com/skip2/go-qrcode) — генерация QR-кодов

---

## Лицензия

MIT. См. файл [LICENSE](LICENSE).

---

## Автор

**Георгий Агафонов** — [@mtzvd](https://github.com/mtzvd)
