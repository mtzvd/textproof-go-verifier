package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"net/http"
	"strings"
	"time"

	"blockchain-verifier/internal/blockchain"
	"blockchain-verifier/internal/viewmodels"
	"blockchain-verifier/web/templates"

	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"github.com/skip2/go-qrcode"
)

// API представляет собой HTTP API сервер
type API struct {
	blockchain *blockchain.Blockchain
	router     *mux.Router
}

// NewAPI создает новый экземпляр API
func NewAPI(bc *blockchain.Blockchain) *API {
	api := &API{
		blockchain: bc,
		router:     mux.NewRouter(),
	}

	api.setupRoutes()
	return api
}

// setupRoutes настраивает маршруты API
func (api *API) setupRoutes() {
	// Web UI routes
	api.router.HandleFunc("/", api.handleHome).Methods("GET")
	api.router.HandleFunc("/deposit", api.handleDepositPage).Methods("GET")
	api.router.HandleFunc("/verify", api.handleVerifyPage).Methods("GET")
	api.router.HandleFunc("/verify/{id}", api.handleVerifyPage).Methods("GET")
	api.router.HandleFunc("/test", api.handleTest).Methods("GET")

	// API routes
	api.router.HandleFunc("/api/deposit", api.handleDeposit).Methods("POST")
	api.router.HandleFunc("/api/verify/id/{id}", api.handleVerifyByID).Methods("GET")
	api.router.HandleFunc("/api/verify/text", api.handleVerifyByText).Methods("POST")
	api.router.HandleFunc("/api/stats", api.handleStats).Methods("GET")
	api.router.HandleFunc("/api/blockchain", api.handleBlockchainInfo).Methods("GET")
	api.router.HandleFunc("/api/qrcode/{id}", api.handleQRCode).Methods("GET")
	api.router.HandleFunc("/api/badge/{id}", api.handleBadge).Methods("GET")

	// Static files
	fs := http.FileServer(http.Dir("web/static"))
	api.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// 404 handler
	api.router.NotFoundHandler = http.HandlerFunc(api.handleNotFound)
}

func (api *API) handleTest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("<!DOCTYPE html><html><body><h1>Test HTML</h1></body></html>"))
}

// ServeHTTP реализует интерфейс http.Handler
func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.router.ServeHTTP(w, r)
}

// renderHTML упрощает рендеринг HTML шаблонов
func (api *API) renderHTML(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(r.Context(), w); err != nil {
		log.Printf("Render error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Тогда обработчики будут выглядеть так:
func (api *API) handleHome(w http.ResponseWriter, r *http.Request) {
	allBlocks := api.blockchain.GetAllBlocks()

	authors := make(map[string]bool)
	var lastAdded time.Time

	for _, block := range allBlocks {
		authors[block.Data.AuthorName] = true
		if block.Timestamp.After(lastAdded) {
			lastAdded = block.Timestamp
		}
	}

	stats := viewmodels.StatsResponse{
		TotalBlocks:   len(allBlocks),
		UniqueAuthors: len(authors),
		LastAdded:     lastAdded,
		ChainValid:    api.blockchain.ValidateChain(),
	}

	api.renderHTML(w, r, templates.Home(stats))
}

func (api *API) handleDepositPage(w http.ResponseWriter, r *http.Request) {
	api.renderHTML(w, r, templates.Deposit())
}

func (api *API) handleVerifyPage(w http.ResponseWriter, r *http.Request) {
	api.renderHTML(w, r, templates.Verify())
}

// handleDeposit обрабатывает запрос на депонирование текста
func (api *API) handleDeposit(w http.ResponseWriter, r *http.Request) {
	var req viewmodels.DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.sendError(w, http.StatusBadRequest, "Неверный формат запроса", err)
		return
	}

	// Валидация
	if req.AuthorName == "" || req.Title == "" || req.Text == "" {
		api.sendError(w, http.StatusBadRequest, "Все обязательные поля должны быть заполнены", nil)
		return
	}

	// Вычисляем хеш текста
	hash := sha256.Sum256([]byte(req.Text))
	contentHash := hex.EncodeToString(hash[:])

	// Извлекаем фрагменты текста (первые и последние 2-3 слова)
	textStart := extractTextFragment(req.Text, 3, true)
	textEnd := extractTextFragment(req.Text, 3, false)

	// Создаем данные для блока
	data := blockchain.DepositData{
		AuthorName:  req.AuthorName,
		Title:       req.Title,
		TextStart:   textStart,
		TextEnd:     textEnd,
		ContentHash: contentHash,
		PublicKey:   req.PublicKey,
	}

	// Добавляем блок в цепочку
	block, err := api.blockchain.AddBlock(data)
	if err != nil {
		api.sendError(w, http.StatusInternalServerError, "Не удалось добавить блок", err)
		return
	}

	// Генерируем URL для QR-кода и бейджа
	baseURL := getBaseURL(r)
	qrcodeURL := fmt.Sprintf("%s/api/qrcode/%s", baseURL, block.ID)
	badgeURL := fmt.Sprintf("%s/api/badge/%s", baseURL, block.ID)

	// Формируем ответ
	resp := viewmodels.DepositResponse{
		ID:        block.ID,
		Hash:      block.Hash,
		Timestamp: block.Timestamp,
		QRCodeURL: qrcodeURL,
		BadgeURL:  badgeURL,
	}

	api.sendJSON(w, http.StatusCreated, resp)
}

// handleVerifyByID обрабатывает проверку по ID
func (api *API) handleVerifyByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	block, err := api.blockchain.GetBlockByID(id)
	if err != nil {
		api.sendError(w, http.StatusNotFound, "Блок не найден", err)
		return
	}

	resp := viewmodels.VerificationResponse{
		Found:     true,
		BlockID:   block.ID,
		Author:    block.Data.AuthorName,
		Title:     block.Data.Title,
		Timestamp: block.Timestamp,
		Hash:      block.Data.ContentHash,
		Matches:   true, // Если блок найден, то хеш уже совпадает (мы не проверяем текст, только что блок существует)
	}

	api.sendJSON(w, http.StatusOK, resp)
}

// handleVerifyByText обрабатывает проверку по тексту
func (api *API) handleVerifyByText(w http.ResponseWriter, r *http.Request) {
	var req viewmodels.VerifyByTextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.sendError(w, http.StatusBadRequest, "Неверный формат запроса", err)
		return
	}

	// Вычисляем хеш текста
	hash := sha256.Sum256([]byte(req.Text))
	contentHash := hex.EncodeToString(hash[:])

	// Ищем блок с таким хешем
	var foundBlock *blockchain.Block
	allBlocks := api.blockchain.GetAllBlocks()
	for _, block := range allBlocks {
		if block.Data.ContentHash == contentHash {
			foundBlock = block
			break
		}
	}

	if foundBlock == nil {
		resp := viewmodels.VerificationResponse{
			Found: false,
		}
		api.sendJSON(w, http.StatusOK, resp)
		return
	}

	resp := viewmodels.VerificationResponse{
		Found:     true,
		BlockID:   foundBlock.ID,
		Author:    foundBlock.Data.AuthorName,
		Title:     foundBlock.Data.Title,
		Timestamp: foundBlock.Timestamp,
		Hash:      foundBlock.Data.ContentHash,
		Matches:   true,
	}

	api.sendJSON(w, http.StatusOK, resp)
}

