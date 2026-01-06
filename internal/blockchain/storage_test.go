package blockchain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStorage(t *testing.T) {
	t.Run("create new storage", func(t *testing.T) {
		tempDir := filepath.Join(os.TempDir(), "textproof-test-storage")
		defer os.RemoveAll(tempDir)

		storage, err := NewStorage(tempDir)
		AssertNoError(t, err, "NewStorage should not error")
		AssertNotNil(t, storage, "Storage should not be nil")

		// Проверяем что директории созданы
		_, err = os.Stat(tempDir)
		AssertNoError(t, err, "Data directory should exist")

		backupDir := filepath.Join(tempDir, "backups")
		_, err = os.Stat(backupDir)
		AssertNoError(t, err, "Backup directory should exist")
	})

	t.Run("invalid directory", func(t *testing.T) {
		// Пытаемся создать storage с недопустимыми символами в пути
		invalidPath := string([]byte{0x00}) + "invalid"
		_, err := NewStorage(invalidPath)
		if err == nil {
			t.Error("Expected error for invalid directory")
		}
	})
}

func TestStorage_SaveAndLoadChain(t *testing.T) {
	tempStorage := NewTempDirStorage(t)
	defer tempStorage.Close()

	bc := NewBlockchainWithStorage(tempStorage, 2)

	// Добавляем блоки
	for i := 0; i < 3; i++ {
		data := CreateTestBlock(
			fmt.Sprintf("Author%d", i),
			fmt.Sprintf("Title%d", i),
			fmt.Sprintf("Text%d", i),
		)
		_, err := bc.AddBlock(data)
		AssertNoError(t, err)
	}

	originalLength := len(bc.Chain)

	// Сохраняем цепочку
	err := bc.saveChain()
	AssertNoError(t, err, "saveChain should not error")

	// Загружаем цепочку в новый блокчейн
	storage := tempStorage.GetStorage()
	loadedBC, err := storage.LoadChain()
	AssertNoError(t, err, "LoadChain should not error")
	AssertNotNil(t, loadedBC, "Loaded blockchain should not be nil")

	// Проверяем что длина совпадает
	AssertEqual(t, len(loadedBC.Chain), originalLength, "Chain length should match")

	// Проверяем что genesis блок совпадает
	AssertEqual(t, loadedBC.Chain[0].ID, "000-000-000", "Genesis ID should match")

	// Проверяем что последний блок совпадает
	lastOriginal := bc.Chain[len(bc.Chain)-1]
	lastLoaded := loadedBC.Chain[len(loadedBC.Chain)-1]
	AssertEqual(t, lastLoaded.ID, lastOriginal.ID, "Last block ID should match")
	AssertEqual(t, lastLoaded.Hash, lastOriginal.Hash, "Last block hash should match")
}

func TestStorage_SaveBlock(t *testing.T) {
	tempStorage := NewTempDirStorage(t)
	defer tempStorage.Close()

	block := &Block{
		ID:        "test-001",
		Timestamp: time.Now(),
		Data: DepositData{
			AuthorName:  "Test Author",
			Title:       "Test Title",
			ContentHash: "test-hash",
		},
		PrevHash: "prev-hash",
		Hash:     "current-hash",
		Nonce:    12345,
	}

	// Сохраняем блок
	err := tempStorage.SaveBlock(block)
	AssertNoError(t, err, "SaveBlock should not error")

	// Загружаем блок
	loaded, err := tempStorage.GetBlock("test-001")
	AssertNoError(t, err, "GetBlock should not error")
	AssertNotNil(t, loaded, "Loaded block should not be nil")

	// Проверяем данные
	AssertEqual(t, loaded.ID, block.ID, "Block ID")
	AssertEqual(t, loaded.Hash, block.Hash, "Block hash")
	AssertEqual(t, loaded.Data.AuthorName, block.Data.AuthorName, "Author name")
	AssertEqual(t, loaded.Nonce, block.Nonce, "Nonce")
}

func TestStorage_GetBlock_NotFound(t *testing.T) {
	tempStorage := NewTempDirStorage(t)
	defer tempStorage.Close()

	_, err := tempStorage.GetBlock("non-existent-id")
	AssertError(t, err, "Should error for non-existent block")
}

func TestStorage_GetAllBlocks(t *testing.T) {
	tempStorage := NewTempDirStorage(t)
	defer tempStorage.Close()

	// Сохраняем несколько блоков
	blocks := []*Block{
		{ID: "001", Timestamp: time.Now(), Hash: "hash1"},
		{ID: "002", Timestamp: time.Now(), Hash: "hash2"},
		{ID: "003", Timestamp: time.Now(), Hash: "hash3"},
	}

	for _, block := range blocks {
		err := tempStorage.SaveBlock(block)
		AssertNoError(t, err)
	}

	// Получаем все блоки
	allBlocks, err := tempStorage.GetAllBlocks()
	AssertNoError(t, err, "GetAllBlocks should not error")
	AssertEqual(t, len(allBlocks), 3, "Should return 3 blocks")
}

