package utils

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Dialdb(dbpath string) (*mongo.Client, *context.Context) {
	var err error
	log.Println("dialing mongodb...")
	// [mongodb:]//[id]:[pw]@[address]
	clientOptions := options.Client().
		ApplyURI(dbpath)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("db connected!")
	return client, &ctx
}

func Closedb(client *mongo.Client, ctx *context.Context) {
	client.Disconnect(*ctx)
	log.Println("closed database connection")
}
