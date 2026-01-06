package blockchain

import (
	"strings"
	"testing"
)

// Вспомогательные функции для тестирования

// NewBlockchainWithStorage создает новый блокчейн для тестов
// Использует временную директорию и реальный blockchain.NewBlockchain

// CreateTestBlock создает тестовый блок с данными

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// AssertEqual проверяет равенство двух значений
func AssertEqual(t *testing.T, got, want interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
		if len(msgAndArgs) > 0 {
			t.Errorf("Context: %v", msgAndArgs...)
		}
	}
}

// AssertNotEqual проверяет неравенство двух значений
func AssertNotEqual(t *testing.T, got, notWant interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if got == notWant {
		t.Errorf("got %v, did not want %v", got, notWant)
		if len(msgAndArgs) > 0 {
			t.Errorf("Context: %v", msgAndArgs...)
		}
	}
}

// AssertNil проверяет что значение nil
func AssertNil(t *testing.T, got interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if got != nil {
		t.Errorf("expected nil, got %v", got)
		if len(msgAndArgs) > 0 {
			t.Errorf("Context: %v", msgAndArgs...)
		}
	}
}

// AssertNotNil проверяет что значение не nil
func AssertNotNil(t *testing.T, got interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if got == nil {
		t.Errorf("expected non-nil value")
		if len(msgAndArgs) > 0 {
			t.Errorf("Context: %v", msgAndArgs...)
		}
	}
}

// AssertError проверяет что есть ошибка
func AssertError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err == nil {
		t.Errorf("expected error, got nil")
		if len(msgAndArgs) > 0 {
			t.Errorf("Context: %v", msgAndArgs...)
		}
	}
}

// AssertNoError проверяет что нет ошибки
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		if len(msgAndArgs) > 0 {
			t.Errorf("Context: %v", msgAndArgs...)
		}
	}
}

// AssertContains проверяет что строка содержит подстроку
func AssertContains(t *testing.T, haystack, needle string, msgAndArgs ...interface{}) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("expected %q to contain %q", haystack, needle)
		if len(msgAndArgs) > 0 {
			t.Errorf("Context: %v", msgAndArgs...)
		}
	}
}

// AssertStatusCode проверяет HTTP status code
func AssertStatusCode(t *testing.T, got, want int, msgAndArgs ...interface{}) {
	t.Helper()
	if got != want {
		t.Errorf("status code: got %d, want %d", got, want)
		if len(msgAndArgs) > 0 {
			t.Errorf("Context: %v", msgAndArgs...)
		}
	}
}
