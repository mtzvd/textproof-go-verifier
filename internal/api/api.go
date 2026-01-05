package api

import (
	"net/http"

	"blockchain-verifier/internal/blockchain"

	"github.com/gorilla/mux"
)

const (
	MaxTextLength   = 1_000_000
	MaxAuthorLength = 200
	MaxTitleLength  = 500
)

// API представляет собой HTTP API сервер
type API struct {
	blockchain *blockchain.Blockchain
	router     *mux.Router
}

// NewAPI создает новый экземпляр API
func NewAPI(bc *blockchain.Blockchain) *API {
	api := &API{
		blockchain: bc,
		router:     mux.NewRouter(),
	}
	api.setupRoutes()
	return api
}

// setupRoutes настраивает маршруты API
func (api *API) setupRoutes() {
	// Web UI routes
	api.router.HandleFunc("/", api.handleHome).Methods("GET")
	api.router.HandleFunc("/deposit", api.handleDepositPage).Methods("GET")
	api.router.HandleFunc("/deposit/result/{id}", api.handleDepositResult).Methods("GET")
	api.router.HandleFunc("/verify", api.handleVerifyPage).Methods("GET")
	api.router.HandleFunc("/verify/{id}", api.handleVerifyDirectLink).Methods("GET")
	api.router.HandleFunc("/verify/result/{id}", api.handleVerifyResultPage).Methods("GET")
	api.router.HandleFunc("/about", api.handleAboutPage).Methods("GET")
	api.router.HandleFunc("/privacy", api.handlePrivacyPage).Methods("GET")
	api.router.HandleFunc("/terms", api.handleTermsPage).Methods("GET")
	// API routes
	api.router.HandleFunc("/api/deposit", api.handleDeposit).Methods("POST")
	api.router.HandleFunc("/api/verify/id", api.handleVerifyByIDSubmit).Methods("POST")
	api.router.HandleFunc("/api/verify/text", api.handleVerifyByTextSubmit).Methods("POST")
	api.router.HandleFunc("/api/stats", api.handleStats).Methods("GET")
	api.router.HandleFunc("/api/blockchain", api.handleBlockchainInfo).Methods("GET")
	api.router.HandleFunc("/api/qrcode/{id}", api.handleQRCode).Methods("GET")
	api.router.HandleFunc("/api/badge/{id}", api.handleBadge).Methods("GET")

	// Static files
	fs := http.FileServer(http.Dir("web/static"))
	api.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// 404 handler
	api.router.NotFoundHandler = http.HandlerFunc(api.handleNotFound)
}

// ServeHTTP реализует интерфейс http.Handler
func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.router.ServeHTTP(w, r)
}
