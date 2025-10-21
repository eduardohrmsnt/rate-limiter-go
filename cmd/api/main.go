package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/eduardohermesneto/rate-limiter/config"
	"github.com/eduardohermesneto/rate-limiter/internal/domain"
	"github.com/eduardohermesneto/rate-limiter/internal/infra/storage"
	"github.com/eduardohermesneto/rate-limiter/internal/infra/web"
	"github.com/eduardohermesneto/rate-limiter/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var store domain.Storage
	redisStore, err := storage.NewRedisStorage(
		cfg.RedisHost,
		cfg.RedisPort,
		cfg.RedisPassword,
		cfg.RedisDB,
	)
	if err != nil {
		log.Printf("Failed to connect to Redis: %v. Using memory storage", err)
		store = storage.NewMemoryStorage()
	} else {
		store = redisStore
	}
	defer store.Close()

	limiter := usecase.NewRateLimiter(
		store,
		cfg.RateLimitIP,
		cfg.RateLimitToken,
		cfg.BlockDuration,
	)

	middleware := web.NewRateLimiterMiddleware(limiter)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", web.HealthHandler)
	mux.HandleFunc("/test", web.TestHandler)

	handler := middleware.Handle(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: handler,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")
}
