package api

import (
	"io/fs"
	"net/http"
	"time"

	"blockchain-verifier/internal/blockchain"
	"blockchain-verifier/web"

	"github.com/gorilla/mux"

	httpSwagger "github.com/swaggo/http-swagger"

	"golang.org/x/time/rate"
)

const (
	MaxTextLength   = 1_000_000
	MaxAuthorLength = 200
	MaxTitleLength  = 500
	MaxBodySize     = 2 << 20 // 2MB
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
	// Глобальные middleware
	api.router.Use(securityHeaders)
	api.router.Use(corsMiddleware)

	// Rate limiter для POST-эндпоинтов (~10 req/min per IP)
	rl := newIPRateLimiter(rate.Every(6*time.Second), 10)

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
	api.router.HandleFunc("/docs", api.handleDocsPage).Methods("GET")
	// API routes (с rate limiting и ограничением body)
	api.router.HandleFunc("/api/deposit", rl.middleware(maxBody(MaxBodySize, api.handleDeposit))).Methods("POST")
	api.router.HandleFunc("/api/verify/id", rl.middleware(maxBody(MaxBodySize, api.handleVerifyByIDSubmit))).Methods("POST")
	api.router.HandleFunc("/api/verify/text", rl.middleware(maxBody(MaxBodySize, api.handleVerifyByTextSubmit))).Methods("POST")
	api.router.HandleFunc("/api/qrcode/{id}", api.handleQRCode).Methods("GET")
	api.router.HandleFunc("/api/badge/{id}", api.handleBadge).Methods("GET")

	// JSON API v1 (PUBLIC API, с rate limiting и ограничением body)
	api.router.HandleFunc("/api/v1/deposit", rl.middleware(maxBody(MaxBodySize, api.handleDepositJSON))).Methods("POST")
	api.router.HandleFunc("/api/v1/verify/id", rl.middleware(maxBody(MaxBodySize, api.handleVerifyByIDJSON))).Methods("POST")
	api.router.HandleFunc("/api/v1/verify/text", rl.middleware(maxBody(MaxBodySize, api.handleVerifyByTextJSON))).Methods("POST")
	api.router.HandleFunc("/api/v1/stats", api.handleStats).Methods("GET")
	api.router.HandleFunc("/api/v1/blockchain", api.handleBlockchainInfo).Methods("GET")
	api.router.HandleFunc("/api/v1/blockchain/export", api.handleBlockchainExport).Methods("GET")

	// Static files (embedded)
	staticSub, _ := fs.Sub(web.StaticFS, "static")
	fileServer := http.FileServer(http.FS(staticSub))
	api.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))

	// 404 handler
	api.router.NotFoundHandler = http.HandlerFunc(api.handleNotFound)
	// Swagger UI route
	api.router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // URL к swagger spec
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DomID("swagger-ui"),
	))
}

// ServeHTTP реализует интерфейс http.Handler
func (api *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.router.ServeHTTP(w, r)
}
