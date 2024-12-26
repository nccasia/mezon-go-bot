package main

import (
	"fmt"
	"mezon-go-bot/config"
	"mezon-go-bot/internal/constants"
	"mezon-go-bot/internal/logger"
	"net/http"

	"go.uber.org/zap"
)

// healthCheckHandler handles the health check request
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "ok"}`))
}

var bot IBot

func main() {
	// Load Config
	cfg := config.LoadConfig()

	// Setup Logger
	log := logger.NewLogger(cfg.LogFile)
	defer log.Sync() // Flush log

	// Start Bot Checkin
	var err error
	bot, err = NewBot(cfg, log)
	if err != nil {
		log.Fatal("Failed to initialize bot checkin", zap.Error(err))
	}

	// registry all command here
	bot.RegisterCmd(constants.NCC8_COMMAND, Ncc8Handler)

	bot.Start()

	// Register the health check endpoint
	http.HandleFunc("/health", healthCheckHandler)

	// Define the port
	port := "8080"

	log.Info("Starting server on port", zap.Any("port", port))
	fmt.Println("Starting server on port:", port)

	// Start the HTTP server
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Error starting server", zap.Error(err))
		fmt.Println("Error starting server", err)
	}
}
