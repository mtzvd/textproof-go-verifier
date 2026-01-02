package viewmodels

import "net/http"

func BuildHomeNavBar(r *http.Request) NavBar {
	return NavBar{
		Brand: "TextProof",
		Icon:  "fas fa-fingerprint",
		Items: []NavBarItem{
			{
				Label:  "Главная",
				Href:   "/",
				Icon:   "fas fa-home",
				Active: r.URL.Path == "/",
				Align:  "start",
			},
			{
				Label:  "Депонировать",
				Href:   "/deposit",
				Icon:   "fas fa-upload",
				Active: r.URL.Path == "/deposit",
				Align:  "start",
			},
			{
				Label:  "Проверить",
				Href:   "/verify",
				Icon:   "fas fa-search",
				Active: r.URL.Path == "/verify",
				Align:  "start",
			},
			{
				Label: "API Docs",
				Href:  "/api/docs",
				Icon:  "fas fa-book",
				Align: "end",
			},
		},
	}
}
