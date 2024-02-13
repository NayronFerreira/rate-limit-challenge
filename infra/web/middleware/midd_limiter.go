package middleware

import (
	"net/http"
	"strings"

	limiter "github.com/NayronFerreira/rate-limit-challenge/ratelimiter"
)

func RateLimitMiddleware(next http.Handler, rateLimiter *limiter.RateLimiter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("API_KEY")

		if token != "" && rateLimiter.TokenExists(token) {

			isBlocked, err := rateLimiter.CheckRateLimitForKey(r.Context(), token, true)
			if err != nil {
				http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if isBlocked {
				http.Error(w, "Your Token have reached the maximum number of requests or actions allowed within a certain time frame.", http.StatusTooManyRequests)
				return
			}

		} else {

			ip := strings.Split(r.RemoteAddr, ":")[0]
			isBlocked, err := rateLimiter.CheckRateLimitForKey(r.Context(), ip, false)
			if err != nil {
				http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if isBlocked {
				http.Error(w, "Your IP have reached the maximum number of requests or actions allowed within a certain time frame.", http.StatusTooManyRequests)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
