package factory

import (
	"errors"
	"os"
	"payment-aggregator/internal/logger"
	"payment-aggregator/payment"
	"payment-aggregator/payment/paymentMethods/sansgetirsin"

	"github.com/joho/godotenv"
)

// flexible map factory (registry)
var AggregatorFactories = map[string]func() payment.FlowRunner{
	"sansgetirsin": sansgetirsin.NewFromEnv,
	// other aggregators...
}

// strict but simple factory (single entry point)
// needs AGGREGATOR env var to be changed manually
func AggregatorFromEnv() (payment.FlowRunner, error) {

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		logger.ErrorLogger.Println("Error loading .env file, using environment variables directly")
	}

	method := os.Getenv("AGGREGATOR")

	if factoryFunc, ok := AggregatorFactories[method]; ok {
		return factoryFunc(), nil
	}
	return nil, errors.New("unsupported aggregator: " + method)
}
