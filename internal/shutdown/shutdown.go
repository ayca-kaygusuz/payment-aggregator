package shutdown

import (
	"os"
	"os/signal"
	"payment-aggregator/internal/database"
	"payment-aggregator/internal/logger"
	"syscall"
)

// WaitForShutdown listens for OS signals and handles graceful shutdown.
func WaitForShutdown(db *database.Database) {
	// Create a channel to listen for OS signals (e.g., SIGINT, SIGTERM)
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal
	<-signalChannel

	// Handle graceful shutdown
	logger.InfoLogger.Println("Shutdown signal received, closing connections and cleaning up...")

	// Perform cleanup actions, such as closing the database connection
	err := db.Close()
	if err != nil {
		logger.ErrorLogger.Printf("Error during shutdown: %v", err)
	} else {
		logger.InfoLogger.Println("Database connection closed successfully.")
	}
}
