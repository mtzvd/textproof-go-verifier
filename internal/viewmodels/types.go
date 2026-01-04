package viewmodels

import "time"

// Запрос на депонирование текста
type DepositRequest struct {
	AuthorName string `json:"author_name"`
	Title      string `json:"title"`
	Text       string `json:"text"`
	PublicKey  string `json:"public_key,omitempty"`
}

// Ответ на депонирование
type DepositResponse struct {
	ID        string    `json:"id"`
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
	QRCodeURL string    `json:"qrcode_url"`
	BadgeURL  string    `json:"badge_url"`
	Duplicate bool      `json:"duplicate"`
}

// Запрос на проверку по ID
type VerifyByIDRequest struct {
	ID string `json:"id"`
}

// Запрос на проверку по тексту
type VerifyByTextRequest struct {
	Text string `json:"text"`
}

// Ответ на проверку
type VerificationResponse struct {
	Found     bool      `json:"found"`
	BlockID   string    `json:"block_id,omitempty"`
	Author    string    `json:"author,omitempty"`
	Title     string    `json:"title,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	Hash      string    `json:"hash,omitempty"`
	Matches   bool      `json:"matches,omitempty"` // Совпадает ли хеш
}

// Ответ со статистикой
type StatsResponse struct {
	TotalBlocks   int       `json:"total_blocks"`
	UniqueAuthors int       `json:"unique_authors"`
	LastAdded     time.Time `json:"last_added"`
	ChainValid    bool      `json:"chain_valid"`
}

// Ответ с информацией о блокчейне
type BlockchainInfoResponse struct {
	Length     int    `json:"length"`
	Difficulty int    `json:"difficulty"`
	Valid      bool   `json:"valid"`
	LastBlock  string `json:"last_block"`
}

// Общий ответ об ошибке
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Code    string `json:"code,omitempty"`
}

// ViewModel результата депонирования для отображения на странице
type DepositResultVM struct {
	ID        string
	Title     string
	Author    string
	Hash      string
	Timestamp time.Time
	QRCodeURL string
	BadgeURL  string
	VerifyURL string
}

// FlashData структура для передачи flash-сообщений в шаблон
type FlashData struct {
	Show        bool
	Type        string
	Message     string
	IsDuplicate bool
}
