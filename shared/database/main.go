package database

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

func init() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("Failed to connect to mongo: {}", err)
	}
	log.Println("Connected to mongo")
	mongoClient = client
}

// GetMongoClient returns the MongoDB client instance.
func GetMongoClient() *mongo.Client {
	return mongoClient
}
func GetDB() *mongo.Database {
	return mongoClient.Database("alerts")
}
