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
			"Документация API",
			nav,
			templates.DocsContent(),
		),
	)
}
