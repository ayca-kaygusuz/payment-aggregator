package database

import (
	"context"
	"time"

	"payment-aggregator/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// NewDatabase initializes a new MongoDB connection and returns a Database instance.
func NewDatabase(uri, dbName, collectionName string) (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	collection := client.Database(dbName).Collection(collectionName)

	return &Database{
		client:     client,
		collection: collection,
	}, nil
}

// InsertPayment inserts a new payment record into the database.
func (db *Database) InsertPayment(payment models.PaymentModel) error {
	payment.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	_, err := db.collection.InsertOne(context.Background(), payment)
	return err
}

// Close cleans up the database connection.
func (db *Database) Close() error {
	return db.client.Disconnect(context.Background())
}
