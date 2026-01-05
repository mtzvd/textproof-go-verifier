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
	"blockchain-verifier/web/templates/components"

	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"github.com/skip2/go-qrcode"
)

const (
	MaxTextLength   = 1_000_000
	MaxAuthorLength = 200
	MaxTitleLength  = 500
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
	api.router.HandleFunc("/test", api.handleTest).Methods("GET")
	api.router.HandleFunc("/deposit/result/{id}", api.handleDepositResult).Methods("GET")
	api.router.HandleFunc("/verify", api.handleVerifyPage).Methods("GET")
	api.router.HandleFunc("/verify/{id}", api.handleVerifyDirectLink).Methods("GET")
	api.router.HandleFunc("/verify/result/{id}", api.handleVerifyResultPage).Methods("GET")
	api.router.HandleFunc("/about", api.handleAboutPage).Methods("GET")
	api.router.HandleFunc("/privacy", api.handlePrivacyPage).Methods("GET")
	// API routes
	api.router.HandleFunc("/api/deposit", api.handleDeposit).Methods("POST")
	api.router.HandleFunc("/api/verify/id", api.handleVerifyByIDSubmit).Methods("POST")
	api.router.HandleFunc("/api/verify/text", api.handleVerifyByTextSubmit).Methods("POST")
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

// handleTest возвращает простой HTML для тестирования
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

// handleHome обрабатывает главную страницу
func (api *API) handleHome(w http.ResponseWriter, r *http.Request) {
	// Собираем статистику
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

	navVM := viewmodels.BuildHomeNavBar(r)
	nav := mapNavBar(navVM)
	statCards := mapStatsCards(stats)

	api.renderHTML(
		w,
		r,
		templates.Base(
			"Главная",
			nav,
			templates.HomeContent(statCards),
		),
	)
}

// handleDepositPage обрабатывает страницу депонирования
func (api *API) handleDepositPage(w http.ResponseWriter, r *http.Request) {
	navVM := viewmodels.BuildHomeNavBar(r)
	nav := mapNavBar(navVM)

	api.renderHTML(
		w,
		r,
		templates.Base(
			"Депонирование текста",
			nav,
			templates.DepositContent(),
		),
	)
}

// Функция валидации депозита
func (api *API) validateDepositRequest(req viewmodels.DepositRequest) error {
	if strings.TrimSpace(req.AuthorName) == "" {
		return fmt.Errorf("имя автора не может быть пустым")
	}
	if len(req.AuthorName) > MaxAuthorLength {
		return fmt.Errorf("имя автора слишком длинное (макс %d символов)", MaxAuthorLength)
	}

	if strings.TrimSpace(req.Title) == "" {
		return fmt.Errorf("название не может быть пустым")
	}
	if len(req.Title) > MaxTitleLength {
		return fmt.Errorf("название слишком длинное (макс %d символов)", MaxTitleLength)
	}

	if strings.TrimSpace(req.Text) == "" {
		return fmt.Errorf("текст не может быть пустым")
	}
	if len(req.Text) > MaxTextLength {
		return fmt.Errorf("текст слишком длинный (макс %d символов)", MaxTextLength)
	}

	return nil
}

// handleDeposit обрабатывает запрос на депонирование текста
func (api *API) handleDeposit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		api.sendError(w, http.StatusBadRequest, "Неверные данные формы", err)
		return
	}

	req := viewmodels.DepositRequest{
		AuthorName: r.FormValue("author_name"),
		Title:      r.FormValue("title"),
		Text:       r.FormValue("text"),
		PublicKey:  r.FormValue("public_key"),
	}

	// Валидация
	if err := api.validateDepositRequest(req); err != nil {
		api.sendError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Вычисляем хеш текста
	hash := sha256.Sum256([]byte(req.Text))
	contentHash := hex.EncodeToString(hash[:])

	_, existedBefore := api.blockchain.HasContentHash(contentHash)

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

	//Устанавливаем flash message вместо query параметров
	flashData := make(map[string]string)
	if existedBefore {
		flashData["duplicate"] = "true"
		setFlash(w, "warning", "duplicate", flashData)
	} else {
		setFlash(w, "success", "new_deposit", flashData)
	}

	redirectURL := fmt.Sprintf("/deposit/result/%s", block.ID)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
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
		Matches:   true, // Если блок найден, то хеш уже совпадает
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

	if strings.TrimSpace(req.Text) == "" {
		api.sendError(w, http.StatusBadRequest, "Текст не может быть пустым", nil)
		return
	}

	// Валидация размера текста
	if len(req.Text) > MaxTextLength {
		api.sendError(w, http.StatusBadRequest,
			fmt.Sprintf("Текст слишком длинный (макс %d символов)", MaxTextLength), nil)
		return
	}

	// Вычисляем хеш текста
	hash := sha256.Sum256([]byte(req.Text))
	contentHash := hex.EncodeToString(hash[:])

	// O(1) поиск через индекс
	block, exists := api.blockchain.HasContentHash(contentHash)
	if !exists {
		resp := viewmodels.VerificationResponse{
			Found: false,
		}
		api.sendJSON(w, http.StatusOK, resp)
		return
	}

	resp := viewmodels.VerificationResponse{
		Found:     true,
		BlockID:   block.ID,
		Author:    block.Data.AuthorName,
		Title:     block.Data.Title,
		Timestamp: block.Timestamp,
		Hash:      block.Data.ContentHash,
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

	// Формируем данные
	baseURL := getBaseURL(r)
	qrCodeURL := fmt.Sprintf("%s/api/qrcode/%s", baseURL, id)
	verifyURL := fmt.Sprintf("%s/verify/%s", baseURL, id)

	// ВАЖНО: правильный Content-Type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Рендерим templ-компонент
	err = components.Badge(
		block.Data.Title,
		block.Data.AuthorName,
		block.ID,
		qrCodeURL,
		block.Timestamp.Format("02.01.2006 15:04"),
		verifyURL,
	).Render(r.Context(), w)

	if err != nil {
		log.Printf("Badge render error: %v", err)
		http.Error(w, "Failed to render badge", http.StatusInternalServerError)
	}
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

// mapNavBar преобразует viewmodel NavBar в компонент NavBar
func mapNavBar(vm viewmodels.NavBar) components.NavBar {
	items := make([]components.NavBarItem, 0, len(vm.Items))

	for _, it := range vm.Items {
		items = append(items, components.NavBarItem{
			Label:  it.Label,
			Href:   it.Href,
			Icon:   it.Icon,
			Active: it.Active,
			Align:  it.Align,
		})
	}

	return components.NavBar{
		Brand: vm.Brand,
		Icon:  vm.Icon,
		Items: items,
	}
}

// handleDepositResult обрабатывает страницу результата депонирования
func (api *API) handleDepositResult(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	block, err := api.blockchain.GetBlockByID(id)
	if err != nil {
		api.sendError(w, http.StatusNotFound, "Результат депозита не найден", err)
		return
	}

	// Читаем flash message (будет FlashData{Show: false} при прямом переходе)
	flashData := getFlashData(r, w)

	nav := mapNavBar(viewmodels.BuildHomeNavBar(r))
	formattedTime := block.Timestamp.Format("02.01.2006 15:04:05")

	api.renderHTML(
		w,
		r,
		templates.Base(
			"Результат депонирования",
			nav,
			templates.DepositResultPage(
				block.ID,
				block.Data.ContentHash,
				formattedTime,
				fmt.Sprintf("%s/api/qrcode/%s", getBaseURL(r), block.ID),
				fmt.Sprintf("%s/verify/%s", getBaseURL(r), block.ID),
				fmt.Sprintf("%s/api/badge/%s", getBaseURL(r), block.ID),
				block.Data.AuthorName,
				block.Data.Title,
				flashData,
			),
		),
	)
}

// handleVerifyPage - главная страница проверки с табами
func (api *API) handleVerifyPage(w http.ResponseWriter, r *http.Request) {
	navVM := viewmodels.BuildHomeNavBar(r)
	nav := mapNavBar(navVM)
	flashData := getFlashData(r, w)

	api.renderHTML(
		w,
		r,
		templates.Base(
			"Проверка текста",
			nav,
			templates.VerifyContent(flashData),
		),
	)
}

// handleVerifyByIDPage - страница проверки по ID
func (api *API) handleVerifyByIDPage(w http.ResponseWriter, r *http.Request) {
	navVM := viewmodels.BuildHomeNavBar(r)
	nav := mapNavBar(navVM)

	api.renderHTML(
		w,
		r,
		templates.Base(
			"Проверка по ID",
			nav,
			templates.VerifyByID(nil, "", ""),
		),
	)
}

// handleVerifyByTextPage - страница проверки по тексту
func (api *API) handleVerifyByTextPage(w http.ResponseWriter, r *http.Request) {
	navVM := viewmodels.BuildHomeNavBar(r)
	nav := mapNavBar(navVM)

	api.renderHTML(
		w,
		r,
		templates.Base(
			"Проверка по тексту",
			nav,
			templates.VerifyByText(nil, "", ""),
		),
	)
}

// handleVerifyDirectLink - прямая ссылка /verify/{id} (автоматическая проверка)
func (api *API) handleVerifyDirectLink(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	block, err := api.blockchain.GetBlockByID(id)

	navVM := viewmodels.BuildHomeNavBar(r)
	nav := mapNavBar(navVM)
	flashData := getFlashData(r, w)

	if err != nil {
		// Блок не найден - показываем ошибку на странице verify
		setFlash(w, "danger", "not_found", map[string]string{
			"id": id,
		})

		api.renderHTML(
			w,
			r,
			templates.Base(
				"Проверка текста",
				nav,
				templates.VerifyContent(flashData),
			),
		)
		return
	}

	// Блок найден - показываем результат
	result := viewmodels.VerificationResponse{
		Found:     true,
		BlockID:   block.ID,
		Author:    block.Data.AuthorName,
		Title:     block.Data.Title,
		Timestamp: block.Timestamp,
		Hash:      block.Data.ContentHash,
		Matches:   true,
	}

	// Устанавливаем flash для успешной проверки
	setFlash(w, "success", "verified", nil)

	api.renderHTML(
		w,
		r,
		templates.Base(
			"Результат проверки",
			nav,
			templates.VerifyResult(result, getFlashData(r, w)),
		),
	)
}

// handleVerifyByIDSubmit - обработка формы проверки по ID
func (api *API) handleVerifyByIDSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/verify", http.StatusSeeOther)
		return
	}

	id := strings.TrimSpace(r.FormValue("id"))
	if id == "" {
		setFlash(w, "danger", "empty_id", nil)
		http.Redirect(w, r, "/verify", http.StatusSeeOther)
		return
	}

	// Пытаемся найти блок
	block, err := api.blockchain.GetBlockByID(id)

	if err != nil {
		// Блок не найден - возвращаем на форму с ошибкой
		setFlash(w, "warning", "not_found", map[string]string{
			"id": id,
		})
		http.Redirect(w, r, "/verify", http.StatusSeeOther)
		return
	}

	// Блок найден - редирект на страницу результата
	setFlash(w, "success", "verified", nil)
	http.Redirect(w, r, fmt.Sprintf("/verify/result/%s", block.ID), http.StatusSeeOther)
}

