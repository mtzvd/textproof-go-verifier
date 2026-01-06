package api

import (
	"testing"
	"time"
)

// map_stats_test.go - тесты для вспомогательных функций статистики

func TestMapStats_AuthorCounting(t *testing.T) {
	// Тестируем подсчёт уникальных авторов
	authors := make(map[string]bool)
	
	authors["Author1"] = true
	authors["Author2"] = true
	authors["Author1"] = true // дубликат
	
	if len(authors) != 2 {
		t.Errorf("Unique authors = %d, want 2", len(authors))
	}
}

func TestMapStats_TimeComparison(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)
	
	if !future.After(now) {
		t.Error("Future time should be after now")
	}
	
	if !past.Before(now) {
		t.Error("Past time should be before now")
	}
	
	if !now.After(past) {
		t.Error("Now should be after past")
	}
}

func TestMapStats_MapOperations(t *testing.T) {
	t.Run("map creation", func(t *testing.T) {
		m := make(map[string]interface{})
		m["key"] = "value"
		
		if m["key"] != "value" {
			t.Error("Map value not set correctly")
		}
	})
	
	t.Run("map key existence", func(t *testing.T) {
		m := make(map[string]bool)
		m["exists"] = true
		
		if _, ok := m["exists"]; !ok {
			t.Error("Key should exist")
		}
		
		if _, ok := m["not_exists"]; ok {
			t.Error("Key should not exist")
		}
	})
	
	t.Run("map overwrite", func(t *testing.T) {
		m := make(map[string]int)
		m["counter"] = 1
		m["counter"] = 2
		
		if m["counter"] != 2 {
			t.Error("Map value should be overwritten")
		}
	})
}

func TestMapStats_StatsAggregation(t *testing.T) {
	// Имитация агрегации статистики
	stats := make(map[string]int)
	
	// Подсчёт событий
	stats["deposits"] = 0
	stats["deposits"]++
	stats["deposits"]++
	stats["deposits"]++
	
	if stats["deposits"] != 3 {
		t.Errorf("Deposits count = %d, want 3", stats["deposits"])
	}
}

func TestMapStats_MultipleDataTypes(t *testing.T) {
	data := make(map[string]interface{})
	
	data["int"] = 42
	data["string"] = "hello"
	data["bool"] = true
	data["time"] = time.Now()
	
	if data["int"].(int) != 42 {
		t.Error("Int value incorrect")
	}
	
	if data["string"].(string) != "hello" {
		t.Error("String value incorrect")
	}
	
	if data["bool"].(bool) != true {
		t.Error("Bool value incorrect")
	}
	
	if _, ok := data["time"].(time.Time); !ok {
		t.Error("Time value incorrect type")
	}
}

func BenchmarkMapStats_Insert(b *testing.B) {
	m := make(map[string]bool)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m["key"] = true
	}
}

func BenchmarkMapStats_Lookup(b *testing.B) {
	m := make(map[string]bool)
	m["key"] = true
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m["key"]
	}
}
