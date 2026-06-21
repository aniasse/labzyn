// Command server is the Labzyn API server entrypoint.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aniasse/labzyn/server/internal/api"
	"github.com/aniasse/labzyn/server/internal/config"
	"github.com/aniasse/labzyn/server/internal/database"
	"github.com/aniasse/labzyn/server/internal/llm"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("labzyn-server: %v", err)
	}
}

func run() error {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	log.Println("labzyn-server: connected to database")

	if err := database.Migrate(ctx, pool); err != nil {
		return err
	}
	log.Println("labzyn-server: migrations up to date")

	llmClient := llm.New(cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel)
	if llmClient.Enabled() {
		log.Printf("labzyn-server: agent enabled (model %s)", cfg.LLMModel)
	} else {
		log.Println("labzyn-server: agent disabled (no LABZYN_LLM_API_KEY)")
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.NewServer(pool, cfg.JWTSecret, cfg.GatewayURL, llmClient),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("labzyn-server: listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("labzyn-server: http error: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Println("labzyn-server: shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	_ = os.Stdout.Sync()
	return nil
}
