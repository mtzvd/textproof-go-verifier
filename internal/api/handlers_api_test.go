package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"blockchain-verifier/internal/blockchain"
	"blockchain-verifier/internal/testutil"
	"blockchain-verifier/internal/viewmodels"
)

func TestAPI_HandleStats(t *testing.T) {
	// Создаём тестовый блокчейн
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 2)
	api := NewAPI(bc)

	// Добавляем тестовые блоки с уникальным текстом
	for i := 0; i < 3; i++ {
		data := blockchain.CreateTestBlock(
			"Author"+string(rune('A'+i)),
			"Title"+string(rune('0'+i)),
			"Text number "+string(rune('0'+i)),
		)
		bc.AddBlock(data)
	}

	// Создаём запрос
	req := httptest.NewRequest("GET", "/api/v1/stats", nil)
	resp := httptest.NewRecorder()

	// Выполняем запрос
	api.handleStats(resp, req)

	// Проверяем статус код
	testutil.AssertStatusCode(t, resp.Code, http.StatusOK)

	// Парсим ответ
	var stats viewmodels.StatsResponse
	testutil.ParseJSONResponse(t, resp, &stats)

	// Проверяем данные
	testutil.AssertEqual(t, stats.TotalBlocks, 4, "total blocks") // genesis + 3
	testutil.AssertEqual(t, stats.UniqueAuthors, 4, "unique authors") // genesis author + 3 test authors
	testutil.AssertEqual(t, stats.ChainValid, true, "chain valid")
}

func TestAPI_HandleBlockchainInfo(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 3)
	api := NewAPI(bc)

	// Добавляем блок
	data := blockchain.CreateTestBlock("Author", "Title", "Text")
	bc.AddBlock(data)

	// Запрос
	req := httptest.NewRequest("GET", "/api/v1/blockchain", nil)
	resp := httptest.NewRecorder()

	api.handleBlockchainInfo(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusOK)

	var info viewmodels.BlockchainInfoResponse
	testutil.ParseJSONResponse(t, resp, &info)

	testutil.AssertEqual(t, info.Length, 2, "chain length")
	testutil.AssertEqual(t, info.Difficulty, 3, "difficulty")
	testutil.AssertEqual(t, info.Valid, true, "valid")
	testutil.AssertNotEqual(t, info.LastBlock, "", "last block ID")
}

func TestAPI_HandleQRCode(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	// Добавляем блок
	data := blockchain.CreateTestBlock("Author", "Title", "Text")
	block, _ := bc.AddBlock(data)

	// Создаём запрос с ID в URL
	req := httptest.NewRequest("GET", "/api/v1/qrcode/"+block.ID, nil)
	resp := httptest.NewRecorder()

	// Добавляем переменные маршрута (для mux.Vars)
	req = mux.SetURLVars(req, map[string]string{"id": block.ID})

	api.handleQRCode(resp, req)

	testutil.AssertStatusCode(t, resp.Code, http.StatusOK)

	// Проверяем Content-Type
	contentType := resp.Header().Get("Content-Type")
	testutil.AssertEqual(t, contentType, "image/png", "content type")

	// Проверяем что есть данные
	if resp.Body.Len() == 0 {
		t.Error("QR code body is empty")
	}
}

func TestAPI_HandleQRCode_NotFound(t *testing.T) {
	storage := blockchain.NewTestStorage()
	bc := blockchain.NewBlockchainWithStorage(storage, 1)
	api := NewAPI(bc)

	req := httptest.NewRequest("GET", "/api/v1/qrcode/999-999-999", nil)
	resp := httptest.NewRecorder()

	req = mux.SetURLVars(req, map[string]string{"id": "999-999-999"})

	api.handleQRCode(resp, req)

	// Должна быть ошибка
	testutil.AssertStatusCode(t, resp.Code, http.StatusNotFound)
}
