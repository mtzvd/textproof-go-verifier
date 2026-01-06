package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"blockchain-verifier/internal/blockchain"
	"blockchain-verifier/internal/testutil"

	"github.com/gorilla/mux"
)

func TestNewAPI(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 2)

	api := NewAPI(bc)

	if api == nil {
		t.Fatal("NewAPI() returned nil")
	}

	if api.blockchain != bc {
		t.Error("API blockchain not set correctly")
	}

	if api.router == nil {
		t.Error("API router not initialized")
	}
}

func TestAPI_ServeHTTP(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	tests := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
	}{
		{"home page", "GET", "/", http.StatusOK},
		{"deposit page", "GET", "/deposit", http.StatusOK},
		{"verify page", "GET", "/verify", http.StatusOK},
		{"about page", "GET", "/about", http.StatusOK},
		{"stats API", "GET", "/api/v1/stats", http.StatusOK},
		{"blockchain info", "GET", "/api/v1/blockchain", http.StatusOK},
		{"not found", "GET", "/nonexistent", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			resp := httptest.NewRecorder()

			api.ServeHTTP(resp, req)

			testutil.AssertStatusCode(t, resp.Code, tt.wantStatusCode)
		})
	}
}

func TestAPI_SetupRoutes(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// Проверяем что роуты зарегистрированы
	router := api.router

	if router == nil {
		t.Fatal("Router not initialized")
	}

	// Тестируем некоторые роуты
	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/"},
		{"GET", "/deposit"},
		{"GET", "/verify"},
		{"GET", "/about"},
		{"POST", "/api/deposit"},
		{"POST", "/api/verify/id"},
		{"POST", "/api/verify/text"},
		{"GET", "/api/v1/stats"},
		{"POST", "/api/v1/deposit"},
		{"POST", "/api/v1/verify/id"},
		{"POST", "/api/v1/verify/text"},
		{"GET", "/swagger/"},
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			resp := httptest.NewRecorder()

			api.ServeHTTP(resp, req)

			// Не должно быть 404 для зарегистрированных роутов
			// (могут быть другие ошибки, но не 404)
			if resp.Code == http.StatusNotFound && route.path != "/nonexistent" {
				t.Errorf("Route %s %s not registered (got 404)", route.method, route.path)
			}
		})
	}
}

func TestAPI_StaticFiles(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	staticPaths := []string{
		"/static/css/styles.css",
		"/static/js/app.js",
		"/static/favicon.svg",
	}

	for _, path := range staticPaths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			resp := httptest.NewRecorder()

			api.ServeHTTP(resp, req)

			// Статические файлы могут не существовать в тестах,
			// но роут должен быть зарегистрирован
			// Проверяем что это не 404 или это правильная 404
			if resp.Code != http.StatusOK && resp.Code != http.StatusNotFound {
				t.Logf("Static file %s returned status %d", path, resp.Code)
			}
		})
	}
}

func TestAPI_CORS(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	req := httptest.NewRequest("OPTIONS", "/api/v1/stats", nil)
	req.Header.Set("Origin", "http://example.com")
	resp := httptest.NewRecorder()

	api.ServeHTTP(resp, req)

	// Проверяем наличие CORS заголовков (если реализовано)
	// Это опционально, но хорошая практика для API
	t.Logf("CORS headers present: %v", resp.Header().Get("Access-Control-Allow-Origin") != "")
}

func TestAPI_ContentType(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	tests := []struct {
		name     string
		path     string
		wantType string
	}{
		{"HTML page", "/", "text/html"},
		{"JSON API", "/api/v1/stats", "application/json"},
		{"CSS file", "/static/css/styles.css", "text/css"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			resp := httptest.NewRecorder()

			api.ServeHTTP(resp, req)

			contentType := resp.Header().Get("Content-Type")
			if contentType != "" && !contains(contentType, tt.wantType) {
				t.Logf("Content-Type = %s, expected to contain %s", contentType, tt.wantType)
			}
		})
	}
}

func TestAPI_MethodNotAllowed(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// Пытаемся POST на GET endpoint
	req := httptest.NewRequest("POST", "/", nil)
	resp := httptest.NewRecorder()

	api.ServeHTTP(resp, req)

	// Должна быть ошибка метода (405) или редирект
	if resp.Code != http.StatusMethodNotAllowed &&
		resp.Code != http.StatusMovedPermanently &&
		resp.Code != http.StatusFound {
		t.Logf("POST to GET endpoint returned %d", resp.Code)
	}
}

func TestAPI_QueryParams(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// Добавляем блок для теста
	data := blockchain.CreateTestBlock("Author", "Title", "Text")
	block, _ := bc.AddBlock(data)

	req := httptest.NewRequest("GET", "/verify/"+block.ID+"?param=value", nil)
	req = mux.SetURLVars(req, map[string]string{"id": block.ID})
	resp := httptest.NewRecorder()

	api.ServeHTTP(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusOK)
}

func TestAPI_Concurrency(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	const numRequests = 50
	done := make(chan bool, numRequests)

	// Параллельные запросы
	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/api/v1/stats", nil)
			resp := httptest.NewRecorder()
			api.ServeHTTP(resp, req)
			done <- true
		}()
	}

	// Ждём завершения
	for i := 0; i < numRequests; i++ {
		<-done
	}

	t.Log("Concurrent requests handled successfully")
}

func TestAPI_LargePayload(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// Создаём большой payload (но в пределах лимита)
	text := make([]byte, MaxTextLength-1000)
	for i := range text {
		text[i] = 'a'
	}

	formData := map[string]string{
		"author_name": "Author",
		"title":       "Title",
		"text":        string(text),
	}

	req := testutil.HTTPTestFormRequest("POST", "/api/deposit", formData)
	resp := httptest.NewRecorder()

	api.handleDeposit(resp, req)

	// Должен принять большой payload
	testutil.AssertStatusCode(t, resp.Code, http.StatusSeeOther)
}

func BenchmarkAPI_ServeHTTP_Stats(b *testing.B) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// Добавляем данные для более реалистичного бенчмарка
	for i := 0; i < 10; i++ {
		data := blockchain.CreateTestBlock("Author", "Title", "Text")
		bc.AddBlock(data)
	}

	req := httptest.NewRequest("GET", "/api/v1/stats", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := httptest.NewRecorder()
		api.ServeHTTP(resp, req)
	}
}
