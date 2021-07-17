package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nsqio/go-nsq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// var db *mgo.Session
var (
	client *mongo.Client
	ctx    context.Context
	ops    []Option
)

type poll struct {
	Options []string
}

type Option struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Title   string             `bson:title`
	Options []string           `bson:options`
}

func loadOptions() ([]string, error) {

	collection := client.Database("ballots").Collection("polls")
	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)
	cur.All(ctx, &ops)

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	return ops[0].Options, cur.Err()
}

func dialdb(dbpath string) error {
	var err error
	log.Println("dialing mongodb: localhost")
	// [mongodb:]//[id]:[pw]@[address]
	clientOptions := options.Client().
		ApplyURI("mongodb+srv://mongo:mongo@cluster0.56m12.mongodb.net/ballots?retryWrites=true&w=majority")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Println("fail to connecting mongodb serve")
	}

	return err
}

func closedb() {
	client.Disconnect(ctx)
	log.Println("closed database connection")
}

func publishVotes(votes <-chan string) <-chan struct{} {
	stopchan := make(chan struct{}, 1)
	pub, _ := nsq.NewProducer("localhost:4150", nsq.NewConfig())
	go func() {
		for vote := range votes {
			pub.Publish("votes", []byte(vote))
		}
		log.Println("Publisher: Stopping")
		pub.Stop()
		log.Println("Publisher: Stopped")
		stopchan <- struct{}{}
	}()
	return stopchan
}

func main() {
	var stoplock sync.Mutex

	dbpath := "mongodb+srv://mongo:mongo@cluster0.56m12.mongodb.net"

	stop := false
	stopChan := make(chan struct{}, 1)
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		stoplock.Lock()
		stop = true
		stoplock.Unlock()
		log.Println("Stopping...")
		stopChan <- struct{}{}
		closeConn()
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	if err := dialdb(dbpath); err != nil {
		log.Fatalln("failed to dial MongoDB: ", err)
	}

	votes := make(chan string)
	publisherStoppedChan := publishVotes(votes)
	twitterStoppedChan := startTwitterStream(stopChan, votes)
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			closeConn()
			stoplock.Lock()
			if stop {
				stoplock.Unlock()
				return
			}
			stoplock.Unlock()
		}
	}()
	<-twitterStoppedChan
	close(votes)
	<-publisherStoppedChan
}
