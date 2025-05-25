package sansgetirsin

import (
	"bufio"
	"fmt"
	"os"
	"payment-aggregator/internal/logger"
	"payment-aggregator/models"
	"payment-aggregator/payment"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *SansgetirsinAggregator) InteractiveDepositFlow(amount float64) (payment.DepositResponse, models.PaymentModel, error) {

	// Initialize session and get accounts
	token, err := s.InitializeSession()
	if err != nil {
		return payment.DepositResponse{}, models.PaymentModel{}, fmt.Errorf("failed to initialize session: %w", err)
	}

	accounts, err := s.GetAccounts(token, amount)
	if err != nil {
		return payment.DepositResponse{}, models.PaymentModel{}, fmt.Errorf("failed to get accounts: %w", err)
	}

	if len(accounts) == 0 {
		logger.WarningLogger.Println("No accounts found.")
		return payment.DepositResponse{}, models.PaymentModel{}, fmt.Errorf("no accounts available")
	}

	fmt.Println("Available bank accounts:")
	for i, account := range accounts {
		fmt.Printf("Account #%d:\n", i+1)
		if logo, ok := account["logo"].(string); ok {
			fmt.Printf("  Logo: %s\n", logo)
		}
		fmt.Printf("  _id: %s\n", account["_id"])
		fmt.Printf("  Name: %s\n", account["name"])

		if innerAccounts, ok := account["accounts"].([]interface{}); ok && len(innerAccounts) > 0 {
			fmt.Printf("  Accounts: \n")
			for _, innerAccount := range innerAccounts {
				if innerMap, ok := innerAccount.(map[string]interface{}); ok {
					fmt.Printf("    - ID: %s\n", innerMap["_id"])
					if fields, ok := innerMap["fields"].([]interface{}); ok {
						for _, field := range fields {
							if fieldMap, ok := field.(map[string]interface{}); ok {
								fmt.Printf("      %s: %s\n", fieldMap["name"], fieldMap["value"])
							}
						}
					}
				}
			}
		}

		fmt.Println("---")
	}

	fmt.Print("Enter the account number to use: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	choice, err := strconv.Atoi(input[:len(input)-1])
	if err != nil || choice < 1 || choice > len(accounts) {
		return payment.DepositResponse{}, models.PaymentModel{}, fmt.Errorf("invalid selection")
	}

	selected := accounts[choice-1]
	accountList, ok := selected["accounts"].([]interface{})
	if !ok || len(accountList) == 0 {
		return payment.DepositResponse{}, models.PaymentModel{}, fmt.Errorf("no inner accounts found")
	}

	accountMap, ok := accountList[0].(map[string]interface{})
	if !ok {
		return payment.DepositResponse{}, models.PaymentModel{}, fmt.Errorf("invalid account format")
	}

	bankID, ok := accountMap["_id"].(string)
	if !ok {
		return payment.DepositResponse{}, models.PaymentModel{}, fmt.Errorf("bank ID is not a string")
	}

	extraData := map[string]interface{}{
		"description": "Test deposit",
	}

	resp, err := s.MakeDepositWithData(token, bankID, amount, extraData)
	if err != nil {
		return payment.DepositResponse{}, models.PaymentModel{}, err
	}

	// Extract fields from selected accountMap
	payerName := ""
	iban := ""
	bankName, _ := selected["name"].(string)

	if fields, ok := accountMap["fields"].([]interface{}); ok {
		for _, field := range fields {
			if fieldMap, ok := field.(map[string]interface{}); ok {
				name := fmt.Sprintf("%v", fieldMap["name"])
				value := fmt.Sprintf("%v", fieldMap["value"])

				switch name {
				case "IBAN":
					iban = value
				case "Name", "Full Name", "Payer", "Account Holder":
					payerName = value
				}
			}
		}
	}

	// Construct full payment model
	paymentDoc := models.PaymentModel{
		TransactionID:   resp.TransactionID,
		Amount:          resp.Amount,
		Status:          resp.Status,
		TransactionType: "deposit",
		PayerName:       payerName,
		Aggregator:      "Sans Getirsin",
		IBAN:            iban,
		BankName:        bankName,
		CreatedAt:       primitive.NewDateTimeFromTime(time.Now()),
	}

	return resp, paymentDoc, nil

}