// handleVerifyByTextSubmit - обработка формы проверки по тексту
func (api *API) handleVerifyByTextSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/verify", http.StatusSeeOther)
		return
	}

	text := r.FormValue("text")
	if strings.TrimSpace(text) == "" {
		setFlash(w, "danger", "empty_text", nil)
		http.Redirect(w, r, "/verify", http.StatusSeeOther)
		return
	}

	// Валидация размера
	if len(text) > MaxTextLength {
		setFlash(w, "danger", "text_too_long", map[string]string{
			"max": fmt.Sprintf("%d", MaxTextLength),
		})
		http.Redirect(w, r, "/verify", http.StatusSeeOther)
		return
	}

	// Вычисляем хеш текста
	hash := sha256.Sum256([]byte(text))
	contentHash := hex.EncodeToString(hash[:])

	// O(1) поиск через индекс
	block, exists := api.blockchain.HasContentHash(contentHash)

	if !exists {
		// Текст не найден
		setFlash(w, "warning", "text_not_found", nil)
		http.Redirect(w, r, "/verify", http.StatusSeeOther)
		return
	}

	// Текст найден - редирект на страницу результата
	setFlash(w, "success", "verified", nil)
	http.Redirect(w, r, fmt.Sprintf("/verify/result/%s", block.ID), http.StatusSeeOther)
}

