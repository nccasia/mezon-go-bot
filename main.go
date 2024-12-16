package main

import (
	"ncc/go-mezon-bot/config"
	"ncc/go-mezon-bot/internal/bot"
	"ncc/go-mezon-bot/internal/logger"
	"net/http"

	"go.uber.org/zap"
)

// healthCheckHandler handles the health check request
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "ok"}`))
}

func main() {
	// Load Config
	cfg := config.LoadConfig()

	// Setup Logger
	log := logger.NewLogger(cfg.LogFile)
	defer log.Sync() // Flush log

	// Start Bot Checkin
	b, err := bot.NewBot(cfg, log)
	if err != nil {
		log.Fatal("Failed to initialize bot", zap.Error(err))
	}
	b.StartCheckin()

	// Register the health check endpoint
	http.HandleFunc("/health", healthCheckHandler)

	// Define the port
	port := "8080"

	log.Info("Starting server on port", zap.Any("port", port))

	// Start the HTTP server
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Error starting server", zap.Error(err))
	}
}
