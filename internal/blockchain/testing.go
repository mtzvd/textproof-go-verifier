package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// NewBlockchainWithStorage создает новый блокчейн для тестов
func NewBlockchainWithStorage(storage interface{}, difficulty int) *Blockchain {
	// Проверяем если это TempDirStorage - у него есть и BlockStorage и реальный Storage
	if tempStorage, ok := storage.(*TempDirStorage); ok {
		bc, err := NewBlockchain(tempStorage.GetStorage(), difficulty)
		if err != nil {
			return nil
		}
		bc.blockStorage = tempStorage
		return bc
	}

	// Если передан BlockStorage, используем его напрямую для тестирования
	if blockStorage, ok := storage.(BlockStorage); ok {
		bc := &Blockchain{
			Chain:            make([]*Block, 0),
			Difficulty:       difficulty,
			storage:          nil, // не используем файловое хранилище в тестах
			blockStorage:     blockStorage,
			contentHashIndex: make(map[string]*Block),
		}

		// Создаем genesis блок
		genesis := GenesisBlock()
		bc.Chain = append(bc.Chain, genesis)
		bc.contentHashIndex[genesis.Data.ContentHash] = genesis

		// Сохраняем genesis в тестовое хранилище
		if err := blockStorage.SaveBlock(genesis); err != nil {
			return nil
		}

		return bc
	}

	// Иначе создаем обычное файловое хранилище
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("test-bc-%d", time.Now().UnixNano()))

	realStorage, err := NewStorage(tempDir)
	if err != nil {
		return nil
	}

	bc, err := NewBlockchain(realStorage, difficulty)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil
	}

	return bc
}

// CreateTestBlock создает тестовый блок с данными
func CreateTestBlock(authorName, title, text string) DepositData {
	textStart := text
	if len(text) > 50 {
		textStart = text[:50]
	}

	textEnd := text
	if len(text) > 50 {
		textEnd = text[len(text)-50:]
	}

	// Compute the actual content hash from the text
	hash := sha256.Sum256([]byte(text))
	contentHash := hex.EncodeToString(hash[:])

	return DepositData{
		AuthorName:  authorName,
		Title:       title,
		TextStart:   textStart,
		TextEnd:     textEnd,
		ContentHash: contentHash,
		PublicKey:   "test-public-key",
	}
}
