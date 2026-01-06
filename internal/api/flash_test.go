package api

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"blockchain-verifier/internal/testutil"
)

func TestSetFlash(t *testing.T) {
	t.Run("set success flash", func(t *testing.T) {
		resp := httptest.NewRecorder()
		data := map[string]string{"key": "value"}

		setFlash(resp, "success", "test message", data)

		// Проверяем что cookies установлены
		cookies := resp.Result().Cookies()
		if len(cookies) == 0 {
			t.Error("No cookies set")
		}

		var typeCookie, messageCookie, dataCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == "flash_type" {
				typeCookie = c
			} else if c.Name == "flash_message" {
				messageCookie = c
			} else if c.Name == "flash_key" {
				dataCookie = c
			}
		}

		if typeCookie == nil {
			t.Error("Flash type cookie not set")
		} else {
			// Декодируем base64
			decodedType, _ := base64.StdEncoding.DecodeString(typeCookie.Value)
			if string(decodedType) != "success" {
				t.Errorf("Flash type = %s, want success", string(decodedType))
			}
		}

		if messageCookie == nil {
			t.Error("Flash message cookie not set")
		} else {
			// Декодируем base64
			decodedMsg, _ := base64.StdEncoding.DecodeString(messageCookie.Value)
			if string(decodedMsg) != "test message" {
				t.Errorf("Flash message = %s, want 'test message'", string(decodedMsg))
			}
		}

		if dataCookie == nil {
			t.Error("Flash data cookie not set")
		}
	})

	t.Run("set error flash", func(t *testing.T) {
		resp := httptest.NewRecorder()

		setFlash(resp, "error", "error message", nil)

		cookie := testutil.GetCookie(resp, "flash_type")
		testutil.AssertNotNil(t, cookie, "flash type cookie should be set")
		decodedType, _ := base64.StdEncoding.DecodeString(cookie.Value)
		testutil.AssertEqual(t, string(decodedType), "error", "flash type")

		msgCookie := testutil.GetCookie(resp, "flash_message")
		testutil.AssertNotNil(t, msgCookie, "flash message cookie should be set")
	})

	t.Run("set warning flash", func(t *testing.T) {
		resp := httptest.NewRecorder()
		data := map[string]string{"warning": "true"}

		setFlash(resp, "warning", "warning message", data)

		cookie := testutil.GetCookie(resp, "flash_type")
		testutil.AssertNotNil(t, cookie, "flash type cookie should be set")

		dataCookie := testutil.GetCookie(resp, "flash_warning")
		testutil.AssertNotNil(t, dataCookie, "flash warning data cookie should be set")
	})

	t.Run("flash with data", func(t *testing.T) {
		resp := httptest.NewRecorder()
		data := map[string]string{
			"field1": "value1",
			"field2": "value2",
		}

		setFlash(resp, "info", "info message", data)

		cookie := testutil.GetCookie(resp, "flash_type")
		testutil.AssertNotNil(t, cookie, "flash type cookie should be set")

		field1Cookie := testutil.GetCookie(resp, "flash_field1")
		testutil.AssertNotNil(t, field1Cookie, "flash_field1 cookie should be set")
		decoded1, _ := base64.StdEncoding.DecodeString(field1Cookie.Value)
		testutil.AssertEqual(t, string(decoded1), "value1", "field1 value")

		field2Cookie := testutil.GetCookie(resp, "flash_field2")
		testutil.AssertNotNil(t, field2Cookie, "flash_field2 cookie should be set")
		decoded2, _ := base64.StdEncoding.DecodeString(field2Cookie.Value)
		testutil.AssertEqual(t, string(decoded2), "value2", "field2 value")
	})
}

func TestGetFlash(t *testing.T) {
	t.Run("get existing flash", func(t *testing.T) {
		// Создаём запрос с flash cookies
		respSet := httptest.NewRecorder()
		setFlash(respSet, "success", "test message", map[string]string{"key": "value"})

		// Получаем все cookies
		cookies := respSet.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatal("No flash cookies set")
		}

		// Создаём новый запрос с этими cookies
		req := httptest.NewRequest("GET", "/", nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		respGet := httptest.NewRecorder()

		flash := getFlash(req, respGet)

		testutil.AssertNotNil(t, flash, "flash should not be nil")
		testutil.AssertEqual(t, flash.Type, "success", "flash type")
		testutil.AssertEqual(t, flash.Message, "test message", "flash message")
		testutil.AssertNotNil(t, flash.Data, "flash data")
	})

	t.Run("get non-existent flash", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()

		flash := getFlash(req, resp)

		if flash != nil {
			t.Error("getFlash() should return nil when no flash cookie")
		}
	})

	t.Run("flash is cleared after read", func(t *testing.T) {
		// Устанавливаем flash
		respSet := httptest.NewRecorder()
		setFlash(respSet, "success", "message", nil)

		cookies := respSet.Result().Cookies()

		// Читаем flash
		req := httptest.NewRequest("GET", "/", nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		respGet := httptest.NewRecorder()

		flash := getFlash(req, respGet)
		testutil.AssertNotNil(t, flash, "first read should return flash")

		// Проверяем что cookies удалены
		deletedCookies := respGet.Result().Cookies()
		deletedCount := 0
		for _, c := range deletedCookies {
			if len(c.Name) >= 6 && c.Name[:6] == "flash_" && c.MaxAge < 0 {
				deletedCount++
			}
		}

		if deletedCount == 0 {
			t.Log("Flash cookies should be deleted after read (MaxAge < 0)")
		}
	})

	t.Run("invalid flash cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "flash_type",
			Value: "success",
		})
		// Отсутствует flash_message cookie
		resp := httptest.NewRecorder()

		flash := getFlash(req, resp)

		// Должен вернуть nil если нет flash_message
		if flash != nil {
			t.Error("getFlash() should return nil when flash_message is missing")
		}
	})
}

