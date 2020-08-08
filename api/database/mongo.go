package database

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client *mongo.Client
)

func Connect(databaseUri string) {
	var err error
	Client, err = mongo.NewClient(options.Client().ApplyURI(databaseUri))
	if err != nil {
		log.Fatalf("Error creating mongodb client: %v", err)
	}

	err = Client.Connect(context.Background())
	if err != nil {
		log.Fatalf("Error connecting to mongodb: %v", err)
	}
}

func Disconnect() {
	ctx := context.Background()
	if err := Client.Disconnect(ctx); err != nil {
		panic(err)
	}
}