func TestStorage_CreateBackup(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "textproof-backup-test")
	defer os.RemoveAll(tempDir)

	storage, err := NewStorage(tempDir)
	AssertNoError(t, err)

	// Создаём блокчейн и сохраняем его
	bc := NewBlockchainWithStorage(NewTestStorage(), 1)
	data := CreateTestBlock("Author", "Title", "Text")
	bc.AddBlock(data)

	// Сохраняем в файл
	chainData := struct {
		Chain      []*Block `json:"chain"`
		Difficulty int      `json:"difficulty"`
	}{
		Chain:      bc.Chain,
		Difficulty: bc.Difficulty,
	}

	file, _ := os.Create(storage.chainFile)
	json.NewEncoder(file).Encode(chainData)
	file.Close()

	// Создаём бэкап
	err = storage.CreateBackup()
	AssertNoError(t, err, "CreateBackup should not error")

	// Проверяем что бэкап создан
	backupFiles, err := os.ReadDir(storage.backupDir)
	AssertNoError(t, err)
	if len(backupFiles) == 0 {
		t.Error("Backup file was not created")
	}
}

func TestStorage_CleanupOldBackups(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "textproof-cleanup-test")
	defer os.RemoveAll(tempDir)

	storage, err := NewStorage(tempDir)
	AssertNoError(t, err)

	// Создаём больше MaxBackups файлов
	for i := 0; i < 8; i++ {
		backupName := filepath.Join(storage.backupDir, fmt.Sprintf("backup_%d.json", i))
		err := os.WriteFile(backupName, []byte("test"), 0644)
		AssertNoError(t, err)
		time.Sleep(10 * time.Millisecond) // Чтобы время модификации различалось
	}

	// Запускаем cleanup
	err = storage.cleanupOldBackups()
	AssertNoError(t, err)

	// Проверяем что осталось только MaxBackups файлов
	files, _ := os.ReadDir(storage.backupDir)
	if len(files) > MaxBackups {
		t.Errorf("Too many backups: got %d, want max %d", len(files), MaxBackups)
	}
}

func TestStorage_WAL(t *testing.T) {
	tempDir := filepath.Join(os.TempDir(), "textproof-wal-test")
	defer os.RemoveAll(tempDir)

	storage, err := NewStorage(tempDir)
	AssertNoError(t, err)

	// Создаём блок
	block := &Block{
		ID:        "wal-001",
		Timestamp: time.Now(),
		Data:      DepositData{AuthorName: "WAL Test"},
		Hash:      "wal-hash",
	}

	// Записываем в WAL
	err = storage.WriteToWAL(block)
	AssertNoError(t, err, "WriteToWAL should not error")

	// Проверяем что файл WAL создан
	_, err = os.Stat(storage.walFile)
	AssertNoError(t, err, "WAL file should exist")

	// Очищаем WAL
	err = storage.ClearWAL()
	AssertNoError(t, err, "ClearWAL should not error")
}

func TestStorage_Concurrency(t *testing.T) {
	tempStorage := NewTempDirStorage(t)
	defer tempStorage.Close()

	const numGoroutines = 20
	done := make(chan bool, numGoroutines)

	// Параллельная запись блоков
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			block := &Block{
				ID:        fmt.Sprintf("concurrent-%d", id),
				Timestamp: time.Now(),
				Data:      DepositData{AuthorName: "Concurrent"},
				Hash:      fmt.Sprintf("hash-%d", id),
			}
			err := tempStorage.SaveBlock(block)
			if err != nil {
				t.Errorf("SaveBlock error in goroutine %d: %v", id, err)
			}
			done <- true
		}(i)
	}

	// Ждём завершения
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Проверяем что все блоки сохранены
	blocks, err := tempStorage.GetAllBlocks()
	AssertNoError(t, err)
	if len(blocks) != numGoroutines {
		t.Errorf("Expected %d blocks, got %d", numGoroutines, len(blocks))
	}
}

func BenchmarkStorage_SaveBlock(b *testing.B) {
	// Заменить на TestStorage вместо настоящего Storage
	storage := NewTestStorage()

	block := &Block{
		ID:        "bench-001",
		Timestamp: time.Now(),
		Data:      DepositData{AuthorName: "Benchmark"},
		Hash:      "bench-hash",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.SaveBlock(block)
	}
}
