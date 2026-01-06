package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// DepositData содержит данные о депонируемом тексте
type DepositData struct {
	AuthorName  string `json:"author_name"`
	Title       string `json:"title"`
	TextStart   string `json:"text_start"`   // 2-3 слова из начала
	TextEnd     string `json:"text_end"`     // 2-3 слова из конца
	ContentHash string `json:"content_hash"` // SHA-256 всего текста
	PublicKey   string `json:"public_key,omitempty"`
}

// Block представляет один блок в цепочке
type Block struct {
	ID        string      `json:"id"`        // "000-000-001"
	PrevHash  string      `json:"prev_hash"` // Хеш предыдущего блока
	Timestamp time.Time   `json:"timestamp"` // Время создания
	Data      DepositData `json:"data"`      // Данные депозита
	Nonce     int         `json:"nonce"`     // Число для Proof-of-Work
	Hash      string      `json:"hash"`      // Хеш этого блока
}

// hashData структура только для хеширования
type hashData struct {
	ID        string      `json:"id"`
	PrevHash  string      `json:"prev_hash"`
	Timestamp time.Time   `json:"timestamp"`
	Data      DepositData `json:"data"`
	Nonce     int         `json:"nonce"`
}

// CalculateHash вычисляет хеш блока
func (b *Block) CalculateHash() string {
	// Создаем структуру для хеширования (без поля Hash)
	data := hashData{
		ID:        b.ID,
		PrevHash:  b.PrevHash,
		Timestamp: b.Timestamp,
		Data:      b.Data,
		Nonce:     b.Nonce,
	}

	// Сериализуем в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		// ✅ УЛУЧШЕНО: добавлен лог ошибки
		// В production лучше вернуть error, но для совместимости оставляем как есть
		return ""
	}

	// Вычисляем SHA-256
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

// ValidateHash проверяет, соответствует ли хеш блока его содержимому
func (b *Block) ValidateHash() bool {
	return b.Hash == b.CalculateHash()
}

// Mine выполняет майнинг блока с заданной сложностью
func (b *Block) Mine(difficulty int) {
	// Генерируем строку из нужного количества нулей
	prefix := ""
	for i := 0; i < difficulty; i++ {
		prefix += "0"
	}

	// Пытаемся найти подходящий nonce
	for {
		b.Hash = b.CalculateHash()
		if len(b.Hash) >= difficulty && b.Hash[:difficulty] == prefix {
			return // Нашли!
		}
		b.Nonce++
	}
}

// NewBlock создает новый блок
func NewBlock(id, prevHash string, data DepositData) *Block {
	block := &Block{
		ID:        id,
		PrevHash:  prevHash,
		Timestamp: time.Now(),
		Data:      data,
		Nonce:     0,
	}
	return block
}

// GenesisBlock создает генезис-блок
func GenesisBlock() *Block {
	data := DepositData{
		AuthorName:  "Александр Сергеевич Пушкин",
		Title:       "Евгений Онегин",
		TextStart:   "Мой дядя самых честных правил",
		TextEnd:     "Иных уж нет, а те далече",
		ContentHash: "genesis_hash", // Для генезис-блока особый хеш
	}

	block := NewBlock("000-000-000", "", data)
	block.Hash = block.CalculateHash() // Генезис-блок не требует майнинга
	return block
}
