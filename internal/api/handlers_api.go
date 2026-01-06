package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/skip2/go-qrcode"

	"blockchain-verifier/internal/viewmodels"
	"blockchain-verifier/web/templates/components"
)

// handleStats godoc
//
// @Summary      Статистика блокчейна
// @Description  Возвращает общую информацию и статистику о состоянии блокчейна
// @Tags         Stats
// @Produce      json
// @Success      200 {object} viewmodels.StatsResponse "Статистика блокчейна"
// @Failure      500 {object} viewmodels.ErrorResponse "Ошибка получения статистики"
// @Router       /api/stats [get]
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

// handleBlockchainInfo godoc
//
// @Summary      Информация о блокчейне
// @Description  Возвращает техническую информацию о блокчейне
// @Tags         Stats
// @Produce      json
// @Success      200 {object} viewmodels.BlockchainInfoResponse "Информация о блокчейне"
// @Router       /api/blockchain [get]
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

// handleQRCode godoc
//
// @Summary      Генерация QR-кода
// @Description  Генерирует QR-код для быстрой проверки текста по ID блока
// @Tags         Utils
// @Produce      image/png
// @Param        id path string true "ID блока" example("000-000-001")
// @Success      200 {file} binary "QR-код в формате PNG"
// @Failure      400 {object} viewmodels.ErrorResponse "Неверный ID"
// @Failure      404 {object} viewmodels.ErrorResponse "Блок не найден"
// @Failure      500 {object} viewmodels.ErrorResponse "Ошибка генерации QR-кода"
// @Router       /api/qrcode/{id} [get]
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

// handleBadge godoc
//
// @Summary      HTML Badge для встраивания
// @Description  Генерирует HTML код badge для встраивания на веб-страницы
// @Tags         Utils
// @Produce      text/html
// @Param        id path string true "ID блока" example("000-000-001")
// @Success      200 {string} string "HTML код badge"
// @Failure      400 {object} viewmodels.ErrorResponse "Неверный ID"
// @Failure      404 {object} viewmodels.ErrorResponse "Блок не найден"
// @Router       /api/badge/{id} [get]
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
