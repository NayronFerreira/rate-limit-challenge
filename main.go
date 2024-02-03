package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/NayronFerreira/rate-limit-challenge/limiter"
	"github.com/NayronFerreira/rate-limit-challenge/middleware"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

func loadEnv() error {
	err := godotenv.Load(".env")
	if err != nil {
		return err
	}
	return nil
}

func main() {

	err := loadEnv()
	if err != nil {
		log.Fatal("Erro ao carregar as variáveis de ambiente:", err)
	}

	ipMaxRequestsPerSecondStr := os.Getenv("IP_MAX_REQUESTS_PER_SECOND")

	token1MaxRequestsPerSecondStr := os.Getenv("TOKEN_1_MAX_REQUESTS_PER_SECOND")
	token2MaxRequestsPerSecondStr := os.Getenv("TOKEN_2_MAX_REQUESTS_PER_SECOND")
	token3MaxRequestsPerSecondStr := os.Getenv("TOKEN_3_MAX_REQUESTS_PER_SECOND")
	token4MaxRequestsPerSecondStr := os.Getenv("TOKEN_4_MAX_REQUESTS_PER_SECOND")
	token5MaxRequestsPerSecondStr := os.Getenv("TOKEN_5_MAX_REQUESTS_PER_SECOND")

	lockDurationStr := os.Getenv("LOCK_DURATION_SECONDS")
	blockDurationStr := os.Getenv("BLOCK_DURATION_SECONDS")

	webPort := os.Getenv("APP_WEB_PORT")
	redisURL := os.Getenv("REDIS_URL")

	ipMaxRequestsPerSecond, err := strconv.Atoi(ipMaxRequestsPerSecondStr)
	if err != nil {
		log.Fatal("Erro ao converter MAX_REQUESTS_PER_SECOND para int:", err)
	}

	token1MaxRequestsPerSecond, err := strconv.Atoi(token1MaxRequestsPerSecondStr)
	if err != nil {
		log.Fatal("Erro ao converter TOKEN_1_MAX_REQUESTS_PER_SECOND para int:", err)
	}

	token2MaxRequestsPerSecond, err := strconv.Atoi(token2MaxRequestsPerSecondStr)
	if err != nil {
		log.Fatal("Erro ao converter TOKEN_2_MAX_REQUESTS_PER_SECOND para int:", err)
	}

	token3MaxRequestsPerSecond, err := strconv.Atoi(token3MaxRequestsPerSecondStr)
	if err != nil {
		log.Fatal("Erro ao converter TOKEN_3_MAX_REQUESTS_PER_SECOND para int:", err)
	}

	token4MaxRequestsPerSecond, err := strconv.Atoi(token4MaxRequestsPerSecondStr)
	if err != nil {
		log.Fatal("Erro ao converter TOKEN_4_MAX_REQUESTS_PER_SECOND para int:", err)
	}

	token5MaxRequestsPerSecond, err := strconv.Atoi(token5MaxRequestsPerSecondStr)
	if err != nil {
		log.Fatal("Erro ao converter TOKEN_5_MAX_REQUESTS_PER_SECOND para int:", err)
	}

	lockDurationSeconds, err := strconv.Atoi(lockDurationStr)
	if err != nil {
		log.Fatal("Erro ao converter LOCK_DURATION_SECONDS para int:", err)
	}

	blockDurationSeconds, err := strconv.Atoi(blockDurationStr)
	if err != nil {
		log.Fatal("Erro ao converter LOCK_DURATION_SECONDS para int:", err)
	}

	tokenConfig := map[string]int64{
		"TOKEN_1": int64(token1MaxRequestsPerSecond),
		"TOKEN_2": int64(token2MaxRequestsPerSecond),
		"TOKEN_3": int64(token3MaxRequestsPerSecond),
		"TOKEN_4": int64(token4MaxRequestsPerSecond),
		"TOKEN_5": int64(token5MaxRequestsPerSecond),
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisURL, //"localhost:6379"
		Password: "",
		DB:       0,
	})

	rateLimiter := limiter.NewLimiter(redisClient, tokenConfig, int64(lockDurationSeconds), int64(blockDurationSeconds), int64(ipMaxRequestsPerSecond))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Requisição bem-sucedida!")
	})

	if err = rateLimiter.RegisterToken(context.Background()); err != nil {
		log.Fatal("Erro ao registrar o token:", err)
	}

	// Crie um middleware RateLimitMiddleware com o Limiter
	rateLimitMiddleware := middleware.RateLimitMiddleware(http.DefaultServeMux, rateLimiter)

	// Crie um servidor HTTP com o middleware RateLimitMiddleware
	server := &http.Server{
		Addr:    ":" + webPort,
		Handler: rateLimitMiddleware,
	}

	// Inicie o servidor HTTP em uma goroutine
	go func() {
		log.Println("Servidor HTTP iniciado na porta:", webPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Erro ao iniciar o servidor:", err)
		}
	}()

	// Crie um canal para sinais do sistema (como SIGINT e SIGTERM)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Aguarde um sinal para encerrar o servidor
	<-signalChan

	// Crie um contexto com timeout para encerrar o servidor
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Encerre o servidor HTTP
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Erro ao encerrar o servidor:", err)
	}

	log.Println("Servidor encerrado")
}
