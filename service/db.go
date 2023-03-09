package service

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDb() (*mongo.Client, error) {
	uri := "mongodb://db:27017/"
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	var err error
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI))
	if err != nil {
		return nil, err
	}
	var result bson.M
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result); err != nil {
		return nil, err
	}
	log.Println("Pinged your deployment. You successfully connected to MongoDB!")
	return client, nil
}
