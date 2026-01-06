package blockchain

import (
	"errors"
	"testing"
)

func TestBlockchainError(t *testing.T) {
	t.Run("create error", func(t *testing.T) {
		err := NewBlockchainError("TEST_CODE", "test message", nil)
		
		if err == nil {
			t.Fatal("NewBlockchainError() returned nil")
		}
		
		if err.Code != "TEST_CODE" {
			t.Errorf("Code = %s, want TEST_CODE", err.Code)
		}
		
		if err.Message != "test message" {
			t.Errorf("Message = %s, want 'test message'", err.Message)
		}
	})

	t.Run("error with cause", func(t *testing.T) {
		cause := errors.New("original error")
		err := NewBlockchainError("TEST_CODE", "wrapper message", cause)
		
		if err.Err != cause {
			t.Error("Cause not preserved")
		}
	})

	t.Run("error string", func(t *testing.T) {
		err := NewBlockchainError("TEST_CODE", "test message", nil)
		str := err.Error()
		
		if str == "" {
			t.Error("Error() returned empty string")
		}
		
		// Должен содержать код и сообщение
		if !contains(str, "TEST_CODE") {
			t.Errorf("Error string %q should contain code", str)
		}
		if !contains(str, "test message") {
			t.Errorf("Error string %q should contain message", str)
		}
	})

	t.Run("error unwrap", func(t *testing.T) {
		cause := errors.New("cause")
		err := NewBlockchainError("CODE", "message", cause)
		
		unwrapped := errors.Unwrap(err)
		if unwrapped != cause {
			t.Error("Unwrap() did not return original cause")
		}
	})
}

func TestBlockchainError_Codes(t *testing.T) {
	// Тестируем стандартные коды ошибок
	codes := []string{
		"GENESIS_CREATION_FAILED",
		"CHAIN_SAVE_FAILED",
		"CHAIN_LOAD_FAILED",
		"BLOCK_NOT_FOUND",
		"INVALID_BLOCK",
		"INVALID_RANGE",
	}

	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			err := NewBlockchainError(code, "test", nil)
			if err.Code != code {
				t.Errorf("Code = %s, want %s", err.Code, code)
			}
		})
	}
}

func TestDuplicateBlockError(t *testing.T) {
	err := &DuplicateBlockError{}
	
	if err.Error() == "" {
		t.Error("DuplicateBlockError.Error() returned empty string")
	}
	
	// Проверяем что это тип error
	var _ error = err
}

func TestBlockchainError_AsError(t *testing.T) {
	// Проверяем что BlockchainError реализует интерфейс error
	var err error = NewBlockchainError("CODE", "message", nil)
	
	if err == nil {
		t.Error("BlockchainError should implement error interface")
	}
}

func TestBlockchainError_ErrorsIs(t *testing.T) {
	// Тестируем работу с errors.Is
	cause := errors.New("original")
	err := NewBlockchainError("CODE", "message", cause)
	
	if !errors.Is(err, cause) {
		t.Error("errors.Is() should recognize wrapped error")
	}
}

func TestBlockchainError_ErrorsAs(t *testing.T) {
	// Тестируем работу с errors.As
	err := NewBlockchainError("CODE", "message", nil)
	
	var bcErr *BlockchainError
	if !errors.As(err, &bcErr) {
		t.Error("errors.As() should recognize BlockchainError")
	}
	
	if bcErr.Code != "CODE" {
		t.Errorf("Extracted error code = %s, want CODE", bcErr.Code)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
