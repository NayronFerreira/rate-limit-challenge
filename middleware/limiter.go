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

			isBlocked, err := limiter.CheckRateLimit(r.Context(), "token:"+token, true)
			if err != nil {
				http.Error(w, "Internal Error Server: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if isBlocked {
				http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame.", http.StatusTooManyRequests)
				return
			}

		} else {

			ip := strings.Split(r.RemoteAddr, ":")[0]
			isBlocked, err := limiter.CheckRateLimit(r.Context(), "ip:"+ip, false)
			if err != nil {
				http.Error(w, "Internal Error Server: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if isBlocked {
				http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame.", http.StatusTooManyRequests)
				return
			}

		}

		next.ServeHTTP(w, r)
	})
}
