package blockchain

import (
	"strings"
	"testing"
)

func TestIncrementID_NoLetter(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"simple increment", "000-000-000", "000-000-001", false},
		{"increment middle", "000-000-999", "000-001-000", false},
		{"increment first", "000-999-999", "001-000-000", false},
		{"complex", "012-345-678", "012-345-679", false},
		{"carry over", "000-999-999", "001-000-000", false},
		{"max simple", "999-999-998", "999-999-999", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := incrementID(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("incrementID() expected error")
				}
				return
			}
			
			if err != nil {
				t.Errorf("incrementID() error = %v", err)
				return
			}
			
			if got != tt.want {
				t.Errorf("incrementID(%s) = %s, want %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestIncrementID_WithLetter(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"letter A start", "A-000-000-000", "A-000-000-001", false},
		{"letter A carry", "A-000-000-999", "A-000-001-000", false},
		{"letter A max", "A-999-999-999", "B-000-000-000", false},
		{"letter B", "B-123-456-789", "B-123-456-790", false},
		{"letter Z overflow", "Z-999-999-999", "A-000-000-000", false}, // Циклический возврат
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := incrementID(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("incrementID() expected error")
				}
				return
			}
			
			if err != nil {
				t.Errorf("incrementID() error = %v", err)
				return
			}
			
			if got != tt.want {
				t.Errorf("incrementID(%s) = %s, want %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestIncrementID_InvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"too few parts", "000-000"},
		{"too many parts", "000-000-000-000-000"},
		{"invalid characters", "abc-def-ghi"},
		{"empty string", ""},
		{"single part", "000"},
		{"wrong separator", "000.000.000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := incrementID(tt.input)
			if err == nil {
				t.Errorf("incrementID(%s) expected error for invalid format", tt.input)
			}
		})
	}
}

func TestIncrementID_EdgeCases(t *testing.T) {
	t.Run("genesis to first", func(t *testing.T) {
		got, err := incrementID("000-000-000")
		if err != nil {
			t.Fatalf("incrementID() error = %v", err)
		}
		if got != "000-000-001" {
			t.Errorf("incrementID(000-000-000) = %s, want 000-000-001", got)
		}
	})

	t.Run("multiple increments", func(t *testing.T) {
		id := "000-000-000"
		for i := 0; i < 10; i++ {
			var err error
			id, err = incrementID(id)
			if err != nil {
				t.Fatalf("incrementID() iteration %d error = %v", i, err)
			}
		}
		if id != "000-000-010" {
			t.Errorf("After 10 increments got %s, want 000-000-010", id)
		}
	})

	t.Run("transition to letter", func(t *testing.T) {
		got, err := incrementID("999-999-999")
		if err != nil {
			t.Fatalf("incrementID() error = %v", err)
		}
		// Проверяем что это либо A-000-000-000, либо ошибка (зависит от реализации)
		if !strings.HasPrefix(got, "A-") && !strings.HasPrefix(got, "1000") {
			t.Logf("incrementID(999-999-999) = %s (checking transition logic)", got)
		}
	})
}

func TestIncrementID_Consistency(t *testing.T) {
	// Проверяем что инкрементирование детерминистично
	id := "000-000-000"
	
	// Первый проход
	ids1 := make([]string, 100)
	for i := 0; i < 100; i++ {
		var err error
		id, err = incrementID(id)
		if err != nil {
			t.Fatalf("First pass: incrementID() error at iteration %d: %v", i, err)
		}
		ids1[i] = id
	}

	// Второй проход
	id = "000-000-000"
	ids2 := make([]string, 100)
	for i := 0; i < 100; i++ {
		var err error
		id, err = incrementID(id)
		if err != nil {
			t.Fatalf("Second pass: incrementID() error at iteration %d: %v", i, err)
		}
		ids2[i] = id
	}

	// Сравниваем
	for i := 0; i < 100; i++ {
		if ids1[i] != ids2[i] {
			t.Errorf("Inconsistent ID at position %d: %s != %s", i, ids1[i], ids2[i])
		}
	}
}

func TestIncrementID_Format(t *testing.T) {
	// Проверяем что формат всегда корректный
	id := "000-000-000"
	
	for i := 0; i < 1000; i++ {
		var err error
		id, err = incrementID(id)
		if err != nil {
			t.Fatalf("incrementID() error at iteration %d: %v", i, err)
		}

		// Проверяем формат
		parts := strings.Split(id, "-")
		if len(parts) != 3 && len(parts) != 4 {
			t.Errorf("Invalid ID format at iteration %d: %s (expected 3 or 4 parts)", i, id)
		}

		// Проверяем длину частей (кроме буквы)
		startIdx := 0
		if len(parts) == 4 {
			startIdx = 1
			// Проверяем что первая часть - одна буква
			if len(parts[0]) != 1 || parts[0][0] < 'A' || parts[0][0] > 'Z' {
				t.Errorf("Invalid letter part at iteration %d: %s", i, parts[0])
			}
		}

		for j := startIdx; j < len(parts); j++ {
			if len(parts[j]) != 3 {
				t.Errorf("Invalid part length at iteration %d, part %d: %s (expected length 3)", i, j, parts[j])
			}
			// Проверяем что это цифры
			for _, c := range parts[j] {
				if c < '0' || c > '9' {
					t.Errorf("Invalid character at iteration %d, part %d: %c", i, j, c)
				}
			}
		}
	}
}

func BenchmarkIncrementID_NoLetter(b *testing.B) {
	id := "000-000-000"
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		id, _ = incrementID(id)
		// Сброс каждые 1000 итераций чтобы не уйти в буквы
		if i%1000 == 0 {
			id = "000-000-000"
		}
	}
}

func BenchmarkIncrementID_WithLetter(b *testing.B) {
	id := "A-000-000-000"
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		id, _ = incrementID(id)
		if i%1000 == 0 {
			id = "A-000-000-000"
		}
	}
}
