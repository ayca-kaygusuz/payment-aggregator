package factory

import (
	"fmt"
	"os"
	"payment-aggregator/payment"
	"payment-aggregator/payment/paymentMethods/sansgetirsin"
)

func FlowRunnerFromEnv() (payment.FlowRunner, error) {

	aggregatorName := os.Getenv("AGGREGATOR")

	if aggregatorName == "" {
		return nil, fmt.Errorf("AGGREGATOR environment variable is missing")
	}

	switch aggregatorName {
	case "sansgetirsin":
		return sansgetirsin.NewFromEnv(), nil

	// Add other aggregators with their custom flow implementations
	default:
		return nil, fmt.Errorf("unsupported aggregator: %s", aggregatorName)
	}
}
