package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"blockchain-verifier/internal/blockchain"
	"blockchain-verifier/internal/testutil"
	"blockchain-verifier/internal/viewmodels"
)

// Эти тесты для публичного JSON API (/api/v1/*)

func TestPublicAPI_DepositEndToEnd(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// 1. Депонируем текст
	reqData := viewmodels.DepositRequest{
		AuthorName: "John Doe",
		Title:      "Test Article",
		Text:       "This is a complete end-to-end test",
		PublicKey:  "test-key-123",
	}

	body := testutil.CreateJSONBody(t, reqData)
	req := testutil.HTTPTestRequest("POST", "/api/v1/deposit", body)
	resp := httptest.NewRecorder()

	api.handleDepositJSON(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusOK)

	var depositResp viewmodels.DepositResponsePublic
	testutil.ParseJSONResponse(t, resp, &depositResp)

	blockID := depositResp.BlockID
	testutil.AssertNotEqual(t, blockID, "", "block ID")
	testutil.AssertEqual(t, depositResp.Success, true, "should succeed")

	// 2. Проверяем что можем найти по ID
	verifyReq := viewmodels.VerifyByIDRequest{ID: blockID}
	body2 := testutil.CreateJSONBody(t, verifyReq)
	req2 := testutil.HTTPTestRequest("POST", "/api/v1/verify/id", body2)
	resp2 := httptest.NewRecorder()

	api.handleVerifyByIDJSON(resp2, req2)

	testutil.AssertStatusCode(t, resp2.Code, http.StatusOK)

	var verifyResp viewmodels.VerificationResponse
	testutil.ParseJSONResponse(t, resp2, &verifyResp)

	testutil.AssertEqual(t, verifyResp.Found, true, "found")
	testutil.AssertEqual(t, verifyResp.BlockID, blockID, "block ID")
	testutil.AssertEqual(t, verifyResp.Author, "John Doe", "author")

	// 3. Проверяем что можем найти по тексту
	verifyTextReq := viewmodels.VerifyByTextRequest{
		Text: "This is a complete end-to-end test",
	}
	body3 := testutil.CreateJSONBody(t, verifyTextReq)
	req3 := testutil.HTTPTestRequest("POST", "/api/v1/verify/text", body3)
	resp3 := httptest.NewRecorder()

	api.handleVerifyByTextJSON(resp3, req3)

	testutil.AssertStatusCode(t, resp3.Code, http.StatusOK)

	var verifyTextResp viewmodels.VerificationResponse
	testutil.ParseJSONResponse(t, resp3, &verifyTextResp)

	testutil.AssertEqual(t, verifyTextResp.Found, true, "found by text")
	testutil.AssertEqual(t, verifyTextResp.BlockID, blockID, "block ID matches")
}

func TestPublicAPI_ErrorResponses(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	tests := []struct {
		name       string
		endpoint   string
		method     string
		body       interface{}
		wantStatus int
	}{
		{
			name:       "deposit empty author",
			endpoint:   "/api/v1/deposit",
			method:     "POST",
			body:       viewmodels.DepositRequest{Title: "T", Text: "T"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "verify empty ID",
			endpoint:   "/api/v1/verify/id",
			method:     "POST",
			body:       viewmodels.VerifyByIDRequest{ID: ""},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "verify empty text",
			endpoint:   "/api/v1/verify/text",
			method:     "POST",
			body:       viewmodels.VerifyByTextRequest{Text: ""},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := testutil.CreateJSONBody(t, tt.body)
			req := testutil.HTTPTestRequest(tt.method, tt.endpoint, body)
			resp := httptest.NewRecorder()

			api.ServeHTTP(resp, req)

			testutil.AssertStatusCode(t, resp.Code, tt.wantStatus)

			var errorResp viewmodels.ErrorResponse
			testutil.ParseJSONResponse(t, resp, &errorResp)
			testutil.AssertNotEqual(t, errorResp.Error, "", "error message")
		})
	}
}

func TestPublicAPI_ContentTypeValidation(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	endpoints := []string{
		"/api/v1/deposit",
		"/api/v1/verify/id",
		"/api/v1/verify/text",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			// Запрос без Content-Type: application/json
			req := httptest.NewRequest("POST", endpoint, nil)
			req.Header.Set("Content-Type", "text/plain")
			resp := httptest.NewRecorder()

			api.ServeHTTP(resp, req)

			// Может быть ошибка парсинга или BadRequest
			if resp.Code == http.StatusOK {
				t.Log("API accepts non-JSON content type")
			}
		})
	}
}

func TestPublicAPI_Stats_Performance(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// Добавляем много блоков
	for i := 0; i < 100; i++ {
		data := blockchain.CreateTestBlock("Author", "Title", fmt.Sprintf("Text number %d", i))
		bc.AddBlock(data)
	}

	req := httptest.NewRequest("GET", "/api/v1/stats", nil)
	resp := httptest.NewRecorder()

	api.handleStats(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusOK)

	var stats viewmodels.StatsResponse
	testutil.ParseJSONResponse(t, resp, &stats)

	testutil.AssertEqual(t, stats.TotalBlocks, 101, "total blocks") // 100 + genesis
	if stats.UniqueAuthors < 1 {
		t.Error("Should have at least 1 unique author")
	}
}

