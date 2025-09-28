// Package db
package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	clientInstance *mongo.Client
	once           sync.Once
)

func Connect() *mongo.Database {
	once.Do(func() {
		dbLink := os.Getenv("DATABASE_LINK")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbLink))

		if err != nil {
			log.Println("Mongo connect error", err)
		}
		clientInstance = client
	})

	return clientInstance.Database("chat")
}

func InsertOne(coll string, data any) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := Connect()

	return db.Collection(coll).InsertOne(ctx, data)
}

func InsertMany(coll string, data []any) (*mongo.InsertManyResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := Connect()
	return db.Collection(coll).InsertMany(ctx, data)
}

func Find(coll string, key any, result any) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := Connect()
	csr, err := db.Collection(coll).Find(ctx, key)

	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	defer csr.Close(ctx)

	for csr.Next(ctx) {
		if err := csr.Decode(result); err != nil {
			fmt.Println(err.Error())
			break
		}
	}

	return "find success", nil
}

func FindOne(coll string, key any, result any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := Connect()

	return db.Collection(coll).FindOne(ctx, key).Decode(result)
}

func CheckDoc(coll string, key any) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := Connect()
	return db.Collection(coll).CountDocuments(ctx, key)
}

func UpdateOne(coll string, filter any, update any) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := Connect()
	return db.Collection(coll).UpdateOne(ctx, filter, update)
}

func DeleteOne(coll string, key any) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := Connect()
	return db.Collection(coll).DeleteOne(ctx, key)
}
