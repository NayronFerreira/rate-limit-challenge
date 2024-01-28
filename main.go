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
	// Carregue as variáveis de ambiente do arquivo .env
	err := loadEnv()
	if err != nil {
		log.Fatal("Erro ao carregar as variáveis de ambiente:", err)
	}

	// Obtenha as variáveis de ambiente configuradas
	ipMaxRequestsPerSecondStr := os.Getenv("IP_MAX_REQUESTS_PER_SECOND")
	tokenMaxRequestsPerSecondStr := os.Getenv("TOKEN_MAX_REQUESTS_PER_SECOND")

	lockDurationStr := os.Getenv("LOCK_DURATION_SECONDS")
	blockDurationStr := os.Getenv("BLOCK_DURATION_SECONDS")
	webPort := os.Getenv("APP_WEB_PORT")
	redisURL := os.Getenv("REDIS_URL")

	// Converta as variáveis de ambiente para os tipos apropriados
	ipMaxRequestsPerSecond, err := strconv.Atoi(ipMaxRequestsPerSecondStr)
	if err != nil {
		log.Fatal("Erro ao converter MAX_REQUESTS_PER_SECOND para int:", err)
	}

	tokenMaxRequestsPerSecond, err := strconv.Atoi(tokenMaxRequestsPerSecondStr)
	if err != nil {
		log.Fatal("Erro ao converter MAX_REQUESTS_PER_SECOND para int:", err)
	}

	lockDurationSeconds, err := strconv.Atoi(lockDurationStr)
	if err != nil {
		log.Fatal("Erro ao converter LOCK_DURATION_SECONDS para int:", err)
	}
	blockDurationSeconds, err := strconv.Atoi(blockDurationStr)
	if err != nil {
		log.Fatal("Erro ao converter LOCK_DURATION_SECONDS para int:", err)
	}

	log.Println("lockDurationSeconds= ", lockDurationSeconds)

	// Crie um cliente Redis (substitua estas linhas com a configuração real do seu cliente Redis)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisURL, //"localhost:6379"
		Password: "",       // no password set
		DB:       0,        // use default DB
	})

	redisStore := &limiter.RedisStore{
		Client: redisClient,
	}

	//TODO: Tokens com limites de taxa personalizados
	if err = redisStore.SetTokenRateLimit(context.Background(), "dfweuihrfi8943gj902ghu94jf0wj", 10, 5); err != nil {
		log.Fatal("Erro ao configurar o limite de taxa para token1:", err)
	}

	if err = redisStore.SetTokenRateLimit(context.Background(), "fjvre3489g5uj93w2rtguj34r903uj", 15, 5); err != nil {
		log.Fatal("Erro ao configurar o limite de taxa para token2:", err)
	}

	// Crie uma instância do Limiter com o cliente Redis
	rateLimiter := limiter.NewLimiter(redisStore, tokenMaxRequestsPerSecond, ipMaxRequestsPerSecond, lockDurationSeconds, blockDurationSeconds)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Requisição bem-sucedida!")
	})

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
