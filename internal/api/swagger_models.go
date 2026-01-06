package api

import "time"

// Swagger Models для API документации

// DepositRequest представляет запрос на депонирование текста
type DepositRequest struct {
	AuthorName string `json:"author_name" example:"Иван Иванов"`  // Имя автора
	Title      string `json:"title" example:"Моя статья"`         // Название произведения
	Content    string `json:"content" example:"Полный текст..."`  // Полный текст документа
	PublicKey  string `json:"public_key" example:"-----BEGIN..."` // Публичный ключ (опционально)
}

// DepositResponse представляет ответ после успешного депонирования
type DepositResponse struct {
	Success   bool      `json:"success" example:"true"`                          // Статус успеха
	Message   string    `json:"message" example:"Текст успешно зарегистрирован"` // Сообщение
	BlockID   string    `json:"block_id" example:"000-000-001"`                  // ID блока
	Timestamp time.Time `json:"timestamp" example:"2026-01-05T15:04:05Z"`        // Время регистрации
	Hash      string    `json:"hash" example:"a1b2c3d4..."`                      // Хеш содержимого
	QRCodeURL string    `json:"qr_code_url" example:"/api/qrcode/000-000-001"`   // URL QR-кода
	VerifyURL string    `json:"verify_url" example:"/verify/000-000-001"`        // URL для проверки
}

// VerifyByIDRequest представляет запрос на проверку по ID
type VerifyByIDRequest struct {
	BlockID string `json:"block_id" example:"000-000-001"` // ID блока для проверки
}

// VerifyByTextRequest представляет запрос на проверку по тексту
type VerifyByTextRequest struct {
	Content string `json:"content" example:"Полный текст документа..."` // Текст для проверки
}

// VerifyResponse представляет результат проверки
type VerifyResponse struct {
	Found     bool      `json:"found" example:"true"`                               // Найден ли текст
	BlockID   string    `json:"block_id,omitempty" example:"000-000-001"`           // ID блока
	Author    string    `json:"author,omitempty" example:"Иван Иванов"`             // Автор
	Title     string    `json:"title,omitempty" example:"Моя статья"`               // Название
	Timestamp time.Time `json:"timestamp,omitempty" example:"2026-01-05T15:04:05Z"` // Время регистрации
	Hash      string    `json:"hash,omitempty" example:"a1b2c3d4..."`               // Хеш содержимого
	Message   string    `json:"message,omitempty" example:"Текст найден"`           // Сообщение
}

// BlockchainStats представляет статистику блокчейна
type BlockchainStats struct {
	TotalBlocks    int       `json:"total_blocks" example:"150"`                     // Общее количество блоков
	LastBlockID    string    `json:"last_block_id" example:"000-000-150"`            // ID последнего блока
	LastBlockTime  time.Time `json:"last_block_time" example:"2026-01-05T15:04:05Z"` // Время последнего блока
	ChainValid     bool      `json:"chain_valid" example:"true"`                     // Валидность цепочки
	Difficulty     int       `json:"difficulty" example:"4"`                         // Сложность PoW
	TotalAuthors   int       `json:"total_authors" example:"75"`                     // Уникальных авторов
	AverageMinTime float64   `json:"average_mining_time" example:"2.5"`              // Среднее время майнинга (сек)
}

// ErrorResponse представляет структуру ошибки
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`                              // Статус (всегда false для ошибок)
	Error   string `json:"error" example:"Неверный формат данных"`               // Описание ошибки
	Code    int    `json:"code,omitempty" example:"400"`                         // HTTP код ошибки
	Details string `json:"details,omitempty" example:"Поле 'title' обязательно"` // Детали ошибки
}

// BadgeResponse представляет HTML badge для встраивания
type BadgeResponse struct {
	HTML string `json:"html" example:"<div class='textproof-badge'>...</div>"` // HTML код badge
}
