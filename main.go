package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/NayronFerreira/rate-limit-challenge/config"
	"github.com/NayronFerreira/rate-limit-challenge/infra/database"
	server "github.com/NayronFerreira/rate-limit-challenge/infra/web"
	"github.com/NayronFerreira/rate-limit-challenge/infra/web/middleware"
	"github.com/NayronFerreira/rate-limit-challenge/ratelimiter"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	redisClient := database.NewRedisClient(config)

	dbRedis := database.NewRedisDataLimiter(redisClient)

	rateLimiter := ratelimiter.NewLimiter(dbRedis, config.TokenMaxRequestsPerSecond, int64(config.LockDurationSeconds), int64(config.BlockDurationSeconds), int64(config.IPMaxRequestsPerSecond))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Requisição bem-sucedida!")
	})

	if err = rateLimiter.RegisterToken(context.Background()); err != nil {
		log.Fatal("Erro ao registrar o token:", err)
	}

	rateLimitMiddleware := middleware.RateLimitMiddleware(http.DefaultServeMux, rateLimiter)

	srv := server.New(config.WebPort, rateLimitMiddleware)

	go func() {
		log.Println("Servidor HTTP iniciado na porta:", config.WebPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Erro ao iniciar o servidor:", err)
		}
	}()

	server.WaitForShutdown(srv)
}
