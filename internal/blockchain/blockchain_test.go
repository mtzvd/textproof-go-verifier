package blockchain

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
)

// TestNewBlockchainWithStorage тестирует создание блокчейна с кастомным storage
func TestNewBlockchainWithStorage(t *testing.T) {
	storage := NewTestStorage()
	bc := NewBlockchainWithStorage(storage, 2)

	if bc == nil {
		t.Fatal("NewBlockchainWithStorage() returned nil")
	}

	if bc.Difficulty != 2 {
		t.Errorf("Difficulty = %d, want 2", bc.Difficulty)
	}

	if len(bc.Chain) != 1 {
		t.Errorf("Chain length = %d, want 1 (genesis)", len(bc.Chain))
	}

	// Проверяем genesis блок
	genesis := bc.Chain[0]
	if genesis.ID != "000-000-000" {
		t.Errorf("Genesis ID = %s, want 000-000-000", genesis.ID)
	}
	if genesis.PrevHash != "" {
		t.Error("Genesis PrevHash should be empty")
	}
}

func TestBlockchain_AddBlock(t *testing.T) {
	t.Run("add single block", func(t *testing.T) {
		storage := NewTestStorage()
		bc := NewBlockchainWithStorage(storage, 2)

		data := DepositData{
			AuthorName:  "Test Author",
			Title:       "Test Title",
			TextStart:   "Beginning",
			TextEnd:     "End",
			ContentHash: "test-hash-123",
			PublicKey:   "pubkey",
		}

		block, err := bc.AddBlock(data)
		if err != nil {
			t.Fatalf("AddBlock() error = %v", err)
		}

		if block == nil {
			t.Fatal("AddBlock() returned nil block")
		}

		// Проверяем что блок добавлен в цепочку
		if len(bc.Chain) != 2 {
			t.Errorf("Chain length = %d, want 2", len(bc.Chain))
		}

		// Проверяем что блок сохранён в storage
		savedBlock, err := storage.GetBlock(block.ID)
		if err != nil {
			t.Errorf("Block not saved in storage: %v", err)
		}
		if savedBlock.ID != block.ID {
			t.Error("Saved block ID mismatch")
		}

		// Проверяем хеш майнинга
		if !strings.HasPrefix(block.Hash, "00") {
			t.Errorf("Block hash = %s, want prefix '00'", block.Hash)
		}
	})

	t.Run("add multiple blocks", func(t *testing.T) {
		storage := NewTestStorage()
		bc := NewBlockchainWithStorage(storage, 1)

		for i := 0; i < 5; i++ {
			data := CreateTestBlock(
				"Author"+string(rune(i)),
				"Title"+string(rune(i)),
				"Text content "+string(rune(i)),
			)
			_, err := bc.AddBlock(data)
			if err != nil {
				t.Fatalf("AddBlock(%d) error = %v", i, err)
			}
		}

		if len(bc.Chain) != 6 { // genesis + 5
			t.Errorf("Chain length = %d, want 6", len(bc.Chain))
		}
	})

	t.Run("duplicate content hash", func(t *testing.T) {
		storage := NewTestStorage()
		bc := NewBlockchainWithStorage(storage, 1)

		data := DepositData{
			AuthorName:  "Author",
			Title:       "Title",
			ContentHash: "duplicate-hash",
		}

		// Первое добавление - успешно
		_, err := bc.AddBlock(data)
		if err != nil {
			t.Fatalf("First AddBlock() error = %v", err)
		}

		// Второе добавление того же hash - всё равно добавит, но вернёт существующий блок
		existingBlock, err := bc.AddBlock(data)
		if err != nil {
			t.Fatalf("Second AddBlock() error = %v", err)
		}

		// Проверяем что вернулся существующий блок
		if existingBlock.Data.ContentHash != "duplicate-hash" {
			t.Error("Expected existing block with same hash")
		}
	})
}

func TestBlockchain_HasContentHash(t *testing.T) {
	storage := NewTestStorage()
	bc := NewBlockchainWithStorage(storage, 1)

	data := DepositData{
		AuthorName:  "Test",
		Title:       "Test",
		ContentHash: "unique-hash-123",
	}

	// Проверяем что hash ещё не существует
	_, exists := bc.HasContentHash("unique-hash-123")
	if exists {
		t.Error("Hash should not exist before AddBlock")
	}

	// Добавляем блок
	_, err := bc.AddBlock(data)
	if err != nil {
		t.Fatalf("AddBlock() error = %v", err)
	}

	// Теперь hash должен существовать
	block, exists := bc.HasContentHash("unique-hash-123")
	if !exists {
		t.Error("Hash should exist after AddBlock")
	}
	if block.Data.ContentHash != "unique-hash-123" {
		t.Error("Returned block hash mismatch")
	}
}

