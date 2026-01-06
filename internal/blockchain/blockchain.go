package blockchain

import (
	"fmt"
	"os"
	"sync"
)

// Blockchain представляет цепочку блоков
type Blockchain struct {
	Chain      []*Block `json:"chain"`
	Difficulty int      `json:"difficulty"`

	mu sync.RWMutex

	storage      *Storage
	blockStorage BlockStorage // для тестов

	// индекс для O(1) проверки дубликатов текста
	contentHashIndex map[string]*Block
}

// NewBlockchain создает новую цепочку блоков
func NewBlockchain(storage *Storage, difficulty int) (*Blockchain, error) {
	bc := &Blockchain{
		Difficulty: difficulty,
		storage:    storage,

		contentHashIndex: make(map[string]*Block),
	}

	// Пытаемся загрузить существующую цепочку
	loadedBC, err := storage.LoadChain()
	if err != nil {
		// Если файла нет, создаем новую цепочку
		if os.IsNotExist(err) {
			if err := bc.createGenesis(); err != nil {
				return nil, NewBlockchainError("GENESIS_CREATION_FAILED", "failed to create genesis block", err)
			}

			// Сохраняем новую цепочку
			if err := bc.saveChain(); err != nil {
				return nil, NewBlockchainError("CHAIN_SAVE_FAILED", "failed to save new chain", err)
			}

			return bc, nil
		} else {
			// Другая ошибка
			return nil, NewBlockchainError("CHAIN_LOAD_FAILED", "failed to load chain", err)
		}
	} else {
		// Используем загруженную цепочку
		bc.Chain = loadedBC.Chain
		bc.rebuildContentHashIndex()
		// Восстанавливаем из WAL, если он есть
		if err := bc.recoverFromWAL(); err != nil {
			// Пытаемся восстановить из бэкапа
			// Это не критическая ошибка - просто логируем
			// fmt.Printf("WAL recovery failed: %v\n", err)

			if err := bc.restoreFromBackup(); err != nil {
				// Если не удалось восстановить, продолжаем с текущей цепочкой
				// fmt.Printf("Backup restore failed: %v\n", err)
			}
		}

		// Проверяем целостность цепочки
		if !bc.ValidateChain() {
			// Пытаемся восстановить из бэкапа
			if err := bc.restoreFromBackup(); err != nil {
				// Если не удалось, создаем новую цепочку
				bc.Chain = []*Block{}
				if err := bc.createGenesis(); err != nil {
					return nil, NewBlockchainError("GENESIS_CREATION_FAILED", "failed to create genesis block after corruption", err)
				}
			}

			// Проверяем снова после восстановления
			if !bc.ValidateChain() {
				// Создаем новую цепочку
				bc.Chain = []*Block{}
				if err := bc.createGenesis(); err != nil {
					return nil, NewBlockchainError("GENESIS_CREATION_FAILED", "failed to create genesis block", err)
				}
			}
		}
	}
	bc.rebuildContentHashIndex()
	return bc, nil
}

// rebuildContentHashIndex перестраивает индекс хешей содержимого
func (bc *Blockchain) rebuildContentHashIndex() {
	bc.contentHashIndex = make(map[string]*Block, len(bc.Chain))

	for _, block := range bc.Chain {
		// genesis тоже попадёт, это нормально
		if block != nil && block.Data.ContentHash != "" {
			bc.contentHashIndex[block.Data.ContentHash] = block
		}
	}
}

// createGenesis создает генезис-блок
func (bc *Blockchain) createGenesis() error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	genesis := GenesisBlock()
	bc.Chain = []*Block{genesis}
	return nil
}

// recoverFromWAL восстанавливает состояние из WAL
func (bc *Blockchain) recoverFromWAL() error {
	blocks, err := bc.storage.ReadWAL()
	if err != nil {
		// Если файла WAL нет - это нормально
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if len(blocks) == 0 {
		return nil // WAL пуст
	}

	// Добавляем блоки из WAL в цепочку
	for _, block := range blocks {
		if err := bc.addBlockInternal(block); err != nil {
			return fmt.Errorf("failed to add block from WAL: %w", err)
		}
	}

	// Сохраняем восстановленную цепочку
	if err := bc.saveChain(); err != nil {
		return fmt.Errorf("failed to save recovered chain: %w", err)
	}

	// Очищаем WAL после успешного восстановления
	if err := bc.storage.ClearWAL(); err != nil {
		// Не критическая ошибка
		return nil
	}

	return nil
}

// restoreFromBackup восстанавливает цепочку из последнего бэкапа
func (bc *Blockchain) restoreFromBackup() error {
	if err := bc.storage.RestoreFromBackup(); err != nil {
		return NewBlockchainError("BACKUP_RESTORE_FAILED", "failed to restore from backup", err)
	}

	// Загружаем восстановленную цепочку
	loadedBC, err := bc.storage.LoadChain()
	if err != nil {
		return NewBlockchainError("CHAIN_LOAD_FAILED", "failed to load restored chain", err)
	}

	bc.mu.Lock()
	bc.Chain = loadedBC.Chain
	bc.mu.Unlock()

	return nil
}

// saveChain сохраняет цепочку
func (bc *Blockchain) saveChain() error {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return bc.storage.SaveChain(bc)
}

// GetLastBlock возвращает последний блок в цепочке
func (bc *Blockchain) GetLastBlock() *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if len(bc.Chain) == 0 {
		return nil
	}
	return bc.Chain[len(bc.Chain)-1]
}

// GetBlockByID ищет блок по ID
func (bc *Blockchain) GetBlockByID(id string) (*Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for _, block := range bc.Chain {
		if block.ID == id {
			return block, nil
		}
	}

	return nil, ErrBlockNotFound
}

