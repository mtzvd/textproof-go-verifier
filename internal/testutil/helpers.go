package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

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

// HTTPTestRequest создает HTTP запрос для тестирования
func HTTPTestRequest(method, url string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, url, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

// HTTPTestFormRequest создает HTTP form запрос для тестирования
func HTTPTestFormRequest(method, urlPath string, formData map[string]string) *http.Request {
	form := url.Values{}
	for key, value := range formData {
		form.Set(key, value)
	}

	req := httptest.NewRequest(method, urlPath, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

// ParseJSONResponse парсит JSON ответ
func ParseJSONResponse(t *testing.T, resp *httptest.ResponseRecorder, target interface{}) {
	t.Helper()

	err := json.NewDecoder(resp.Body).Decode(target)
	if err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}
}

// CreateJSONBody создает JSON body для запроса
func CreateJSONBody(t *testing.T, data interface{}) *bytes.Buffer {
	t.Helper()

	body, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	return bytes.NewBuffer(body)
}

// ResponseRecorderWithCookies создает ResponseRecorder с поддержкой cookies
func ResponseRecorderWithCookies() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// GetCookie получает cookie из response
func GetCookie(resp *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, cookie := range resp.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}
