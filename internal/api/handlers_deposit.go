package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"blockchain-verifier/internal/blockchain"
	"blockchain-verifier/internal/viewmodels"
	"blockchain-verifier/web/templates"

	"github.com/gorilla/mux"
)

// handleDepositPage обрабатывает страницу депонирования
func (api *API) handleDepositPage(w http.ResponseWriter, r *http.Request) {
	navVM := viewmodels.BuildHomeNavBar(r)
	nav := mapNavBar(navVM)

	api.renderHTML(
		w,
		r,
		templates.Base(
			viewmodels.PageMeta{Title: "Депонирование текста", Description: "Зарегистрируйте авторство текста в блокчейне TextProof"},
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
			viewmodels.PageMeta{Title: "Результат депонирования", Description: "Информация о зафиксированном тексте в блокчейне TextProof"},
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
