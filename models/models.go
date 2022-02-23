package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Book struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	Price       int64              `json:"price" bson:"price"`
	Author      string             `json:"author" bson:"author"`
	PublishedAt time.Time          `json:"published_at" bson:"published_at"`
	CreatedAt   time.Time          `json:"-" bson:"created_at"`
}
