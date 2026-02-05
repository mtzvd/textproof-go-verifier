package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"blockchain-verifier/internal/viewmodels"
	"blockchain-verifier/web/templates"

	"github.com/gorilla/mux"
)

// handleVerifyPage - главная страница проверки с табами
func (api *API) handleVerifyPage(w http.ResponseWriter, r *http.Request) {
	navVM := viewmodels.BuildHomeNavBar(r)
	nav := mapNavBar(navVM)
	flashData := getFlashData(r, w)

	api.renderHTML(
		w,
		r,
		templates.Base(
			viewmodels.PageMeta{Title: "Проверка текста", Description: "Проверьте подлинность текста по ID блока или содержимому"},
			nav,
			templates.VerifyContent(flashData),
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
				viewmodels.PageMeta{Title: "Проверка текста", Description: "Проверьте подлинность текста по ID блока или содержимому"},
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
			viewmodels.PageMeta{Title: "Результат проверки", Description: "Результат верификации текста в блокчейне TextProof"},
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
			viewmodels.PageMeta{Title: "Результат проверки", Description: "Результат верификации текста в блокчейне TextProof"},
			nav,
			templates.VerifyResult(result, flashData),
		),
	)
}
