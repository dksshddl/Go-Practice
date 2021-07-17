package test

import (
	"context"
	"log"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type poll struct {
	Options []string
}
type Option struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Title   string             `bson:title`
	Options []string           `bson:options`
}

func TestMongo(t *testing.T) {
	log.Println("dialing mongodb: localhost")
	// [mongodb:]//[id]:[pw]@[address]
	var ops []Option
	clientOptions := options.Client().
		ApplyURI("mongodb+srv://mongo:mongo@cluster0.56m12.mongodb.net/ballots?retryWrites=true&w=majority")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		t.Errorf("failed to connect to mongo, %s", err)
	}
	collection := client.Database("ballots").Collection("polls")
	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		t.Errorf("failed to find to collection, %s", err)
	}
	defer cur.Close(ctx)
	cur.All(ctx, &ops)

	if len(ops) == 0 {
		t.Errorf("failed to find to data, there is no data.")
	}

	if err := cur.Err(); err != nil {
		t.Errorf("cursor error, %s", err)
	}

}
