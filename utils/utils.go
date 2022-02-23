package utils

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ApplyIndices(db *mongo.Database) {
	var bookIndex = mongo.IndexModel{Options: options.Index().SetUnique(true), Keys: "name"}
	_, err := db.Collection(BOOKS).Indexes().CreateOne(context.Background(), bookIndex)
	if err != nil {
		log.Printf("err creating index on collection %s", BOOKS)
	}
}
