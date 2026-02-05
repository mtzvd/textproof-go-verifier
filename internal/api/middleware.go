package api

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// securityHeaders добавляет заголовки безопасности ко всем ответам.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Бейджи предназначены для встраивания через iframe на внешних сайтах
		if !strings.HasPrefix(r.URL.Path, "/api/badge/") {
			w.Header().Set("X-Frame-Options", "DENY")
		}
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-XSS-Protection", "0")
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware добавляет CORS-заголовки для публичного API (/api/v1/).
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/") {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// maxBody ограничивает размер тела запроса.
func maxBody(maxBytes int64, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		next(w, r)
	}
}

// ipRateLimiter ограничивает частоту запросов per-IP.
type ipRateLimiter struct {
	limiters sync.Map
	rate     rate.Limit
	burst    int
}

func newIPRateLimiter(r rate.Limit, burst int) *ipRateLimiter {
	rl := &ipRateLimiter{rate: r, burst: burst}

	// Очистка старых записей каждые 10 минут
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			rl.limiters.Range(func(key, _ any) bool {
				rl.limiters.Delete(key)
				return true
			})
		}
	}()

	return rl
}

func (rl *ipRateLimiter) getLimiter(ip string) *rate.Limiter {
	if v, ok := rl.limiters.Load(ip); ok {
		return v.(*rate.Limiter)
	}
	limiter := rate.NewLimiter(rl.rate, rl.burst)
	rl.limiters.Store(ip, limiter)
	return limiter
}

func (rl *ipRateLimiter) middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !rl.getLimiter(ip).Allow() {
			http.Error(w, `{"error":"Too Many Requests"}`, http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

// clientIP извлекает IP клиента из запроса.
func clientIP(r *http.Request) string {
	// Caddy передаёт реальный IP через X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if ip := strings.TrimSpace(strings.SplitN(xff, ",", 2)[0]); ip != "" {
			return ip
		}
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