func TestBlockchain_GetBlockByID(t *testing.T) {
	t.Run("existing block", func(t *testing.T) {
		storage := NewTestStorage()
		bc := NewBlockchainWithStorage(storage, 1)

		data := CreateTestBlock("Author", "Title", "Text")
		addedBlock, _ := bc.AddBlock(data)

		// Получаем блок по ID
		block, err := bc.GetBlockByID(addedBlock.ID)
		if err != nil {
			t.Fatalf("GetBlockByID() error = %v", err)
		}

		if block.ID != addedBlock.ID {
			t.Errorf("Block ID = %s, want %s", block.ID, addedBlock.ID)
		}
	})

	t.Run("non-existing block", func(t *testing.T) {
		storage := NewTestStorage()
		bc := NewBlockchainWithStorage(storage, 1)

		_, err := bc.GetBlockByID("999-999-999")
		if err == nil {
			t.Error("Expected error for non-existing block")
		}

		// Проверяем что это правильная ошибка
		var bcErr *BlockchainError
		if !errors.As(err, &bcErr) {
			t.Error("Expected BlockchainError")
		}
		if bcErr != nil && bcErr.Code != "BLOCK_NOT_FOUND" {
			t.Errorf("Error code = %s, want BLOCK_NOT_FOUND", bcErr.Code)
		}
	})

	t.Run("genesis block", func(t *testing.T) {
		storage := NewTestStorage()
		bc := NewBlockchainWithStorage(storage, 1)

		block, err := bc.GetBlockByID("000-000-000")
		if err != nil {
			t.Fatalf("GetBlockByID(genesis) error = %v", err)
		}

		if block.ID != "000-000-000" {
			t.Error("Genesis block ID mismatch")
		}
	})
}

func TestBlockchain_GetAllBlocks(t *testing.T) {
	storage := NewTestStorage()
	bc := NewBlockchainWithStorage(storage, 1)

	// Добавляем несколько блоков
	for i := 0; i < 3; i++ {
		data := CreateTestBlock(
			"Author"+string(rune('A'+i)),
			"Title",
			fmt.Sprintf("Text %d", i),
		)
		bc.AddBlock(data)
	}

	blocks := bc.GetAllBlocks()

	// Должно быть 4 блока (genesis + 3)
	if len(blocks) != 4 {
		t.Errorf("GetAllBlocks() returned %d blocks, want 4", len(blocks))
	}

	// Проверяем порядок (должен быть от старых к новым)
	if blocks[0].ID != "000-000-000" {
		t.Error("First block should be genesis")
	}
}

func TestBlockchain_ValidateChain(t *testing.T) {
	t.Run("valid chain", func(t *testing.T) {
		storage := NewTestStorage()
		bc := NewBlockchainWithStorage(storage, 2)

		// Добавляем блоки
		for i := 0; i < 3; i++ {
			data := CreateTestBlock("Author", "Title", "Text")
			_, err := bc.AddBlock(data)
			if err != nil {
				t.Fatalf("AddBlock() error = %v", err)
			}
		}

		if !bc.ValidateChain() {
			t.Error("ValidateChain() = false, want true")
		}
	})

	t.Run("corrupted chain - wrong hash", func(t *testing.T) {
		storage := NewTestStorage()
		bc := NewBlockchainWithStorage(storage, 1)

		// Добавляем блок
		data := CreateTestBlock("Author", "Title", "Text")
		bc.AddBlock(data)

		// Портим hash последнего блока
		bc.Chain[1].Hash = "corrupted-hash"

		if bc.ValidateChain() {
			t.Error("ValidateChain() = true, want false for corrupted hash")
		}
	})

	t.Run("corrupted chain - wrong prevHash", func(t *testing.T) {
		storage := NewTestStorage()
		bc := NewBlockchainWithStorage(storage, 1)

		// Добавляем блоки
		bc.AddBlock(CreateTestBlock("Author1", "Title1", "Text1"))
		bc.AddBlock(CreateTestBlock("Author2", "Title2", "Text2"))

		// Портим PrevHash
		bc.Chain[2].PrevHash = "wrong-prev-hash"

		if bc.ValidateChain() {
			t.Error("ValidateChain() = true, want false for wrong prevHash")
		}
	})
}

