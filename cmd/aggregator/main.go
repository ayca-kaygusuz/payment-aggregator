package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"payment-aggregator/internal/database"
	"payment-aggregator/internal/factory"
	"payment-aggregator/internal/logger"
	"payment-aggregator/internal/shutdown"

	"github.com/joho/godotenv"
)

func main() {
	// Initialize logger
	logger.InitLogger()

	// Load .env file
	envFile := filepath.Join(".", ".env")
	log.Println("Looking for .env file at:", envFile)
	err := godotenv.Load(envFile)
	if err != nil {
		log.Println("No .env file found in the root, using system environment variables.")
	} else {
		log.Println(".env file loaded successfully.")
	}

	// Initialize MongoDB connection
	databaseURI := os.Getenv("DATABASE_PROTOCOL") + os.Getenv("DATABASE_BASE") + ":" + os.Getenv("DATABASE_PORT") + "/" + os.Getenv("DATABASE_NAME")

	logger.InfoLogger.Println("Connecting to MongoDB at:", databaseURI)

	db, err := database.NewDatabase(databaseURI, os.Getenv("DATABASE_NAME"), "Payments")
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer db.Close()

	// Start the server to handle callbacks
	go database.StartServer(db)

	// Start the flow
	flow, err := factory.FlowRunnerFromEnv()
	if err != nil {
		logger.ErrorLogger.Fatalf("Failed to get flow runner: %v", err)
	}

	// Make a deposit flow
	response, responseModel, err := flow.RunDepositFlow(100.0)
	if err != nil {
		logger.ErrorLogger.Printf("Deposit failed: %v", err)
		os.Exit(1)
	}

	// Log if the deposit was successful
	logger.InfoLogger.Printf("Deposit succeeded: %+v", response)

	// Insert payment into MongoDB
	err = db.InsertPayment(responseModel)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to insert payment: %v", err)
		os.Exit(1)
	} else {
		logger.InfoLogger.Println("Payment inserted successfully.")
	}

	// HTTP server to handle callbacks
	go func() {
		callbackURL := os.Getenv("CALLBACK_URL")

		// Encode responseModel into JSON
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(responseModel)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to encode callback payload: %v", err)
			return
		}

		// Send the POST request
		resp, err := http.Post(callbackURL, "application/json", &buf)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to GET callback: %v", err)
			return
		}
		defer resp.Body.Close()

		logger.InfoLogger.Printf("Callback GET to %s, response: %s", callbackURL, resp.Status)
	}()

	// TODO: Make a withdrawal flow

	// Shutdown
	shutdown.WaitForShutdown(db)
}
