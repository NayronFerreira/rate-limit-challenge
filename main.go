package main

import (
	"context"
	"log"
	"net/http"

	"github.com/NayronFerreira/rate-limit-challenge/config"
	"github.com/NayronFerreira/rate-limit-challenge/infra/database"
	server "github.com/NayronFerreira/rate-limit-challenge/infra/web"
	"github.com/NayronFerreira/rate-limit-challenge/infra/web/handler"
	"github.com/NayronFerreira/rate-limit-challenge/infra/web/middleware"
	"github.com/NayronFerreira/rate-limit-challenge/ratelimiter"
)

func main() {
	cfg := loadConfig()
	rateLimiter := setupRateLimiter(cfg)
	startServer(cfg, rateLimiter)
}

func loadConfig() *config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Error loading config:", err)
	}
	return cfg
}

func setupRateLimiter(cfg *config.Config) *ratelimiter.RateLimiter {
	redisClient := database.NewRedisClient(cfg)
	dbRedisDataLimiter := database.NewRedisDataLimiter(redisClient)
	rateLimiter := ratelimiter.NewLimiter(dbRedisDataLimiter, cfg.TokenMaxRequestsPerSecond, int64(cfg.LockDurationSeconds), int64(cfg.BlockDurationSeconds), int64(cfg.IPMaxRequestsPerSecond))

	if err := rateLimiter.RegisterPersonalizedTokens(context.Background()); err != nil {
		log.Fatal("Erro ao registrar o token:", err)
	}

	return rateLimiter
}

func startServer(cfg *config.Config, rateLimiter *ratelimiter.RateLimiter) {
	mux := http.NewServeMux()
	setupRoutes(mux)

	rateLimitMiddleware := middleware.RateLimitMiddleware(mux, rateLimiter)
	srv := server.New(cfg.WebPort, rateLimitMiddleware)

	go func() {
		log.Println("Servidor HTTP iniciado na porta:", cfg.WebPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Erro ao iniciar o servidor:", err)
		}
	}()

	server.WaitForShutdown(srv)
}

func setupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", handler.RootHandler)
}
