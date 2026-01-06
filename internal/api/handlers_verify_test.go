package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"blockchain-verifier/internal/blockchain"
	"blockchain-verifier/internal/testutil"
	"blockchain-verifier/internal/viewmodels"

	"github.com/gorilla/mux"
)

func TestHandleVerifyByIDJSON_Success(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// Добавляем блок
	data := blockchain.CreateTestBlock("Author", "Title", "Test content")
	block, _ := bc.AddBlock(data)

	// Запрос на верификацию
	reqData := viewmodels.VerifyByIDRequest{
		ID: block.ID,
	}

	body := testutil.CreateJSONBody(t, reqData)
	req := testutil.HTTPTestRequest("POST", "/api/v1/verify/id", body)
	resp := httptest.NewRecorder()

	api.handleVerifyByIDJSON(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusOK)

	var result viewmodels.VerificationResponse
	testutil.ParseJSONResponse(t, resp, &result)

	testutil.AssertEqual(t, result.Found, true, "found")
	testutil.AssertEqual(t, result.BlockID, block.ID, "block ID")
	testutil.AssertEqual(t, result.Author, "Author", "author")
}

func TestHandleVerifyByIDJSON_NotFound(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	reqData := viewmodels.VerifyByIDRequest{
		ID: "999-999-999",
	}

	body := testutil.CreateJSONBody(t, reqData)
	req := testutil.HTTPTestRequest("POST", "/api/v1/verify/id", body)
	resp := httptest.NewRecorder()

	api.handleVerifyByIDJSON(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusOK)

	var result viewmodels.VerificationResponse
	testutil.ParseJSONResponse(t, resp, &result)

	testutil.AssertEqual(t, result.Found, false, "should not be found")
}

func TestHandleVerifyByIDJSON_EmptyID(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	reqData := viewmodels.VerifyByIDRequest{
		ID: "",
	}

	body := testutil.CreateJSONBody(t, reqData)
	req := testutil.HTTPTestRequest("POST", "/api/v1/verify/id", body)
	resp := httptest.NewRecorder()

	api.handleVerifyByIDJSON(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusBadRequest)
}

func TestHandleVerifyByTextJSON_Success(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// Добавляем блок с известным текстом
	text := "This is a unique test content for verification"
	data := blockchain.CreateTestBlock("Author", "Title", text)
	block, _ := bc.AddBlock(data)

	// Запрос на верификацию по тексту
	reqData := viewmodels.VerifyByTextRequest{
		Text: text,
	}

	body := testutil.CreateJSONBody(t, reqData)
	req := testutil.HTTPTestRequest("POST", "/api/v1/verify/text", body)
	resp := httptest.NewRecorder()

	api.handleVerifyByTextJSON(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusOK)

	var result viewmodels.VerificationResponse
	testutil.ParseJSONResponse(t, resp, &result)

	testutil.AssertEqual(t, result.Found, true, "found")
	testutil.AssertEqual(t, result.BlockID, block.ID, "block ID")
	testutil.AssertEqual(t, result.Matches, true, "matches")
}

func TestHandleVerifyByTextJSON_NotFound(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	reqData := viewmodels.VerifyByTextRequest{
		Text: "This text does not exist in blockchain",
	}

	body := testutil.CreateJSONBody(t, reqData)
	req := testutil.HTTPTestRequest("POST", "/api/v1/verify/text", body)
	resp := httptest.NewRecorder()

	api.handleVerifyByTextJSON(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusOK)

	var result viewmodels.VerificationResponse
	testutil.ParseJSONResponse(t, resp, &result)

	testutil.AssertEqual(t, result.Found, false, "should not be found")
}

func TestHandleVerifyByTextJSON_EmptyText(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	reqData := viewmodels.VerifyByTextRequest{
		Text: "",
	}

	body := testutil.CreateJSONBody(t, reqData)
	req := testutil.HTTPTestRequest("POST", "/api/v1/verify/text", body)
	resp := httptest.NewRecorder()

	api.handleVerifyByTextJSON(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusBadRequest)
}