// GenerateNextID генерирует ID для следующего блока
func (bc *Blockchain) GenerateNextID() (string, error) {
	lastBlock := bc.GetLastBlock()
	if lastBlock == nil {
		return "000-000-000", nil
	}

	id, err := incrementID(lastBlock.ID)
	if err != nil {
		return "", NewBlockchainError("ID_GENERATION_FAILED", "failed to generate next ID", err)
	}

	return id, nil
}

// AddBlock добавляет новый блок в цепочку
func (bc *Blockchain) AddBlock(data DepositData) (*Block, error) {
	// Проверяем на дубликат
	if existing, exists := bc.HasContentHash(data.ContentHash); exists {
		return existing, nil
	}

	// Генерируем ID для нового блока
	nextID, err := bc.GenerateNextID()
	if err != nil {
		return nil, err
	}

	// Получаем хеш последнего блока
	lastBlock := bc.GetLastBlock()
	prevHash := "0"
	if lastBlock != nil {
		prevHash = lastBlock.Hash
	}

	// Создаем новый блок
	block := NewBlock(nextID, prevHash, data)

	// Майним блок
	block.Mine(bc.Difficulty)

	// Записываем в WAL (только если используется файловое хранилище)
	if bc.storage != nil {
		if err := bc.storage.WriteToWAL(block); err != nil {
			return nil, NewBlockchainError("WAL_WRITE_FAILED", "failed to write block to WAL", err)
		}
	}

	// Добавляем блок в цепочку
	if err := bc.addBlockInternal(block); err != nil {
		if dup, ok := err.(*DuplicateBlockError); ok {
			return dup.Block, nil
		}
		return nil, err
	}

	// Сохраняем блок в хранилище
	if bc.blockStorage != nil {
		if err := bc.blockStorage.SaveBlock(block); err != nil {
			return nil, NewBlockchainError("BLOCK_SAVE_FAILED", "failed to save block", err)
		}
	} else if bc.storage != nil {
		// Сохраняем цепочку
		if err := bc.saveChain(); err != nil {
			return nil, NewBlockchainError("CHAIN_SAVE_FAILED", "failed to save chain", err)
		}

		// Очищаем WAL после успешного сохранения
		if err := bc.storage.ClearWAL(); err != nil {
			// Не критическая ошибка
			// fmt.Printf("Warning: failed to clear WAL: %v\n", err)
		}
	}

	return block, nil
}

// addBlockInternal добавляет блок в цепочку в памяти
func (bc *Blockchain) addBlockInternal(block *Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Атомарная проверка дубликата
	if existing, exists := bc.contentHashIndex[block.Data.ContentHash]; exists {
		return &DuplicateBlockError{Block: existing}
	}

	// Проверяем валидность блока
	if !block.ValidateHash() {
		return ErrInvalidBlockHash
	}

	// Проверяем сложность
	prefix := ""
	for i := 0; i < bc.Difficulty; i++ {
		prefix += "0"
	}
	if block.Hash[:bc.Difficulty] != prefix {
		return ErrInvalidDifficulty
	}

	// Проверяем связь с предыдущим блоком
	if len(bc.Chain) > 0 {
		lastBlock := bc.Chain[len(bc.Chain)-1]
		if block.PrevHash != lastBlock.Hash {
			return ErrPrevHashMismatch
		}
	}

	// Добавляем блок
	bc.Chain = append(bc.Chain, block)
	bc.contentHashIndex[block.Data.ContentHash] = block

	return nil
}

// ValidateChain проверяет целостность всей цепочки
func (bc *Blockchain) ValidateChain() bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if len(bc.Chain) == 0 {
		return true
	}

	// Проверяем генезис-блок
	if !bc.Chain[0].ValidateHash() {
		return false
	}

	// Проверяем остальные блоки
	for i := 1; i < len(bc.Chain); i++ {
		current := bc.Chain[i]
		previous := bc.Chain[i-1]

		// Проверяем хеш текущего блока
		if !current.ValidateHash() {
			return false
		}

		// Проверяем связь с предыдущим блоком
		if current.PrevHash != previous.Hash {
			return false
		}

		// Проверяем сложность
		prefix := ""
		for j := 0; j < bc.Difficulty; j++ {
			prefix += "0"
		}
		if current.Hash[:bc.Difficulty] != prefix {
			return false
		}
	}

	return true
}

// GetChainInfo возвращает информацию о цепочке
func (bc *Blockchain) GetChainInfo() map[string]interface{} {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	info := map[string]interface{}{
		"length":     len(bc.Chain),
		"difficulty": bc.Difficulty,
		"valid":      bc.ValidateChain(),
	}

	if len(bc.Chain) > 0 {
		info["last_block"] = bc.Chain[len(bc.Chain)-1].ID
		info["first_block"] = bc.Chain[0].ID
	}

	return info
}

// GetAllBlocks возвращает все блоки (для отладки)
func (bc *Blockchain) GetAllBlocks() []*Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Возвращаем копию, чтобы избежать гонок данных
	blocks := make([]*Block, len(bc.Chain))
	copy(blocks, bc.Chain)
	return blocks
}

// GetBlocksRange возвращает диапазон блоков
func (bc *Blockchain) GetBlocksRange(start, end int) ([]*Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if start < 0 || end > len(bc.Chain) || start >= end {
		return nil, NewBlockchainError("INVALID_RANGE", "invalid block range", nil)
	}

	blocks := make([]*Block, end-start)
	copy(blocks, bc.Chain[start:end])
	return blocks, nil
}

// HasContentHash проверяет, существует ли блок с данным хешем содержимого
func (bc *Blockchain) HasContentHash(hash string) (*Block, bool) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	block, ok := bc.contentHashIndex[hash]
	return block, ok
}
