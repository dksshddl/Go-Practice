package utils

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var Setting Settings

type poll struct {
	Options []string
}

type Option struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Title   string             `bson:title`
	Options []string           `bson:options`
}

type Settings struct {
	Database []struct {
		Name     string `json:"name"`
		Path     string `json:"path"`
		Id       string `json:"id"`
		Password string `json:"pw"`
	} `json:"database"`
}

func init() {
	Setting = ReadSettings()
}

func LoadOptions(client *mongo.Client, ctx *context.Context) []string {
	ops := make([]Option, 0)
	collection := client.Database("ballots").Collection("polls")
	cur, err := collection.Find(*ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(*ctx)
	cur.All(*ctx, &ops)

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	return ops[0].Options
}

func ReadSettings() Settings {
	file, err := os.OpenFile("../settings.json", os.O_RDONLY, os.FileMode(0644))
	if err != nil {
		log.Fatalln(err)
	}
	decoder := json.NewDecoder(file)
	var s Settings
	decoder.Decode(&s)

	return s
}
