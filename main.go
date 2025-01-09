package main

import (
	"fmt"
	"mezon-go-bot/config"
	"mezon-go-bot/internal/constants"
	"mezon-go-bot/internal/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	go bot.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Register the health check endpoint
	http.HandleFunc("/health", healthCheckHandler)

	// Define the port
	port := "9098"

	log.Info("Starting server on port", zap.Any("port", port))

	// Start the HTTP server in the main goroutine
	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal("Error starting server", zap.Error(err))
		}
	}()

	// Wait for stop signal
	<-stop
	fmt.Println("Received stop signal. Cleaning up...")

	// Clean up before finishing
	cleanup()

	// Stop server
	fmt.Println("Server stopped.")
}

func cleanup() {
	if bot != nil {
		fmt.Println("Closing bot and cleaning up resources...")
		bot.Stop()
	}
	fmt.Println("Cleanup completed.")
}