// handleStats обрабатывает запрос статистики
func (api *API) handleStats(w http.ResponseWriter, r *http.Request) {
	allBlocks := api.blockchain.GetAllBlocks()

	// Считаем уникальных авторов
	authors := make(map[string]bool)
	var lastAdded time.Time

	for _, block := range allBlocks {
		authors[block.Data.AuthorName] = true
		if block.Timestamp.After(lastAdded) {
			lastAdded = block.Timestamp
		}
	}

	resp := viewmodels.StatsResponse{
		TotalBlocks:   len(allBlocks),
		UniqueAuthors: len(authors),
		LastAdded:     lastAdded,
		ChainValid:    api.blockchain.ValidateChain(),
	}

	api.sendJSON(w, http.StatusOK, resp)
}

// handleBlockchainInfo обрабатывает запрос информации о блокчейне
func (api *API) handleBlockchainInfo(w http.ResponseWriter, r *http.Request) {
	info := api.blockchain.GetChainInfo()

	resp := viewmodels.BlockchainInfoResponse{
		Length:     info["length"].(int),
		Difficulty: info["difficulty"].(int),
		Valid:      info["valid"].(bool),
	}

	if lastBlock, ok := info["last_block"]; ok {
		resp.LastBlock = lastBlock.(string)
	}

	api.sendJSON(w, http.StatusOK, resp)
}

// handleQRCode генерирует QR-код для ID
func (api *API) handleQRCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Создаем URL для проверки
	baseURL := getBaseURL(r)
	verifyURL := fmt.Sprintf("%s/verify/%s", baseURL, id)

	// Генерируем QR-код
	png, err := qrcode.Encode(verifyURL, qrcode.Medium, 256)
	if err != nil {
		api.sendError(w, http.StatusInternalServerError, "Не удалось сгенерировать QR-код", err)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
}

// handleBadge генерирует HTML-бейдж для вставки на сайт
func (api *API) handleBadge(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Ищем блок
	block, err := api.blockchain.GetBlockByID(id)
	if err != nil {
		api.sendError(w, http.StatusNotFound, "Блок не найден", err)
		return
	}

	// Генерируем HTML бейдж
	html := fmt.Sprintf(`
	<div style="display: inline-block; padding: 10px; border: 1px solid #ddd; border-radius: 5px; font-family: sans-serif; max-width: 300px; text-align: center;">
		<p style="margin: 0 0 10px 0; font-weight: bold;">Автор: %s</p>
		<p style="margin: 0 0 10px 0; font-size: 0.9em;">ID: %s</p>
		<p style="margin: 0; font-size: 0.8em; color: #666;">Зафиксировано: %s</p>
	</div>
	`, block.Data.AuthorName, block.ID, block.Timestamp.Format("02.01.2006"))

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// handleNotFound обрабатывает 404 ошибки
func (api *API) handleNotFound(w http.ResponseWriter, r *http.Request) {
	api.sendError(w, http.StatusNotFound, "Страница не найдена", nil)
}

// Вспомогательные функции

// sendJSON отправляет JSON ответ
func (api *API) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// sendError отправляет ошибку в формате JSON
func (api *API) sendError(w http.ResponseWriter, status int, message string, err error) {
	resp := viewmodels.ErrorResponse{
		Error: message,
	}

	if err != nil {
		resp.Details = err.Error()
	}

	api.sendJSON(w, status, resp)
}

// getBaseURL возвращает базовый URL из запроса
func getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	host := r.Host
	if host == "" {
		host = "localhost:8080"
	}

	return fmt.Sprintf("%s://%s", scheme, host)
}

// extractTextFragment извлекает фрагмент текста (первые или последние n слов)
func extractTextFragment(text string, n int, fromStart bool) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	if len(words) <= n {
		return strings.Join(words, " ")
	}

	if fromStart {
		return strings.Join(words[:n], " ")
	}

	return strings.Join(words[len(words)-n:], " ")
}
