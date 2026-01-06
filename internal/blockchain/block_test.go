package blockchain

import (
	"strings"
	"testing"
	"time"
)

func TestBlock_CalculateHash(t *testing.T) {
	tests := []struct {
		name  string
		block *Block
	}{
		{
			name: "valid block with data",
			block: &Block{
				ID:        "test-001",
				Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Data: DepositData{
					AuthorName:  "Test Author",
					Title:       "Test Title",
					ContentHash: "abc123",
				},
				PrevHash: "previous-hash",
				Nonce:    0,
			},
		},
		{
			name: "genesis block",
			block: &Block{
				ID:        "000-000-000",
				Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Data:      DepositData{},
				PrevHash:  "",
				Nonce:     0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := tt.block.CalculateHash()

			// Hash должен быть не пустым
			if hash == "" {
				t.Error("CalculateHash() returned empty string")
			}

			// Hash должен быть hex строкой длиной 64 символа (SHA256)
			if len(hash) != 64 {
				t.Errorf("CalculateHash() hash length = %d, want 64", len(hash))
			}

			// Проверяем что это hex строка
			for _, c := range hash {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Errorf("CalculateHash() returned non-hex character: %c", c)
				}
			}

			// Повторный вызов должен вернуть тот же hash
			hash2 := tt.block.CalculateHash()
			if hash != hash2 {
				t.Error("CalculateHash() is not deterministic")
			}
		})
	}
}

func TestBlock_CalculateHash_DifferentInputs(t *testing.T) {
	// Блоки с разными данными должны иметь разные хеши
	block1 := &Block{
		ID:        "test-001",
		Timestamp: time.Now(),
		Data: DepositData{
			AuthorName:  "Author 1",
			Title:       "Title 1",
			ContentHash: "hash1",
		},
		PrevHash: "prev1",
		Nonce:    0,
	}

	block2 := &Block{
		ID:        "test-002",
		Timestamp: time.Now(),
		Data: DepositData{
			AuthorName:  "Author 2",
			Title:       "Title 2",
			ContentHash: "hash2",
		},
		PrevHash: "prev2",
		Nonce:    0,
	}

	hash1 := block1.CalculateHash()
	hash2 := block2.CalculateHash()

	if hash1 == hash2 {
		t.Error("Different blocks produced same hash")
	}
}

func TestBlock_Mine(t *testing.T) {
	tests := []struct {
		name       string
		difficulty int
		wantPrefix string
	}{
		{
			name:       "difficulty 1",
			difficulty: 1,
			wantPrefix: "0",
		},
		{
			name:       "difficulty 2",
			difficulty: 2,
			wantPrefix: "00",
		},
		{
			name:       "difficulty 3",
			difficulty: 3,
			wantPrefix: "000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := &Block{
				ID:        "test-001",
				Timestamp: time.Now(),
				Data: DepositData{
					AuthorName:  "Test Author",
					Title:       "Test Title",
					ContentHash: "abc123",
				},
				PrevHash: "",
				Nonce:    0,
			}

			block.Mine(tt.difficulty)

			// Проверяем что hash начинается с нужного количества нулей
			if !strings.HasPrefix(block.Hash, tt.wantPrefix) {
				t.Errorf("Mine() hash = %s, want prefix %s", block.Hash, tt.wantPrefix)
			}

			// Проверяем что nonce изменился
			if block.Nonce == 0 {
				t.Error("Mine() did not change nonce")
			}

			// Проверяем что hash действительно правильный
			calculatedHash := block.CalculateHash()
			if block.Hash != calculatedHash {
				t.Error("Mine() stored hash doesn't match calculated hash")
			}
		})
	}
}

func TestBlock_Mine_Difficulty0(t *testing.T) {
	// При difficulty 0 майнинг должен пройти быстро
	block := &Block{
		ID:        "test-001",
		Timestamp: time.Now(),
		Data:      DepositData{AuthorName: "Test"},
		PrevHash:  "",
		Nonce:     0,
	}

	start := time.Now()
	block.Mine(0)
	duration := time.Since(start)

	// Проверяем что майнинг завершился быстро (< 100ms)
	if duration > 100*time.Millisecond {
		t.Errorf("Mine(0) took too long: %v", duration)
	}

	// Hash должен быть установлен
	if block.Hash == "" {
		t.Error("Mine(0) did not set hash")
	}
}

func TestBlock_Validation(t *testing.T) {
	t.Run("valid block structure", func(t *testing.T) {
		block := &Block{
			ID:        "test-001",
			Timestamp: time.Now(),
			Data: DepositData{
				AuthorName:  "Test Author",
				Title:       "Test Title",
				TextStart:   "Start of text",
				TextEnd:     "End of text",
				ContentHash: "abc123def456",
				PublicKey:   "public-key-123",
			},
			PrevHash: "previous-hash",
			Nonce:    12345,
			Hash:     "current-hash",
		}

		// Проверяем что все поля заполнены корректно
		if block.ID == "" {
			t.Error("Block ID is empty")
		}
		if block.Data.AuthorName == "" {
			t.Error("Block AuthorName is empty")
		}
		if block.Data.ContentHash == "" {
			t.Error("Block ContentHash is empty")
		}
	})

	t.Run("empty block", func(t *testing.T) {
		block := &Block{}
		hash := block.CalculateHash()

		// Даже пустой блок должен иметь hash
		if hash == "" {
			t.Error("Empty block CalculateHash() returned empty string")
		}
	})
}

func TestDepositData_Completeness(t *testing.T) {
	data := DepositData{
		AuthorName:  "John Doe",
		Title:       "My Article",
		TextStart:   "This is the beginning",
		TextEnd:     "This is the end",
		ContentHash: "sha256hash",
		PublicKey:   "pubkey123",
	}

	// Проверяем что все поля можно записать и прочитать
	if data.AuthorName != "John Doe" {
		t.Error("AuthorName not stored correctly")
	}
	if data.Title != "My Article" {
		t.Error("Title not stored correctly")
	}
	if data.TextStart != "This is the beginning" {
		t.Error("TextStart not stored correctly")
	}
	if data.TextEnd != "This is the end" {
		t.Error("TextEnd not stored correctly")
	}
	if data.ContentHash != "sha256hash" {
		t.Error("ContentHash not stored correctly")
	}
	if data.PublicKey != "pubkey123" {
		t.Error("PublicKey not stored correctly")
	}
}

func BenchmarkBlock_CalculateHash(b *testing.B) {
	block := &Block{
		ID:        "benchmark-block",
		Timestamp: time.Now(),
		Data: DepositData{
			AuthorName:  "Benchmark Author",
			Title:       "Benchmark Title",
			ContentHash: "benchmark-hash",
		},
		PrevHash: "previous-hash",
		Nonce:    0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = block.CalculateHash()
	}
}

func BenchmarkBlock_Mine_Difficulty2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		block := &Block{
			ID:        "benchmark-block",
			Timestamp: time.Now(),
			Data: DepositData{
				AuthorName:  "Benchmark Author",
				Title:       "Benchmark Title",
				ContentHash: "benchmark-hash",
			},
			PrevHash: "previous-hash",
			Nonce:    0,
		}
		block.Mine(2)
	}
}
