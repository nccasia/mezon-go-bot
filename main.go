package main

import (
	"context"
	"fmt"
	"mezon-go-bot/config"
	"mezon-go-bot/internal/constants"
	"mezon-go-bot/internal/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	logger := logger.NewLogger(cfg.LogFile)
	defer logger.Sync() // Flush log

	// Start Bot Checkin
	var err error
	bot, err = NewBot(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize bot checkin", zap.Error(err))
	}

	// Register all commands here
	bot.RegisterCmd(constants.NCC8_COMMAND, Ncc8Handler)

	go bot.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Register the health check endpoint
	http.HandleFunc("/health", healthCheckHandler)

	// Define the port
	port := "9098"

	logger.Info("Starting server on port", zap.Any("port", port))

	// Create an HTTP server instance
	server := &http.Server{
		Addr: ":" + port,
	}

	// Start the HTTP server in the main goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Error starting server", zap.Error(err))
		}
	}()

	// Wait for stop signal
	<-stop
	fmt.Println("Received stop signal. Cleaning up...")

	// Clean up before finishing
	cleanup()

	// Gracefully shutdown the server
	fmt.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server Shutdown Failed", zap.Error(err))
	}

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
