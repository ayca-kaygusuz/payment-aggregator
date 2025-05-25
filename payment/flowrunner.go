package payment

import "payment-aggregator/models"

// to direct the flow between interactive vs simple
// without coupling it with main
type FlowRunner interface {
	RunDepositFlow(amount float64) (DepositResponse, models.PaymentModel, error)
}