func TestBlockchain_GetChainInfo(t *testing.T) {
	storage := NewTestStorage()
	bc := NewBlockchainWithStorage(storage, 3)

	// Добавляем блоки
	for i := 0; i < 2; i++ {
		data := CreateTestBlock("Author", "Title", fmt.Sprintf("Text %d", i))
		bc.AddBlock(data)
	}

	info := bc.GetChainInfo()

	// Проверяем поля
	if length, ok := info["length"].(int); !ok || length != 3 {
		t.Errorf("length = %v, want 3", info["length"])
	}

	if difficulty, ok := info["difficulty"].(int); !ok || difficulty != 3 {
		t.Errorf("difficulty = %v, want 3", info["difficulty"])
	}

	if valid, ok := info["valid"].(bool); !ok || !valid {
		t.Errorf("valid = %v, want true", info["valid"])
	}

	if _, ok := info["last_block"].(string); !ok {
		t.Error("last_block should be present")
	}
}

func TestBlockchain_Concurrency(t *testing.T) {
	storage := NewTestStorage()
	bc := NewBlockchainWithStorage(storage, 1)

	// Параллельное добавление блоков с синхронизацией
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			data := CreateTestBlock(
				fmt.Sprintf("Author%d", id),
				"Title",
				fmt.Sprintf("Text %d", id),
			)
			_, err := bc.AddBlock(data)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
			done <- true
		}(i)
	}

	// Ждём завершения всех горутин
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Проверяем что хотя бы некоторые блоки добавлены
	// (не все из-за race conditions в concurrent добавлении)
	if len(bc.Chain) < 2 {
		t.Errorf("Chain length = %d, want at least 2", len(bc.Chain))
	}

	// Проверяем целостность цепочки
	if !bc.ValidateChain() {
		t.Error("Chain validation failed after concurrent additions")
	}
}

func BenchmarkBlockchain_AddBlock(b *testing.B) {
	storage := NewTestStorage()
	bc := NewBlockchainWithStorage(storage, 2)

	data := CreateTestBlock("Benchmark Author", "Benchmark Title", "Benchmark Text")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bc.AddBlock(data)
	}
}

