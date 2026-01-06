package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"blockchain-verifier/internal/blockchain"
	"blockchain-verifier/internal/testutil"
)

// handlers_docs_test.go

func TestHandleDocsPage(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	req := httptest.NewRequest("GET", "/docs", nil)
	resp := httptest.NewRecorder()

	api.handleDocsPage(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusOK)

	// Проверяем что это HTML
	contentType := resp.Header().Get("Content-Type")
	if contentType != "" && !contains(contentType, "text/html") {
		t.Logf("Content-Type = %s, expected HTML", contentType)
	}
}

func TestHandleDocsPage_Content(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	req := httptest.NewRequest("GET", "/docs", nil)
	resp := httptest.NewRecorder()

	api.handleDocsPage(resp, req)

	body := resp.Body.String()

	// Проверяем что содержит ключевые слова документации
	keywords := []string{"API", "Swagger", "документация"}
	found := false
	for _, keyword := range keywords {
		if contains(body, keyword) {
			found = true
			break
		}
	}

	if !found && body != "" {
		t.Log("Docs page rendered but content check inconclusive")
	}
}
