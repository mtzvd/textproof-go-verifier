package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"blockchain-verifier/internal/blockchain"
	"blockchain-verifier/internal/testutil"
	"blockchain-verifier/internal/viewmodels"
)

func TestHandleDeposit_Success(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	formData := map[string]string{
		"author_name": "John Doe",
		"title":       "Test Article",
		"text":        "This is a test text content for blockchain deposit",
		"public_key":  "test-public-key-123",
	}

	req := testutil.HTTPTestFormRequest("POST", "/api/deposit", formData)
	resp := httptest.NewRecorder()

	api.handleDeposit(resp, req)

	// Проверяем редирект
	testutil.AssertStatusCode(t, resp.Code, http.StatusSeeOther)

	// Проверяем что блок добавлен
	blocks := bc.GetAllBlocks()
	if len(blocks) != 2 { // genesis + новый
		t.Errorf("Expected 2 blocks, got %d", len(blocks))
	}

	// Проверяем данные блока
	lastBlock := blocks[len(blocks)-1]
	testutil.AssertEqual(t, lastBlock.Data.AuthorName, "John Doe", "author name")
	testutil.AssertEqual(t, lastBlock.Data.Title, "Test Article", "title")
}

func TestHandleDeposit_EmptyFields(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	tests := []struct {
		name      string
		formData  map[string]string
		wantError string
	}{
		{
			name: "empty author",
			formData: map[string]string{
				"author_name": "",
				"title":       "Title",
				"text":        "Text",
			},
			wantError: "имя автора не может быть пустым",
		},
		{
			name: "empty title",
			formData: map[string]string{
				"author_name": "Author",
				"title":       "",
				"text":        "Text",
			},
			wantError: "название не может быть пустым",
		},
		{
			name: "empty text",
			formData: map[string]string{
				"author_name": "Author",
				"title":       "Title",
				"text":        "",
			},
			wantError: "текст не может быть пустым",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := testutil.HTTPTestFormRequest("POST", "/api/deposit", tt.formData)
			resp := httptest.NewRecorder()

			api.handleDeposit(resp, req)

			// Должна быть ошибка (400)
			testutil.AssertStatusCode(t, resp.Code, http.StatusBadRequest)
		})
	}
}

func TestHandleDeposit_TooLong(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// Создаём очень длинный текст
	longText := make([]byte, MaxTextLength+1000)
	for i := range longText {
		longText[i] = 'a'
	}

	formData := map[string]string{
		"author_name": "Author",
		"title":       "Title",
		"text":        string(longText),
	}

	req := testutil.HTTPTestFormRequest("POST", "/api/deposit", formData)
	resp := httptest.NewRecorder()

	api.handleDeposit(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusBadRequest)
}

func TestHandleDeposit_Duplicate(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	formData := map[string]string{
		"author_name": "Author",
		"title":       "Title",
		"text":        "Unique text content",
	}

	// Первое добавление
	req1 := testutil.HTTPTestFormRequest("POST", "/api/deposit", formData)
	resp1 := httptest.NewRecorder()
	api.handleDeposit(resp1, req1)

	testutil.AssertStatusCode(t, resp1.Code, http.StatusSeeOther)

	// Второе добавление того же текста
	req2 := testutil.HTTPTestFormRequest("POST", "/api/deposit", formData)
	resp2 := httptest.NewRecorder()
	api.handleDeposit(resp2, req2)

	// Должен быть редирект (дубликаты разрешены, но будет flash)
	testutil.AssertStatusCode(t, resp2.Code, http.StatusSeeOther)

	// Проверяем что блоков только 2 (genesis + 1)
	blocks := bc.GetAllBlocks()
	if len(blocks) != 2 {
		t.Errorf("Expected 2 blocks for duplicate, got %d", len(blocks))
	}
}

func TestHandleDeposit_InvalidFormData(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// Запрос без Content-Type
	req := httptest.NewRequest("POST", "/api/deposit", nil)
	resp := httptest.NewRecorder()

	api.handleDeposit(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusBadRequest)
}

func TestHandleDepositJSON_Success(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	reqData := viewmodels.DepositRequest{
		AuthorName: "John Doe",
		Title:      "Test Article",
		Text:       "This is test content for JSON API",
		PublicKey:  "json-public-key",
	}

	body := testutil.CreateJSONBody(t, reqData)
	req := testutil.HTTPTestRequest("POST", "/api/v1/deposit", body)
	resp := httptest.NewRecorder()

	api.handleDepositJSON(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusOK)

	var result viewmodels.DepositResponsePublic
	testutil.ParseJSONResponse(t, resp, &result)

	testutil.AssertNotEqual(t, result.BlockID, "", "block ID should not be empty")
	testutil.AssertNotEqual(t, result.Hash, "", "hash should not be empty")
	testutil.AssertEqual(t, result.Success, true, "should succeed")
}

func TestHandleDepositJSON_InvalidJSON(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	req := testutil.HTTPTestRequest("POST", "/api/v1/deposit",
		testutil.CreateJSONBody(t, "invalid json"))
	resp := httptest.NewRecorder()

	api.handleDepositJSON(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusBadRequest)
}

func TestHandleDepositJSON_Duplicate(t *testing.T) {
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

	// Второй запрос с тем же содержимым
	body2 := testutil.CreateJSONBody(t, reqData)
	req2 := testutil.HTTPTestRequest("POST", "/api/v1/deposit", body2)
	resp2 := httptest.NewRecorder()
	api.handleDepositJSON(resp2, req2)

	// Должна быть ошибка 409 Conflict
	testutil.AssertStatusCode(t, resp2.Code, http.StatusConflict)
}

func TestHandleDepositPage(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	req := httptest.NewRequest("GET", "/deposit", nil)
	resp := httptest.NewRecorder()

	api.handleDepositPage(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusOK)
	testutil.AssertContains(t, resp.Header().Get("Content-Type"), "text/html", "content type")
}

func TestValidateDepositRequest(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	tests := []struct {
		name    string
		req     viewmodels.DepositRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: viewmodels.DepositRequest{
				AuthorName: "John",
				Title:      "Title",
				Text:       "Content",
			},
			wantErr: false,
		},
		{
			name: "empty author",
			req: viewmodels.DepositRequest{
				AuthorName: "",
				Title:      "Title",
				Text:       "Content",
			},
			wantErr: true,
		},
		{
			name: "too long author",
			req: viewmodels.DepositRequest{
				AuthorName: string(make([]byte, MaxAuthorLength+1)),
				Title:      "Title",
				Text:       "Content",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := api.validateDepositRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDepositRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkHandleDeposit(b *testing.B) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	formData := map[string]string{
		"author_name": "Benchmark Author",
		"title":       "Benchmark Title",
		"text":        "Benchmark text content",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := testutil.HTTPTestFormRequest("POST", "/api/deposit", formData)
		resp := httptest.NewRecorder()
		api.handleDeposit(resp, req)
	}
}
