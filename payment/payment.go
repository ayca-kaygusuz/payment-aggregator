package payment

type DepositResponse struct {
	Status        string  `json:"status"`
	TransactionID string  `json:"transactionId"`
	Amount        float64 `json:"amount"`
	Secret        string  `json:"secret,omitempty"`  // Optional field
	Message       string  `json:"message,omitempty"` // Optional field
}

type Aggregator interface {
	InitializeSession() (string, error)
	GetAccounts(token string, amount float64) ([]map[string]interface{}, error) // Added amount parameter
	MakeDeposit(amount float64) (DepositResponse, error)
}