func TestHandleVerifyByTextJSON_TooLong(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// Создаём слишком длинный текст
	longText := make([]byte, MaxTextLength+1000)
	for i := range longText {
		longText[i] = 'a'
	}

	reqData := viewmodels.VerifyByTextRequest{
		Text: string(longText),
	}

	body := testutil.CreateJSONBody(t, reqData)
	req := testutil.HTTPTestRequest("POST", "/api/v1/verify/text", body)
	resp := httptest.NewRecorder()

	api.handleVerifyByTextJSON(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusBadRequest)
}

func TestHandleVerifyByIDSubmit(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// Добавляем блок
	data := blockchain.CreateTestBlock("Author", "Title", "Text")
	block, _ := bc.AddBlock(data)

	formData := map[string]string{
		"id": block.ID,
	}

	req := testutil.HTTPTestFormRequest("POST", "/api/verify/id", formData)
	resp := httptest.NewRecorder()

	api.handleVerifyByIDSubmit(resp, req)

	// Должен быть редирект
	testutil.AssertStatusCode(t, resp.Code, http.StatusSeeOther)
}

func TestHandleVerifyByTextSubmit(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// Добавляем блок
	text := "Test verification text"
	data := blockchain.CreateTestBlock("Author", "Title", text)
	bc.AddBlock(data)

	formData := map[string]string{
		"text": text,
	}

	req := testutil.HTTPTestFormRequest("POST", "/api/verify/text", formData)
	resp := httptest.NewRecorder()

	api.handleVerifyByTextSubmit(resp, req)

	// Должен быть редирект
	testutil.AssertStatusCode(t, resp.Code, http.StatusSeeOther)
}

func TestHandleVerifyDirectLink(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// Добавляем блок
	data := blockchain.CreateTestBlock("Author", "Title", "Text")
	block, _ := bc.AddBlock(data)

	req := httptest.NewRequest("GET", "/verify/"+block.ID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": block.ID})
	resp := httptest.NewRecorder()

	api.handleVerifyDirectLink(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusOK)
	testutil.AssertContains(t, resp.Header().Get("Content-Type"), "text/html", "content type")
}

func TestHandleVerifyDirectLink_NotFound(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	req := httptest.NewRequest("GET", "/verify/999-999-999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999-999-999"})
	resp := httptest.NewRecorder()

	api.handleVerifyDirectLink(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusOK)
	// Страница должна отобразиться с ошибкой
}

func TestHandleVerifyPage(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	req := httptest.NewRequest("GET", "/verify", nil)
	resp := httptest.NewRecorder()

	api.handleVerifyPage(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusOK)
	testutil.AssertContains(t, resp.Header().Get("Content-Type"), "text/html", "content type")
}

func BenchmarkHandleVerifyByIDJSON(b *testing.B) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// Добавляем блоки для поиска
	for i := 0; i < 100; i++ {
		data := blockchain.CreateTestBlock("Author", "Title", "Text")
		bc.AddBlock(data)
	}

	reqData := viewmodels.VerifyByIDRequest{ID: "000-000-001"}
	body := testutil.CreateJSONBody(&testing.T{}, reqData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := testutil.HTTPTestRequest("POST", "/api/v1/verify/id", body)
		resp := httptest.NewRecorder()
		api.handleVerifyByIDJSON(resp, req)
	}
}

func BenchmarkHandleVerifyByTextJSON(b *testing.B) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// Добавляем блоки
	text := "Benchmark text content"
	data := blockchain.CreateTestBlock("Author", "Title", text)
	bc.AddBlock(data)

	reqData := viewmodels.VerifyByTextRequest{Text: text}
	body := testutil.CreateJSONBody(&testing.T{}, reqData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := testutil.HTTPTestRequest("POST", "/api/v1/verify/text", body)
		resp := httptest.NewRecorder()
		api.handleVerifyByTextJSON(resp, req)
	}
}
