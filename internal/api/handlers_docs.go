package api

import (
	"net/http"

	"blockchain-verifier/internal/viewmodels"
	"blockchain-verifier/web/templates"
)

// handleDocsPage отображает красивую страницу документации со встроенным Swagger
func (api *API) handleDocsPage(w http.ResponseWriter, r *http.Request) {
	navVM := viewmodels.BuildHomeNavBar(r)
	nav := mapNavBar(navVM)

	api.renderHTML(
		w,
		r,
		templates.Base(
			viewmodels.PageMeta{Title: "Документация API", Description: "Полное описание эндпоинтов TextProof API для интеграции"},
			nav,
			templates.DocsContent(),
		),
	)
}