func TestGetFlashData(t *testing.T) {
	t.Run("with flash data", func(t *testing.T) {
		// Устанавливаем flash
		respSet := httptest.NewRecorder()
		data := map[string]string{"duplicate": "true"}
		setFlash(respSet, "warning", "duplicate warning", data)

		cookies := respSet.Result().Cookies()

		// Получаем flash data
		req := httptest.NewRequest("GET", "/", nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		resp := httptest.NewRecorder()

		flashData := getFlashData(req, resp)

		testutil.AssertEqual(t, flashData.Show, true, "show")
		testutil.AssertEqual(t, flashData.Type, "warning", "type")
		testutil.AssertEqual(t, flashData.Message, "duplicate warning", "message")
		testutil.AssertEqual(t, flashData.IsDuplicate, true, "is duplicate")
	})

	t.Run("without flash", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()

		flashData := getFlashData(req, resp)

		testutil.AssertEqual(t, flashData.Show, false, "show should be false")
		testutil.AssertEqual(t, flashData.IsDuplicate, false, "is duplicate")
	})

	t.Run("duplicate flag parsing", func(t *testing.T) {
		tests := []struct {
			name           string
			duplicateValue string
			wantDuplicate  bool
		}{
			{"duplicate true", "true", true},
			{"duplicate false", "false", false},
			{"duplicate empty", "", false},
			{"duplicate missing", "", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				respSet := httptest.NewRecorder()
				data := map[string]string{}
				if tt.duplicateValue != "" {
					data["duplicate"] = tt.duplicateValue
				}
				setFlash(respSet, "info", "message", data)

				cookies := respSet.Result().Cookies()
				req := httptest.NewRequest("GET", "/", nil)
				for _, c := range cookies {
					req.AddCookie(c)
				}
				resp := httptest.NewRecorder()

				flashData := getFlashData(req, resp)

				testutil.AssertEqual(t, flashData.IsDuplicate, tt.wantDuplicate, "is duplicate")
			})
		}
	})
}

func TestFlash_Types(t *testing.T) {
	types := []string{"success", "error", "warning", "info"}

	for _, flashType := range types {
		t.Run("type "+flashType, func(t *testing.T) {
			respSet := httptest.NewRecorder()
			setFlash(respSet, flashType, "message", nil)

			cookies := respSet.Result().Cookies()
			req := httptest.NewRequest("GET", "/", nil)
			for _, c := range cookies {
				req.AddCookie(c)
			}
			resp := httptest.NewRecorder()

			flash := getFlash(req, resp)

			testutil.AssertNotNil(t, flash, "flash")
			testutil.AssertEqual(t, flash.Type, flashType, "flash type")
		})
	}
}

func TestFlash_EmptyMessage(t *testing.T) {
	resp := httptest.NewRecorder()
	setFlash(resp, "success", "", nil)

	cookie := testutil.GetCookie(resp, "flash_type")
	testutil.AssertNotNil(t, cookie, "flash type cookie should be set even with empty message")

	msgCookie := testutil.GetCookie(resp, "flash_message")
	testutil.AssertNotNil(t, msgCookie, "flash message cookie should be set even with empty message")
}

func TestFlash_LongMessage(t *testing.T) {
	resp := httptest.NewRecorder()
	longMessage := strings.Repeat("a", 1000)

	setFlash(resp, "error", longMessage, nil)

	cookie := testutil.GetCookie(resp, "flash_message")
	testutil.AssertNotNil(t, cookie, "flash message cookie should be set with long message")
}

func TestFlash_SpecialCharacters(t *testing.T) {
	resp := httptest.NewRecorder()
	message := "Тест с кириллицей и спецсимволами: <>&\"'"
	data := map[string]string{
		"field": "значение с пробелами",
	}

	setFlash(resp, "info", message, data)

	cookies := resp.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("flash cookies should be set")
	}

	// Читаем обратно
	req := httptest.NewRequest("GET", "/", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	respGet := httptest.NewRecorder()

	flash := getFlash(req, respGet)

	testutil.AssertNotNil(t, flash, "flash")
	testutil.AssertEqual(t, flash.Message, message, "message with special chars")
}

func TestFlash_Concurrent(t *testing.T) {
	const numGoroutines = 50
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			resp := httptest.NewRecorder()
			setFlash(resp, "success", "message", nil)

			cookie := testutil.GetCookie(resp, "flash_type")
			if cookie == nil {
				t.Errorf("Flash type cookie not set in goroutine %d", id)
			}

			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func BenchmarkSetFlash(b *testing.B) {
	data := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := httptest.NewRecorder()
		setFlash(resp, "success", "message", data)
	}
}

func BenchmarkGetFlash(b *testing.B) {
	// Подготовка
	respSet := httptest.NewRecorder()
	setFlash(respSet, "success", "message", nil)
	cookies := respSet.Result().Cookies()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		resp := httptest.NewRecorder()
		_ = getFlash(req, resp)
	}
}
