package api

import (
	"blockchain-verifier/internal/blockchain"
	"blockchain-verifier/internal/viewmodels"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// =====================================================================
// JSON API METHODS (PUBLIC API)
// =====================================================================

// handleDepositJSON godoc
//
// @Summary      Депонирование текста (JSON API)
// @Description  Регистрирует текст в блокчейне и возвращает JSON ответ
// @Tags         Deposit
// @Accept       json
// @Produce      json
// @Param        request body viewmodels.DepositRequest true "Данные для регистрации"
// @Success      200 {object} viewmodels.DepositResponse
// @Failure      400 {object} viewmodels.ErrorResponse
// @Failure      409 {object} viewmodels.ErrorResponse "Текст уже существует"
// @Failure      500 {object} viewmodels.ErrorResponse
// @Router       /api/v1/deposit [post]
func (api *API) handleDepositJSON(w http.ResponseWriter, r *http.Request) {
	var req viewmodels.DepositRequest

	// Парсим JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.sendError(w, http.StatusBadRequest, "Неверный формат JSON", err)
		return
	}

	// Валидация
	if err := api.validateDepositRequest(req); err != nil {
		api.sendError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Вычисляем хеш
	hash := sha256.Sum256([]byte(req.Text))
	contentHash := hex.EncodeToString(hash[:])

	// Проверка на дубликат
	_, existedBefore := api.blockchain.HasContentHash(contentHash)
	if existedBefore {
		api.sendError(w, http.StatusConflict, "Текст уже существует в блокчейне", nil)
		return
	}

	// Извлекаем фрагменты
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

	// Добавляем блок
	block, err := api.blockchain.AddBlock(data)
	if err != nil {
		api.sendError(w, http.StatusInternalServerError, "Не удалось добавить блок", err)
		return
	}

	// Формируем JSON ответ
	response := viewmodels.DepositResponsePublic{
		Success:   true,
		BlockID:   block.ID,
		Hash:      contentHash,
		Timestamp: block.Timestamp,
		VerifyURL: fmt.Sprintf("%s/verify/%s", getBaseURL(r), block.ID),
		QRCodeURL: fmt.Sprintf("%s/api/qrcode/%s", getBaseURL(r), block.ID),
	}

	api.sendJSON(w, http.StatusOK, response)
}

// handleVerifyByIDJSON godoc
//
// @Summary      Проверка по ID (JSON API)
// @Description  Проверяет текст по ID блока и возвращает JSON
// @Tags         Verify
// @Accept       json
// @Produce      json
// @Param        request body viewmodels.VerifyByIDRequest true "ID блока"
// @Success      200 {object} viewmodels.VerificationResponse
// @Failure      400 {object} viewmodels.ErrorResponse
// @Router       /api/v1/verify/id [post]
func (api *API) handleVerifyByIDJSON(w http.ResponseWriter, r *http.Request) {
	var req viewmodels.VerifyByIDRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.sendError(w, http.StatusBadRequest, "Неверный формат JSON", err)
		return
	}

	if strings.TrimSpace(req.ID) == "" {
		api.sendError(w, http.StatusBadRequest, "ID не может быть пустым", nil)
		return
	}

	// Ищем блок
	block, err := api.blockchain.GetBlockByID(req.ID)
	if err != nil {
		resp := viewmodels.VerificationResponse{
			Found: false,
		}
		api.sendJSON(w, http.StatusOK, resp)
		return
	}

	// Формируем ответ
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

// handleVerifyByTextJSON godoc
//
// @Summary      Проверка по тексту (JSON API)
// @Description  Проверяет текст по содержимому и возвращает JSON
// @Tags         Verify
// @Accept       json
// @Produce      json
// @Param        request body viewmodels.VerifyByTextRequest true "Текст для проверки"
// @Success      200 {object} viewmodels.VerificationResponse
// @Failure      400 {object} viewmodels.ErrorResponse
// @Router       /api/v1/verify/text [post]
func (api *API) handleVerifyByTextJSON(w http.ResponseWriter, r *http.Request) {
	var req viewmodels.VerifyByTextRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.sendError(w, http.StatusBadRequest, "Неверный формат JSON", err)
		return
	}

	if strings.TrimSpace(req.Text) == "" {
		api.sendError(w, http.StatusBadRequest, "Текст не может быть пустым", nil)
		return
	}

	// Валидация размера
	if len(req.Text) > MaxTextLength {
		api.sendError(w, http.StatusBadRequest,
			fmt.Sprintf("Текст слишком длинный (макс %d символов)", MaxTextLength), nil)
		return
	}

	// Вычисляем хеш
	hash := sha256.Sum256([]byte(req.Text))
	contentHash := hex.EncodeToString(hash[:])

	// Поиск по хешу
	block, exists := api.blockchain.HasContentHash(contentHash)
	if !exists {
		resp := viewmodels.VerificationResponse{
			Found: false,
		}
		api.sendJSON(w, http.StatusOK, resp)
		return
	}

	// Формируем ответ
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
