package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"blockchain-verifier/internal/viewmodels"
	"blockchain-verifier/web/templates/components"

	"github.com/a-h/templ"
)

// renderHTML упрощает рендеринг HTML шаблонов
func (api *API) renderHTML(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(r.Context(), w); err != nil {
		slog.Error("Render error", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

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
