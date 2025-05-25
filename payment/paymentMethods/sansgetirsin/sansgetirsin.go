package sansgetirsin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"payment-aggregator/internal/logger"
	"payment-aggregator/models"
	"payment-aggregator/payment"
	"strconv"
)

type SansgetirsinAggregator struct {
	BaseURL        string
	Username       string
	APIKey         string
	AdditionalData map[string]interface{}
}

func NewSansgetirsinAggregator(baseURL string) *SansgetirsinAggregator {
	return &SansgetirsinAggregator{BaseURL: baseURL}
}

var _ payment.Aggregator = &SansgetirsinAggregator{}

// NewFromEnv creates a new SansgetirsinAggregator instance from environment variables
func NewFromEnv() payment.FlowRunner {
	baseURL := fmt.Sprintf("https://api-%s.sansgetirsin.com", os.Getenv("SANSGETIRSIN_KEY"))
	logger.InfoLogger.Println("Constructed BaseURL:", baseURL)
	return &SansgetirsinAggregator{
		BaseURL:  baseURL,
		Username: os.Getenv("SANSGETIRSIN_USERNAME"),
		APIKey:   os.Getenv("SANSGETIRSIN_API_KEY"),
		AdditionalData: map[string]interface{}{
			"userId":           os.Getenv("SANSGETIRSIN_USER_ID"),
			"paymentMethod":    mustParseFloat(os.Getenv("SANSGETIRSIN_PAYMENT_METHOD"), 1),
			"maxWithdrawLimit": mustParseFloat(os.Getenv("SANSGETIRSIN_MAX_WITHDRAW_LIMIT"), 1000),
		},
	}
}

// helper to safely parse a string into an int,
// falling back to a default if parsing fails.
func mustParseFloat(s string, def float64) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return f
}

// zero-arg InitializeSession method that uses the one with arguments as a helper
func (s *SansgetirsinAggregator) InitializeSession() (string, error) {
	return s.InitializeSessionWithParams(s.Username, s.APIKey, s.AdditionalData)
}

