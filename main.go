package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NayronFerreira/rate-limit-challenge/config"
	database "github.com/NayronFerreira/rate-limit-challenge/infra/database/impl"
	"github.com/NayronFerreira/rate-limit-challenge/middleware"
	"github.com/NayronFerreira/rate-limit-challenge/ratelimiter"
	"github.com/go-redis/redis/v8"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisURL,
		Password: "",
		DB:       0,
	})

	dbRedis := database.NewRedisDataLimiter(redisClient)

	rateLimiter := ratelimiter.NewLimiter(dbRedis, config.TokenMaxRequestsPerSecond, int64(config.LockDurationSeconds), int64(config.BlockDurationSeconds), int64(config.IPMaxRequestsPerSecond))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Requisição bem-sucedida!")
	})

	if err = rateLimiter.RegisterToken(context.Background()); err != nil {
		log.Fatal("Erro ao registrar o token:", err)
	}

	rateLimitMiddleware := middleware.RateLimitMiddleware(http.DefaultServeMux, rateLimiter)

	server := &http.Server{
		Addr:    ":" + config.WebPort,
		Handler: rateLimitMiddleware,
	}

	go func() {
		log.Println("Servidor HTTP iniciado na porta:", config.WebPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Erro ao iniciar o servidor:", err)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	<-signalChan

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Erro ao encerrar o servidor:", err)
	}

	log.Println("Servidor encerrado")
}
