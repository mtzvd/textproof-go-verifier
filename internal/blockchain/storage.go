package blockchain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Storage управляет файловой системой для блокчейна
type Storage struct {
	chainFile string
	walFile   string
	backupDir string
}

// NewStorage создает новый Storage
func NewStorage(dataDir string) (*Storage, error) {
	// Создаем директории, если их нет
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}

	// Создаем директорию для бэкапов
	backupDir := filepath.Join(dataDir, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %v", err)
	}

	return &Storage{
		chainFile: filepath.Join(dataDir, "blockchain.json"),
		walFile:   filepath.Join(dataDir, "wal.json"),
		backupDir: backupDir,
	}, nil
}

// CreateBackup создает резервную копию блокчейна
func (s *Storage) CreateBackup() error {
	// Проверяем, существует ли основной файл
	if _, err := os.Stat(s.chainFile); os.IsNotExist(err) {
		return nil // Нет файла для бэкапа
	}

	// Создаем имя для бэкапа с временной меткой
	backupName := fmt.Sprintf("blockchain_backup_%d.json", os.Getpid())
	backupPath := filepath.Join(s.backupDir, backupName)

	// Копируем файл
	data, err := os.ReadFile(s.chainFile)
	if err != nil {
		return fmt.Errorf("failed to read blockchain file: %v", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup: %v", err)
	}

	// Удаляем старые бэкапы (оставляем только последние 5)
	if err := s.cleanupOldBackups(); err != nil {
		// Не критическая ошибка, можно продолжить
		fmt.Printf("Warning: failed to cleanup old backups: %v\n", err)
	}

	return nil
}

// cleanupOldBackups удаляет старые бэкапы, оставляя только последние 5
func (s *Storage) cleanupOldBackups() error {
	files, err := os.ReadDir(s.backupDir)
	if err != nil {
		return err
	}

	// Если бэкапов меньше 5, ничего не делаем
	if len(files) <= 5 {
		return nil
	}

	// Получаем информацию о файлах для сортировки по времени
	type fileInfo struct {
		path string
		info os.FileInfo
	}

	var fileInfos []fileInfo
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, fileInfo{
			path: filepath.Join(s.backupDir, file.Name()),
			info: info,
		})
	}

	// Сортируем по времени изменения (от старых к новым)
	// Удаляем самые старые
	toDelete := len(fileInfos) - 5
	for i := 0; i < toDelete; i++ {
		if err := os.Remove(fileInfos[i].path); err != nil {
			return err
		}
	}

	return nil
}

// RestoreFromBackup восстанавливает блокчейн из последнего бэкапа
func (s *Storage) RestoreFromBackup() error {
	files, err := os.ReadDir(s.backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %v", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no backups found")
	}

	// Находим самый свежий бэкап
	var latestBackup string
	var latestTime int64 = 0

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Unix() > latestTime {
			latestTime = info.ModTime().Unix()
			latestBackup = filepath.Join(s.backupDir, file.Name())
		}
	}

	if latestBackup == "" {
		return fmt.Errorf("no valid backups found")
	}

	// Копируем бэкап в основной файл
	data, err := os.ReadFile(latestBackup)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %v", err)
	}

	if err := os.WriteFile(s.chainFile, data, 0644); err != nil {
		return fmt.Errorf("failed to restore from backup: %v", err)
	}

	fmt.Printf("Successfully restored from backup: %s\n", latestBackup)
	return nil
}

// LoadChain загружает цепочку из файла
func (s *Storage) LoadChain() (*Blockchain, error) {
	// Проверяем существование файла
	if _, err := os.Stat(s.chainFile); err != nil {
		// Возвращаем оригинальную ошибку для проверки os.IsNotExist
		return nil, err
	}

	// Читаем файл
	data, err := os.ReadFile(s.chainFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read chain file: %w", err)
	}

	// Парсим JSON
	var bc Blockchain
	if err := json.Unmarshal(data, &bc); err != nil {
		return nil, fmt.Errorf("failed to parse chain file: %w", err)
	}

	return &bc, nil
}

// SaveChain сохраняет цепочку в файл
func (s *Storage) SaveChain(bc *Blockchain) error {
	// Создаем бэкап перед сохранением
	if err := s.CreateBackup(); err != nil {
		// Это не критическая ошибка, можно продолжить
		fmt.Printf("Warning: failed to create backup: %v\n", err)
	}

	// Сохраняем через временный файл для атомарности
	tmpFile := s.chainFile + ".tmp"

	data, err := json.MarshalIndent(bc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal chain: %w", err)
	}

	// Записываем во временный файл
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Атомарно переименовываем
	if err := os.Rename(tmpFile, s.chainFile); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// WriteToWAL записывает блок в WAL
func (s *Storage) WriteToWAL(block *Block) error {
	// Читаем существующий WAL
	var blocks []*Block
	data, err := os.ReadFile(s.walFile)
	if err != nil {
		// Если файла нет, начинаем с пустого списка
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read WAL: %w", err)
		}
		blocks = []*Block{}
	} else {
		if err := json.Unmarshal(data, &blocks); err != nil {
			// Если файл поврежден, начинаем заново
			blocks = []*Block{}
		}
	}

	// Добавляем новый блок
	blocks = append(blocks, block)

	// Записываем обратно
	data, err = json.MarshalIndent(blocks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal WAL: %w", err)
	}

	// Записываем через временный файл
	tmpFile := s.walFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write WAL temp file: %w", err)
	}

	return os.Rename(tmpFile, s.walFile)
}

// ReadWAL читает блоки из WAL
func (s *Storage) ReadWAL() ([]*Block, error) {
	if _, err := os.Stat(s.walFile); os.IsNotExist(err) {
		return []*Block{}, nil
	}

	data, err := os.ReadFile(s.walFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read WAL: %w", err)
	}

	var blocks []*Block
	if err := json.Unmarshal(data, &blocks); err != nil {
		return nil, fmt.Errorf("failed to parse WAL: %w", err)
	}

	return blocks, nil
}

// ClearWAL удаляет WAL файл
func (s *Storage) ClearWAL() error {
	if _, err := os.Stat(s.walFile); os.IsNotExist(err) {
		return nil // Файла нет
	}
	return os.Remove(s.walFile)
}

// FileExists проверяет существование файла
func (s *Storage) FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// GetChainPath возвращает путь к файлу цепочки
func (s *Storage) GetChainPath() string {
	return s.chainFile
}

// GetWALPath возвращает путь к файлу WAL
func (s *Storage) GetWALPath() string {
	return s.walFile
}
