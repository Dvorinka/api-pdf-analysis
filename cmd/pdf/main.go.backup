package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"apiservices/pdf-analysis/internal/pdf/analysis"
	"apiservices/pdf-analysis/internal/pdf/api"
	"apiservices/pdf-analysis/internal/pdf/auth"
)

func main() {
	logger := log.New(os.Stdout, "[pdf] ", log.LstdFlags)

	port := envString("PORT", "8086")
	apiKey := envString("PDF_API_KEY", "dev-pdf-key")
	maxUploadMB := envInt("PDF_MAX_UPLOAD_MB", 20)

	if apiKey == "dev-pdf-key" {
		logger.Println("PDF_API_KEY not set, using default development key")
	}

	service := analysis.NewService(int64(maxUploadMB) << 20)
	handler := api.NewHandler(service)

	mux := http.NewServeMux()
	mux.Handle("/v1/pdf/", auth.Middleware(apiKey)(handler))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Printf("service listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("server failed: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Printf("shutdown error: %v", err)
	}
}

func envString(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
