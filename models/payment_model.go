package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type PaymentModel struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TransactionID   string             `bson:"transaction_id" json:"transaction_id"`
	Amount          float64            `bson:"amount" json:"amount"`
	Status          string             `bson:"status" json:"status"`
	TransactionType string             `bson:"transaction_type" json:"transaction_type"`
	PayerName       string             `bson:"payer_name" json:"payer_name"`
	IBAN            string             `bson:"iban" json:"iban"`
	BankName        string             `bson:"bank_name" json:"bank_name"`
	Aggregator      string             `bson:"aggregator" json:"aggregator"`
	CreatedAt       primitive.DateTime `bson:"created_at" json:"created_at"`
}
