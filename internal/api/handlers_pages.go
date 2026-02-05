package api

import (
	"net/http"
	"time"

	"blockchain-verifier/internal/viewmodels"
	"blockchain-verifier/web/templates"
)

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
			viewmodels.PageMeta{Title: "Главная", Description: "TextProof — система подтверждения авторства текстов через блокчейн с Proof-of-Work"},
			nav,
			templates.HomeContent(statCards),
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
			viewmodels.PageMeta{Title: "О проекте", Description: "Как работает TextProof — технология блокчейн для защиты авторства текстов"},
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
			viewmodels.PageMeta{Title: "Политика конфиденциальности", Description: "Как TextProof обрабатывает и защищает ваши данные"},
			nav,
			templates.PrivacyContent(),
		),
	)
}

// handleTermsPage обрабатывает страницу "Условия использования"
func (api *API) handleTermsPage(w http.ResponseWriter, r *http.Request) {
	navVM := viewmodels.BuildHomeNavBar(r)
	nav := mapNavBar(navVM)

	api.renderHTML(
		w,
		r,
		templates.Base(
			viewmodels.PageMeta{Title: "Условия использования", Description: "Условия использования сервиса TextProof"},
			nav,
			templates.TermsContent(),
		),
	)
}

// handleNotFound обрабатывает 404 ошибки
func (api *API) handleNotFound(w http.ResponseWriter, r *http.Request) {
	api.sendError(w, http.StatusNotFound, "Страница не найдена", nil)
}
