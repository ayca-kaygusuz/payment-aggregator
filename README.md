# Payment Aggregator

This project implements a payment aggregator that supports multiple payment methods.

## Getting Started

1.  Install Go: [https://go.dev/dl/](https://go.dev/dl/)
2.  Clone the repository: `git clone <repository_url>`
3.  Navigate to the project directory: `cd payment-aggregator`
4.  Initialize the Go module: `go mod init payment-aggregator`
5.  Build and run the application: `go run ./cmd/aggregator`

## Configuration

Configuration settings (e.g., API keys, aggregator URLs) can be set in the `internal/config/config.go` file or through environment variables.

## Adding a New Payment Method

1.  Create a new directory under `payment/methods/` for the new payment method (e.g., `payment/methods/newaggregator`).
2.  Implement the `Aggregator` interface in a file within that directory (e.g., `payment/methods/newaggregator/newaggregator.go`).
3.  Add a new case to the `switch` statement in `cmd/aggregator/main.go` to instantiate the new aggregator.

## Dependencies