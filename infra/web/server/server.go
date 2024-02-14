package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NayronFerreira/rate-limit-challenge/config"
	"github.com/NayronFerreira/rate-limit-challenge/infra/database"
	"github.com/NayronFerreira/rate-limit-challenge/infra/web/handler"
	"github.com/NayronFerreira/rate-limit-challenge/ratelimiter"
)

type Server struct {
	*http.Server
}

func New(port string, handler http.Handler) *Server {
	return &Server{
		Server: &http.Server{
			Addr:         ":" + port,
			Handler:      handler,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
}

func (s *Server) Start() {
	log.Println("Starting server on port", s.Addr)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Error starting server:", err)
	}
}

func SetupRateLimiter(cfg *config.Config) *ratelimiter.RateLimiter {
	redisClient := database.NewRedisClient(cfg)
	dbRedisDataLimiter := database.NewRedisDataLimiter(redisClient)
	rateLimiter := ratelimiter.NewLimiter(dbRedisDataLimiter, cfg.TokenMaxRequestsPerSecond, int64(cfg.LockDurationSeconds), int64(cfg.BlockDurationSeconds), int64(cfg.IPMaxRequestsPerSecond))

	if err := rateLimiter.RegisterPersonalizedTokens(context.Background()); err != nil {
		log.Fatal("Erro ao registrar o token:", err)
	}

	return rateLimiter
}

func SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", handler.RootHandler)
}

func WaitForShutdown(server *http.Server) {
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