// initialize session with args
func (s *SansgetirsinAggregator) InitializeSessionWithParams(username, apiKey string, additionalData map[string]interface{}) (string, error) {
	logger.InfoLogger.Println("Sansgetirsin: Initializing session...")
	client := &http.Client{}
	sessionURL := s.BaseURL + "/payment/json"

	requestBody, err := json.Marshal(map[string]interface{}{
		"username":       username,
		"apiKey":         apiKey,
		"additionalData": additionalData,
	})
	if err != nil {
		logger.ErrorLogger.Printf("Sansgetirsin: Failed to marshal session request: %v", err)
		return "", fmt.Errorf("failed to marshal session request: %w", err)
	}

	req, err := http.NewRequest("POST", sessionURL, bytes.NewBuffer(requestBody))
	if err != nil {
		logger.ErrorLogger.Printf("Sansgetirsin: Failed to create session request: %v", err)
		return "", fmt.Errorf("failed to create session request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.ErrorLogger.Printf("Sansgetirsin: Session request failed: %v", err)
		return "", fmt.Errorf("session request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorLogger.Printf("Sansgetirsin: Failed to read session response: %v", err)
		return "", fmt.Errorf("failed to read session response: %w", err)
	}

	var jsonResponse map[string]interface{}
	err = json.Unmarshal(respBody, &jsonResponse)
	if err != nil {
		logger.ErrorLogger.Printf("Sansgetirsin: Failed to unmarshal session response: %v", err)
		return "", fmt.Errorf("failed to unmarshal session response: %w", err)
	}

	data, ok := jsonResponse["data"].(map[string]interface{})
	if !ok {
		logger.ErrorLogger.Println("Sansgetirsin: Data field not found in response")
		return "", fmt.Errorf("data field not found in response")
	}

	token, ok := data["token"].(string)
	if !ok {
		logger.ErrorLogger.Println("Sansgetirsin: Session token not found in response data")
		return "", fmt.Errorf("session token not found in response data")
	}

	logger.InfoLogger.Println("Sansgetirsin: Session initialized successfully.")
	return token, nil
}

func (s *SansgetirsinAggregator) GetAccounts(token string, amount float64) ([]map[string]interface{}, error) {
	logger.InfoLogger.Println("Sansgetirsin: Getting accounts...")
	client := &http.Client{}

	// Construct the request URL (adjust based on API docs)
	accountsURL := fmt.Sprintf("%s/payment/deposit?amount=%.2f", s.BaseURL, amount)

	req, err := http.NewRequest("GET", accountsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token) // Add Authorization header

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	logger.InfoLogger.Printf("Sansgetirsin: Raw accounts response: %s", string(body))

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for errors in the response
	if errorMsg, ok := response["error"].(string); ok {
		logger.ErrorLogger.Printf("Sansgetirsin API error: %s", errorMsg)
		return nil, fmt.Errorf("sansgetirsin API error: %s", errorMsg)
	}

	// Extract the 'data' field
	data, ok := response["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("'data' field not found or not an array in response")
	}

	accounts := make([]map[string]interface{}, len(data))
	for i, account := range data {
		accountMap, ok := account.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("element in 'data' array is not a map")
		}
		accounts[i] = accountMap
	}

	return accounts, nil
}

// MakeDeposit considers only deposit amount
func (s *SansgetirsinAggregator) MakeDeposit(amount float64) (payment.DepositResponse, error) {
	return s.MakeDepositWithData("", "", amount, nil)
}

// MakeDepositWithData makes a deposit to the specified bank account
func (s *SansgetirsinAggregator) MakeDepositWithData(token string, bankID string, amount float64, extraData map[string]interface{}) (payment.DepositResponse, error) {
	logger.InfoLogger.Println("Sansgetirsin: Making deposit...")

	// Construct the request payload (adjust based on API docs)
	payload := map[string]interface{}{
		"bankAccount": bankID, // Use "bankAccount" instead of "bankId" if required
		"amount":      amount,
		// Include extra data if needed
		"extraData": extraData,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return payment.DepositResponse{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// **LOG THE REQUEST PAYLOAD**
	logger.InfoLogger.Printf("Sansgetirsin: Deposit request payload: %s", string(payloadBytes))

	// Construct the request URL (adjust based on API docs)
	depositURL := fmt.Sprintf("%s/payment/deposit", s.BaseURL)

	// Create the HTTP request
	req, err := http.NewRequest("POST", depositURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return payment.DepositResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return payment.DepositResponse{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return payment.DepositResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	logger.InfoLogger.Printf("Sansgetirsin: Raw deposit response: %s", string(body))

	// Unmarshal the response into a map
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return payment.DepositResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for errors in the response
	if errorMsg, ok := response["error"].(string); ok {
		// Handle the error response
		logger.ErrorLogger.Printf("Sansgetirsin API error: %v", errorMsg)
		return payment.DepositResponse{}, fmt.Errorf("sansgetirsin API error: %v", errorMsg)
	}

	// Extract data from the response
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return payment.DepositResponse{}, fmt.Errorf("data not found in response")
	}

	transactionID, ok := data["transactionId"].(string)
	if !ok {
		return payment.DepositResponse{}, fmt.Errorf("transactionId not found in response")
	}

	// The API doesn't return status or message on success, so we'll use some defaults
	status := "success" // Or another appropriate default
	message := "Deposit successful"

	// Create the DepositResponse
	depositResponse := payment.DepositResponse{
		Status:        status,
		TransactionID: transactionID,
		Message:       message,
	}

	// Log the deposit response
	logger.InfoLogger.Printf("Sansgetirsin: Deposit response: %+v", depositResponse)

	return depositResponse, nil
}

func (s *SansgetirsinAggregator) RunDepositFlow(amount float64) (payment.DepositResponse, models.PaymentModel, error) {
	return s.InteractiveDepositFlow(amount)
}
