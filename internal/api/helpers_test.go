package api

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendJSON(t *testing.T) {
	api := &API{}

	t.Run("send struct", func(t *testing.T) {
		resp := httptest.NewRecorder()
		data := map[string]string{"key": "value"}

		api.sendJSON(resp, http.StatusOK, data)

		if resp.Code != http.StatusOK {
			t.Errorf("Status code = %d, want %d", resp.Code, http.StatusOK)
		}

		contentType := resp.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Content-Type = %s, want application/json", contentType)
		}

		if resp.Body.Len() == 0 {
			t.Error("Response body is empty")
		}
	})

	t.Run("send empty data", func(t *testing.T) {
		resp := httptest.NewRecorder()
		api.sendJSON(resp, http.StatusOK, nil)

		if resp.Code != http.StatusOK {
			t.Errorf("Status code = %d, want %d", resp.Code, http.StatusOK)
		}
	})

	t.Run("send different status codes", func(t *testing.T) {
		statuses := []int{
			http.StatusOK,
			http.StatusCreated,
			http.StatusBadRequest,
			http.StatusNotFound,
			http.StatusInternalServerError,
		}

		for _, status := range statuses {
			resp := httptest.NewRecorder()
			api.sendJSON(resp, status, map[string]string{})

			if resp.Code != status {
				t.Errorf("Status code = %d, want %d", resp.Code, status)
			}
		}
	})
}

func TestSendError(t *testing.T) {
	api := &API{}

	t.Run("error without details", func(t *testing.T) {
		resp := httptest.NewRecorder()
		api.sendError(resp, http.StatusBadRequest, "test error", nil)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Status code = %d, want %d", resp.Code, http.StatusBadRequest)
		}

		contentType := resp.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Content-Type = %s, want application/json", contentType)
		}

		// Проверяем что это JSON с error полем
		if !contains(resp.Body.String(), "test error") {
			t.Error("Response should contain error message")
		}
	})

	t.Run("error with details", func(t *testing.T) {
		resp := httptest.NewRecorder()
		innerErr := errors.New("inner error")
		api.sendError(resp, http.StatusInternalServerError, "outer error", innerErr)

		if !contains(resp.Body.String(), "inner error") {
			t.Error("Response should contain inner error details")
		}
	})

	t.Run("different error codes", func(t *testing.T) {
		codes := []int{
			http.StatusBadRequest,
			http.StatusUnauthorized,
			http.StatusNotFound,
			http.StatusConflict,
			http.StatusInternalServerError,
		}

		for _, code := range codes {
			resp := httptest.NewRecorder()
			api.sendError(resp, code, "error", nil)

			if resp.Code != code {
				t.Errorf("Status code = %d, want %d", resp.Code, code)
			}
		}
	})
}

func TestGetBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		tls      bool
		proto    string
		wantHTTP bool
	}{
		{"http default", "localhost:8080", false, "", true},
		{"https with TLS", "example.com", true, "", false},
		{"https with header", "example.com", false, "https", false},
		{"custom host", "192.168.1.1:3000", false, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Host = tt.host

			if tt.tls {
				req.TLS = &tls.ConnectionState{}
			}

			if tt.proto != "" {
				req.Header.Set("X-Forwarded-Proto", tt.proto)
			}

			result := getBaseURL(req)

			if tt.wantHTTP && !strings.HasPrefix(result, "http://") {
				t.Errorf("getBaseURL() = %s, want http:// prefix", result)
			}

			if !tt.wantHTTP && !strings.HasPrefix(result, "https://") {
				t.Errorf("getBaseURL() = %s, want https:// prefix", result)
			}

			if !contains(result, tt.host) {
				t.Errorf("getBaseURL() = %s, should contain host %s", result, tt.host)
			}
		})
	}
}

func TestExtractTextFragment(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		n         int
		fromStart bool
		want      string
	}{
		{
			name:      "first 3 words",
			text:      "one two three four five",
			n:         3,
			fromStart: true,
			want:      "one two three",
		},
		{
			name:      "last 3 words",
			text:      "one two three four five",
			n:         3,
			fromStart: false,
			want:      "three four five",
		},
		{
			name:      "fewer words than n",
			text:      "one two",
			n:         5,
			fromStart: true,
			want:      "one two",
		},
		{
			name:      "empty text",
			text:      "",
			n:         3,
			fromStart: true,
			want:      "",
		},
		{
			name:      "single word",
			text:      "word",
			n:         1,
			fromStart: true,
			want:      "word",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTextFragment(tt.text, tt.n, tt.fromStart)
			if got != tt.want {
				t.Errorf("extractTextFragment() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderHTML(t *testing.T) {
	// Это сложно тестировать без реальных templ компонентов
	// Но можно проверить основные вещи
	t.Skip("Skipping renderHTML test - requires templ components")
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
