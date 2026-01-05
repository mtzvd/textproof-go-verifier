package viewmodels

import "net/http"

func BuildHomeNavBar(r *http.Request) NavBar {
	currentPath := r.URL.Path

	return NavBar{
		Brand: "TextProof",
		Icon:  "fas fa-fingerprint",
		Items: []NavBarItem{
			{
				Label:  "Главная",
				Href:   "/",
				Icon:   "fas fa-home",
				Active: currentPath == "/",
				Align:  "start",
			},
			{
				Label:  "Депонировать",
				Href:   "/deposit",
				Icon:   "fas fa-upload",
				Active: currentPath == "/deposit",
				Align:  "start",
			},
			{
				Label: "Проверить",
				Href:  "/verify",
				Icon:  "fas fa-search",
				Active: currentPath == "/verify" ||
					len(currentPath) > 7 && currentPath[:7] == "/verify",
				Align: "start",
			},
			{
				Label:  "О проекте",
				Href:   "/about",
				Icon:   "fas fa-info-circle",
				Active: currentPath == "/about",
				Align:  "end",
			},
		},
	}
}
