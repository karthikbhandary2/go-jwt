package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


func DBInstance() *mongo.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading the .env file")
	}
	MongoDB := os.Getenv("MONGO_URL")
	log.Printf("MongoDB URL: '%s'", MongoDB)
	if MongoDB == "" {
		log.Fatal("MONGO_URL environment variable is empty")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoDB))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("connected to mongo!")
	return client 
}

var Client *mongo.Client = DBInstance()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection *mongo.Collection = client.Database("go-auth").Collection(collectionName)
	return collection
}