// handleVerifyResultPage - страница результата проверки
func (api *API) handleVerifyResultPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	block, err := api.blockchain.GetBlockByID(id)
	if err != nil {
		// Если блок не найден - редирект на verify с ошибкой
		setFlash(w, "danger", "not_found", map[string]string{"id": id})
		http.Redirect(w, r, "/verify", http.StatusSeeOther)
		return
	}

	result := viewmodels.VerificationResponse{
		Found:     true,
		BlockID:   block.ID,
		Author:    block.Data.AuthorName,
		Title:     block.Data.Title,
		Timestamp: block.Timestamp,
		Hash:      block.Data.ContentHash,
		Matches:   true,
	}

	flashData := getFlashData(r, w)
	navVM := viewmodels.BuildHomeNavBar(r)
	nav := mapNavBar(navVM)

	api.renderHTML(
		w,
		r,
		templates.Base(
			"Результат проверки",
			nav,
			templates.VerifyResult(result, flashData),
		),
	)
}

// handleAboutPage обрабатывает страницу "О проекте"
func (api *API) handleAboutPage(w http.ResponseWriter, r *http.Request) {
	navVM := viewmodels.BuildHomeNavBar(r)
	nav := mapNavBar(navVM)

	api.renderHTML(
		w,
		r,
		templates.Base(
			"О проекте",
			nav,
			templates.AboutContent(),
		),
	)
}

// handlePrivacyPage обрабатывает страницу "Конфиденциальность"
func (api *API) handlePrivacyPage(w http.ResponseWriter, r *http.Request) {
	navVM := viewmodels.BuildHomeNavBar(r)
	nav := mapNavBar(navVM)

	api.renderHTML(
		w,
		r,
		templates.Base(
			"Политика конфиденциальности",
			nav,
			templates.PrivacyContent(),
		),
	)
}