func TestPublicAPI_BlockchainInfo_Detailed(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 3)
	api := NewAPI(bc)

	// Добавляем блоки
	for i := 0; i < 5; i++ {
		data := blockchain.CreateTestBlock("Author", "Title", fmt.Sprintf("Text %d", i))
		bc.AddBlock(data)
	}

	req := httptest.NewRequest("GET", "/api/v1/blockchain", nil)
	resp := httptest.NewRecorder()

	api.handleBlockchainInfo(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusOK)

	var info viewmodels.BlockchainInfoResponse
	testutil.ParseJSONResponse(t, resp, &info)

	testutil.AssertEqual(t, info.Length, 6, "length") // 5 + genesis
	testutil.AssertEqual(t, info.Difficulty, 3, "difficulty")
	testutil.AssertEqual(t, info.Valid, true, "valid")
	testutil.AssertNotEqual(t, info.LastBlock, "", "last block")
}

func TestPublicAPI_DuplicateHandling(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	reqData := viewmodels.DepositRequest{
		AuthorName: "Author",
		Title:      "Title",
		Text:       "Unique content for duplicate test",
	}

	// Первый запрос
	body1 := testutil.CreateJSONBody(t, reqData)
	req1 := testutil.HTTPTestRequest("POST", "/api/v1/deposit", body1)
	resp1 := httptest.NewRecorder()
	api.handleDepositJSON(resp1, req1)

	testutil.AssertStatusCode(t, resp1.Code, http.StatusOK)

	var resp1Data viewmodels.DepositResponsePublic
	testutil.ParseJSONResponse(t, resp1, &resp1Data)
	firstBlockID := resp1Data.BlockID

	// Второй запрос с тем же содержимым
	body2 := testutil.CreateJSONBody(t, reqData)
	req2 := testutil.HTTPTestRequest("POST", "/api/v1/deposit", body2)
	resp2 := httptest.NewRecorder()
	api.handleDepositJSON(resp2, req2)

	// Должна быть ошибка 409 Conflict
	testutil.AssertStatusCode(t, resp2.Code, http.StatusConflict)

	// Проверяем что в блокчейне только один блок
	blocks := bc.GetAllBlocks()
	if len(blocks) != 2 { // genesis + 1
		t.Errorf("Expected 2 blocks, got %d", len(blocks))
	}

	// Проверяем что можем найти оригинальный блок
	verifyReq := viewmodels.VerifyByIDRequest{ID: firstBlockID}
	body3 := testutil.CreateJSONBody(t, verifyReq)
	req3 := testutil.HTTPTestRequest("POST", "/api/v1/verify/id", body3)
	resp3 := httptest.NewRecorder()
	api.handleVerifyByIDJSON(resp3, req3)

	testutil.AssertStatusCode(t, resp3.Code, http.StatusOK)
}

func TestPublicAPI_VerifyNonExistent(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	t.Run("verify non-existent ID", func(t *testing.T) {
		reqData := viewmodels.VerifyByIDRequest{ID: "999-999-999"}
		body := testutil.CreateJSONBody(t, reqData)
		req := testutil.HTTPTestRequest("POST", "/api/v1/verify/id", body)
		resp := httptest.NewRecorder()

		api.handleVerifyByIDJSON(resp, req)

		testutil.AssertStatusCode(t, resp.Code, http.StatusOK)

		var verifyResp viewmodels.VerificationResponse
		testutil.ParseJSONResponse(t, resp, &verifyResp)

		testutil.AssertEqual(t, verifyResp.Found, false, "should not be found")
	})

	t.Run("verify non-existent text", func(t *testing.T) {
		reqData := viewmodels.VerifyByTextRequest{
			Text: "This text definitely does not exist in blockchain",
		}
		body := testutil.CreateJSONBody(t, reqData)
		req := testutil.HTTPTestRequest("POST", "/api/v1/verify/text", body)
		resp := httptest.NewRecorder()

		api.handleVerifyByTextJSON(resp, req)

		testutil.AssertStatusCode(t, resp.Code, http.StatusOK)

		var verifyResp viewmodels.VerificationResponse
		testutil.ParseJSONResponse(t, resp, &verifyResp)

		testutil.AssertEqual(t, verifyResp.Found, false, "should not be found")
	})
}

func BenchmarkPublicAPI_Deposit(b *testing.B) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	reqData := viewmodels.DepositRequest{
		AuthorName: "Benchmark Author",
		Title:      "Benchmark Title",
		Text:       "Benchmark text content",
	}

	body := testutil.CreateJSONBody(&testing.T{}, reqData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := testutil.HTTPTestRequest("POST", "/api/v1/deposit", body)
		resp := httptest.NewRecorder()
		api.handleDepositJSON(resp, req)
	}
}

func BenchmarkPublicAPI_Stats(b *testing.B) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// Добавляем данные
	for i := 0; i < 50; i++ {
		data := blockchain.CreateTestBlock("Author", "Title", "Text")
		bc.AddBlock(data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/stats", nil)
		resp := httptest.NewRecorder()
		api.handleStats(resp, req)
	}
}