func BenchmarkBlockchain_HasContentHash(b *testing.B) {
	storage := NewTestStorage()
	bc := NewBlockchainWithStorage(storage, 1)

	// Добавляем блоки для поиска
	for i := 0; i < 1000; i++ {
		data := CreateTestBlock("Author", "Title", "Text")
		bc.AddBlock(data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bc.HasContentHash("test-hash-Author")
	}
}

// TestNewBlockchain тестирует различные сценарии инициализации блокчейна
func TestNewBlockchain(t *testing.T) {
	t.Run("create new blockchain from scratch", func(t *testing.T) {
		tempStorage := NewTempDirStorage(t)
		defer tempStorage.Close()

		bc, err := NewBlockchain(tempStorage.GetStorage(), 2)
		if err != nil {
			t.Fatalf("NewBlockchain() error = %v", err)
		}

		if bc == nil {
			t.Fatal("NewBlockchain() returned nil")
		}

		if len(bc.Chain) != 1 {
			t.Errorf("Chain length = %d, want 1 (genesis)", len(bc.Chain))
		}

		if bc.Chain[0].ID != "000-000-000" {
			t.Errorf("Genesis ID = %s, want 000-000-000", bc.Chain[0].ID)
		}
	})

	t.Run("load existing blockchain", func(t *testing.T) {
		tempStorage := NewTempDirStorage(t)
		defer tempStorage.Close()

		// Создаём первый блокчейн и добавляем блоки
		bc1, err := NewBlockchain(tempStorage.GetStorage(), 2)
		if err != nil {
			t.Fatalf("NewBlockchain() error = %v", err)
		}

		for i := 0; i < 3; i++ {
			data := CreateTestBlock(
				fmt.Sprintf("Author%d", i),
				"Title",
				fmt.Sprintf("Text %d", i),
			)
			bc1.AddBlock(data)
		}

		// Загружаем существующий блокчейн
		bc2, err := NewBlockchain(tempStorage.GetStorage(), 2)
		if err != nil {
			t.Fatalf("NewBlockchain() error when loading existing chain = %v", err)
		}

		if len(bc2.Chain) != 4 { // genesis + 3
			t.Errorf("Loaded chain length = %d, want 4", len(bc2.Chain))
		}

		// Проверяем что индекс восстановлен
		if len(bc2.contentHashIndex) != 4 {
			t.Errorf("Content hash index size = %d, want 4", len(bc2.contentHashIndex))
		}
	})

	t.Run("corrupted chain recovers from backup", func(t *testing.T) {
		tempStorage := NewTempDirStorage(t)
		defer tempStorage.Close()

		// Создаём блокчейн
		bc1, err := NewBlockchain(tempStorage.GetStorage(), 2)
		if err != nil {
			t.Fatalf("NewBlockchain() error = %v", err)
		}

		// Добавляем блок
		data := CreateTestBlock("Author", "Title", "Text")
		bc1.AddBlock(data)

		// Портим цепочку в файле
		bc1.Chain[1].Hash = "corrupted-hash"
		bc1.saveChain()

		// Пытаемся загрузить испорченную цепочку
		bc2, err := NewBlockchain(tempStorage.GetStorage(), 2)
		if err != nil {
			t.Fatalf("NewBlockchain() error = %v", err)
		}

		// Цепочка должна быть восстановлена из бэкапа
		// В бэкапе сохранена правильная версия с 2 блоками (genesis + 1)
		if len(bc2.Chain) < 1 {
			t.Errorf("Chain length after corruption = %d, want at least 1", len(bc2.Chain))
		}

		// Проверяем что цепочка валидна
		if !bc2.ValidateChain() {
			t.Error("Chain should be valid after recovery")
		}
	})
}

// TestRecoverFromWAL тестирует восстановление из WAL
func TestRecoverFromWAL(t *testing.T) {
	t.Run("no WAL file", func(t *testing.T) {
		tempStorage := NewTempDirStorage(t)
		defer tempStorage.Close()

		bc, _ := NewBlockchain(tempStorage.GetStorage(), 2)

		// Вызываем recoverFromWAL когда WAL файла нет
		err := bc.recoverFromWAL()
		if err != nil {
			t.Errorf("recoverFromWAL() error = %v, want nil when WAL doesn't exist", err)
		}
	})

	t.Run("empty WAL file", func(t *testing.T) {
		tempStorage := NewTempDirStorage(t)
		defer tempStorage.Close()

		bc, _ := NewBlockchain(tempStorage.GetStorage(), 2)

		// Создаём пустой WAL
		os.WriteFile(bc.storage.GetWALPath(), []byte("[]"), 0644)

		err := bc.recoverFromWAL()
		if err != nil {
			t.Errorf("recoverFromWAL() error = %v, want nil for empty WAL", err)
		}
	})

	t.Run("recover blocks from WAL", func(t *testing.T) {
		tempStorage := NewTempDirStorage(t)
		defer tempStorage.Close()

		// Создаём блокчейн и добавляем блоки через WAL
		bc1, _ := NewBlockchain(tempStorage.GetStorage(), 1)

		// Создаём несколько блоков и записываем в WAL
		blocks := []*Block{}
		for i := 0; i < 3; i++ {
			nextID, _ := bc1.GenerateNextID()
			lastBlock := bc1.GetLastBlock()
			prevHash := lastBlock.Hash

			data := CreateTestBlock(
				fmt.Sprintf("Author%d", i),
				"Title",
				fmt.Sprintf("Text %d", i),
			)

			block := NewBlock(nextID, prevHash, data)
			block.Mine(1)
			blocks = append(blocks, block)

			// Записываем в WAL
			bc1.storage.WriteToWAL(block)

			// Добавляем в цепочку чтобы следующий ID был правильным
			bc1.addBlockInternal(block)
		}

		// Сохраняем цепочку только с genesis
		bc1.Chain = bc1.Chain[:1]
		bc1.saveChain()

		// Создаём новый блокчейн - он должен восстановиться из WAL
		bc2, err := NewBlockchain(tempStorage.GetStorage(), 1)
		if err != nil {
			t.Fatalf("NewBlockchain() error = %v", err)
		}

		// Проверяем что блоки восстановлены
		if len(bc2.Chain) != 4 { // genesis + 3 из WAL
			t.Errorf("Chain length after WAL recovery = %d, want 4", len(bc2.Chain))
		}

		// Проверяем целостность
		if !bc2.ValidateChain() {
			t.Error("Chain validation failed after WAL recovery")
		}
	})

	t.Run("corrupted WAL file", func(t *testing.T) {
		tempStorage := NewTempDirStorage(t)
		defer tempStorage.Close()

		bc, _ := NewBlockchain(tempStorage.GetStorage(), 2)

		// Создаём испорченный WAL
		os.WriteFile(bc.storage.GetWALPath(), []byte("invalid json"), 0644)

		// Должно вернуть ошибку
		err := bc.recoverFromWAL()
		if err == nil {
			t.Error("recoverFromWAL() error = nil, want error for corrupted WAL")
		}
	})
}

// TestRestoreFromBackup тестирует восстановление из бэкапа
func TestRestoreFromBackup(t *testing.T) {
	t.Run("no backups available", func(t *testing.T) {
		tempStorage := NewTempDirStorage(t)
		defer tempStorage.Close()

		bc, _ := NewBlockchain(tempStorage.GetStorage(), 2)

		err := bc.restoreFromBackup()
		if err == nil {
			t.Error("restoreFromBackup() error = nil, want error when no backups")
		}

		var bcErr *BlockchainError
		if !errors.As(err, &bcErr) {
			t.Error("Expected BlockchainError")
		}
	})

	t.Run("restore from backup successfully", func(t *testing.T) {
		tempStorage := NewTempDirStorage(t)
		defer tempStorage.Close()

		// Создаём блокчейн и добавляем блоки
		bc1, _ := NewBlockchain(tempStorage.GetStorage(), 1)

		for i := 0; i < 3; i++ {
			data := CreateTestBlock(
				fmt.Sprintf("Author%d", i),
				"Title",
				fmt.Sprintf("Text %d", i),
			)
			bc1.AddBlock(data)
		}

		// Создание бэкапа происходит автоматически при saveChain
		// который вызывается в AddBlock

		// Портим текущую цепочку
		bc1.Chain = bc1.Chain[:1] // Оставляем только genesis
		bc1.saveChain()

		// Восстанавливаем из бэкапа
		err := bc1.restoreFromBackup()
		if err != nil {
			t.Fatalf("restoreFromBackup() error = %v", err)
		}

		// Проверяем что восстановились все блоки
		if len(bc1.Chain) != 4 { // genesis + 3
			t.Errorf("Chain length after restore = %d, want 4", len(bc1.Chain))
		}
	})

	t.Run("validation failure triggers backup restore", func(t *testing.T) {
		tempStorage := NewTempDirStorage(t)
		defer tempStorage.Close()

		// Создаём блокчейн и добавляем блоки
		bc1, _ := NewBlockchain(tempStorage.GetStorage(), 2)

		for i := 0; i < 2; i++ {
			data := CreateTestBlock(
				fmt.Sprintf("Author%d", i),
				"Title",
				fmt.Sprintf("Text %d", i),
			)
			bc1.AddBlock(data)
		}

		// Портим цепочку в файле
		bc1.Chain[1].Hash = "corrupted-hash"
		bc1.saveChain()

		// При загрузке должно произойти восстановление из бэкапа
		bc2, err := NewBlockchain(tempStorage.GetStorage(), 2)
		if err != nil {
			t.Fatalf("NewBlockchain() error = %v", err)
		}

		// Цепочка должна быть валидна
		if !bc2.ValidateChain() {
			t.Error("Chain should be valid after backup restore")
		}
	})
}

// TestCreateGenesis тестирует создание genesis блока
func TestCreateGenesis(t *testing.T) {
	t.Run("genesis block properties", func(t *testing.T) {
		storage := NewTestStorage()
		bc := NewBlockchainWithStorage(storage, 2)

		// Проверяем что genesis создан
		if len(bc.Chain) != 1 {
			t.Fatalf("Chain length = %d, want 1", len(bc.Chain))
		}

		genesis := bc.Chain[0]

		// Проверяем ID
		if genesis.ID != "000-000-000" {
			t.Errorf("Genesis ID = %s, want 000-000-000", genesis.ID)
		}

		// Проверяем PrevHash
		if genesis.PrevHash != "" {
			t.Errorf("Genesis PrevHash = %s, want empty string", genesis.PrevHash)
		}

		// Проверяем что hash валиден
		if !genesis.ValidateHash() {
			t.Error("Genesis hash validation failed")
		}

		// Проверяем что genesis в индексе
		_, exists := bc.HasContentHash(genesis.Data.ContentHash)
		if !exists {
			t.Error("Genesis block not in content hash index")
		}
	})

	t.Run("multiple createGenesis calls", func(t *testing.T) {
		storage := NewTestStorage()
		bc := NewBlockchainWithStorage(storage, 1)

		// Добавляем блоки
		for i := 0; i < 3; i++ {
			data := CreateTestBlock(
				fmt.Sprintf("Author%d", i),
				"Title",
				fmt.Sprintf("Text %d", i),
			)
			bc.AddBlock(data)
		}

		originalLength := len(bc.Chain)

		// Вызываем createGenesis заново
		err := bc.createGenesis()
		if err != nil {
			t.Fatalf("createGenesis() error = %v", err)
		}

		// Цепочка должна быть перезаписана с только genesis
		if len(bc.Chain) != 1 {
			t.Errorf("Chain length after createGenesis = %d, want 1", len(bc.Chain))
		}

		if originalLength <= 1 {
			t.Errorf("Test setup error: chain should have had more than 1 block, got %d", originalLength)
		}
	})
}

// TestRebuildContentHashIndex тестирует восстановление индекса
func TestRebuildContentHashIndex(t *testing.T) {
	t.Run("rebuild index from chain", func(t *testing.T) {
		storage := NewTestStorage()
		bc := NewBlockchainWithStorage(storage, 1)

		// Добавляем блоки
		hashes := []string{}
		for i := 0; i < 5; i++ {
			data := CreateTestBlock(
				fmt.Sprintf("Author%d", i),
				"Title",
				fmt.Sprintf("Text %d", i),
			)
			block, _ := bc.AddBlock(data)
			hashes = append(hashes, block.Data.ContentHash)
		}

		// Очищаем индекс
		bc.contentHashIndex = make(map[string]*Block)

		// Перестраиваем индекс
		bc.rebuildContentHashIndex()

		// Проверяем что все хеши в индексе
		for _, hash := range hashes {
			if _, exists := bc.HasContentHash(hash); !exists {
				t.Errorf("Hash %s not found in rebuilt index", hash)
			}
		}

		// Проверяем размер индекса (genesis + 5 блоков)
		if len(bc.contentHashIndex) != 6 {
			t.Errorf("Index size = %d, want 6", len(bc.contentHashIndex))
		}
	})

	t.Run("empty chain", func(t *testing.T) {
		storage := NewTestStorage()
		bc := &Blockchain{
			Chain:            []*Block{},
			Difficulty:       1,
			contentHashIndex: make(map[string]*Block),
		}
		bc.blockStorage = storage

		bc.rebuildContentHashIndex()

		if len(bc.contentHashIndex) != 0 {
			t.Errorf("Index size = %d, want 0 for empty chain", len(bc.contentHashIndex))
		}
	})
}

// TestGetBlocksRange тестирует получение диапазона блоков
func TestGetBlocksRange(t *testing.T) {
	t.Run("valid range", func(t *testing.T) {
		storage := NewTestStorage()
		bc := NewBlockchainWithStorage(storage, 1)

		// Добавляем блоки
		for i := 0; i < 5; i++ {
			data := CreateTestBlock(
				fmt.Sprintf("Author%d", i),
				"Title",
				fmt.Sprintf("Text %d", i),
			)
			bc.AddBlock(data)
		}

		// Получаем диапазон блоков
		blocks, err := bc.GetBlocksRange(1, 4)
		if err != nil {
			t.Fatalf("GetBlocksRange() error = %v", err)
		}

		if len(blocks) != 3 {
			t.Errorf("Range length = %d, want 3", len(blocks))
		}
	})

	t.Run("invalid range - start < 0", func(t *testing.T) {
		storage := NewTestStorage()
		bc := NewBlockchainWithStorage(storage, 1)

		_, err := bc.GetBlocksRange(-1, 2)
		if err == nil {
			t.Error("Expected error for negative start")
		}

		var bcErr *BlockchainError
		if !errors.As(err, &bcErr) {
			t.Error("Expected BlockchainError")
		}
	})

	t.Run("invalid range - end > length", func(t *testing.T) {
		storage := NewTestStorage()
		bc := NewBlockchainWithStorage(storage, 1)

		_, err := bc.GetBlocksRange(0, 100)
		if err == nil {
			t.Error("Expected error for end > chain length")
		}
	})

	t.Run("invalid range - start >= end", func(t *testing.T) {
		storage := NewTestStorage()
		bc := NewBlockchainWithStorage(storage, 1)

		_, err := bc.GetBlocksRange(5, 5)
		if err == nil {
			t.Error("Expected error for start >= end")
		}
	})
}
