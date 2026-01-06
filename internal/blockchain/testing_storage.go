package blockchain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// TestStorage - in-memory хранилище для тестов
type TestStorage struct {
	mu     sync.RWMutex
	blocks map[string]*Block
	wal    []string // имитация WAL
}

// NewTestStorage создает новое тестовое хранилище в памяти
func NewTestStorage() *TestStorage {
	return &TestStorage{
		blocks: make(map[string]*Block),
		wal:    make([]string, 0),
	}
}

func (s *TestStorage) SaveBlock(block *Block) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Имитация записи в WAL
	data, err := json.Marshal(block)
	if err != nil {
		return err
	}
	s.wal = append(s.wal, string(data))

	// Сохранение в память
	s.blocks[block.ID] = block
	return nil
}

func (s *TestStorage) GetBlock(id string) (*Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	block, exists := s.blocks[id]
	if !exists {
		return nil, fmt.Errorf("block not found: %s", id)
	}
	return block, nil
}

func (s *TestStorage) GetAllBlocks() ([]*Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	blocks := make([]*Block, 0, len(s.blocks))
	for _, block := range s.blocks {
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (s *TestStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.blocks = make(map[string]*Block)
	s.wal = make([]string, 0)
	return nil
}

// TempDirStorage - хранилище в временной директории для тестов
type TempDirStorage struct {
	Dir     string
	storage *Storage
	blocks  map[string]*Block
	mu      sync.RWMutex
	t       *testing.T
}

// NewTempDirStorage создает временную директорию и storage для тестов
func NewTempDirStorage(t *testing.T) *TempDirStorage {
	t.Helper()

	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("textproof-test-%d", os.Getpid()))
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	storage, err := NewStorage(tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create storage: %v", err)
	}

	return &TempDirStorage{
		Dir:     tempDir,
		storage: storage,
		blocks:  make(map[string]*Block),
		t:       t,
	}
}

func (s *TempDirStorage) SaveBlock(block *Block) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Сохраняем в in-memory map
	s.blocks[block.ID] = block
	return nil
}

func (s *TempDirStorage) GetBlock(id string) (*Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	block, exists := s.blocks[id]
	if !exists {
		return nil, fmt.Errorf("block not found: %s", id)
	}
	return block, nil
}

func (s *TempDirStorage) GetAllBlocks() ([]*Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	blocks := make([]*Block, 0, len(s.blocks))
	for _, block := range s.blocks {
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (s *TempDirStorage) Close() error {
	err := os.RemoveAll(s.Dir)
	if err != nil {
		s.t.Logf("Warning: temp dir cleanup error: %v", err)
	}
	return nil
}

// GetStorage возвращает storage для работы с цепочкой (SaveChain/LoadChain)
func (s *TempDirStorage) GetStorage() *Storage {
	return s.storage
}

// MockStorage - мок для тестирования ошибок
type MockStorage struct {
	SaveError     error
	GetError      error
	GetAllError   error
	blocks        map[string]*Block
	SaveCallCount int
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		blocks: make(map[string]*Block),
	}
}

func (s *MockStorage) SaveBlock(block *Block) error {
	s.SaveCallCount++
	if s.SaveError != nil {
		return s.SaveError
	}
	s.blocks[block.ID] = block
	return nil
}

func (s *MockStorage) GetBlock(id string) (*Block, error) {
	if s.GetError != nil {
		return nil, s.GetError
	}
	block, exists := s.blocks[id]
	if !exists {
		return nil, fmt.Errorf("block not found")
	}
	return block, nil
}

func (s *MockStorage) GetAllBlocks() ([]*Block, error) {
	if s.GetAllError != nil {
		return nil, s.GetAllError
	}
	blocks := make([]*Block, 0, len(s.blocks))
	for _, block := range s.blocks {
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func (s *MockStorage) Close() error {
	return nil
}
