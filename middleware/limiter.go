package middleware

import (
	"net/http"
	"strings"

	"github.com/NayronFerreira/rate-limit-challenge/limiter"
)

func RateLimitMiddleware(next http.Handler, limiter *limiter.Limiter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("API_KEY")

		if token != "" {
			// Se um token de API for fornecido, verifique a taxa de solicitações para o token
			isBlocked, err := limiter.CheckRateLimit(r.Context(), token, true)
			if err != nil {
				http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if isBlocked {
				http.Error(w, "Your Token is temporarily blocked for exceeding the request limit.", http.StatusTooManyRequests)
				return
			}

		} else {
			// Se nenhum token de API for fornecido, verifique a taxa de solicitações para o endereço IP
			ip := strings.Split(r.RemoteAddr, ":")[0]
			isBlocked, err := limiter.CheckRateLimit(r.Context(), "ip:"+ip, false)
			if err != nil {
				http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if isBlocked {
				http.Error(w, "Your IP is temporarily blocked for exceeding the request limit.", http.StatusTooManyRequests)
				return
			}
		}

		// Se o token ou IP não estiver bloqueado, continue para o próximo manipulador
		next.ServeHTTP(w, r)
	})
}
