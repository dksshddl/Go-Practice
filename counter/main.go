package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nsqio/go-nsq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	updateDuration = 1 * time.Second
)

var (
	fatalErr   error
	counts     map[string]int
	countsLock sync.Mutex
)

type Settings struct {
	Database []struct {
		Name     string `json:"name"`
		Path     string `json:"path"`
		Id       string `json:"id"`
		Password string `json:"pw"`
	} `json:"database"`
}

func fatal(e error) {
	fmt.Println(e)
	flag.PrintDefaults()
	fatalErr = e
}

func doCount(countsLock *sync.Mutex, counts *map[string]int, pollData *mongo.Collection) {
	countsLock.Lock()
	defer countsLock.Unlock()
	if len(*counts) == 0 {
		log.Println("No new votes, skipping database update")
		return
	}
	log.Println("Updating database...")
	log.Println(*counts)
	ok := true
	for option, count := range *counts {
		sel := bson.M{"options": bson.M{"$in": []string{option}}}
		up := bson.M{"$inc": bson.M{"results." + option: count}}
		if _, err := pollData.UpdateMany(context.Background(), sel, up); err != nil {
			log.Fatalln("failed to update: ", err)
			ok = false
		}
	}
	if ok {
		log.Println("Finished updating database...")
		*counts = nil
	}
}

func main() {

	file, err := os.OpenFile("../settings.json", os.O_RDONLY, os.FileMode(0644))
	if err != nil {
		log.Fatalln(err)
	}
	decoder := json.NewDecoder(file)
	var s Settings
	decoder.Decode(&s)

	defer func() {
		if fatalErr != nil {
			os.Exit(1)
		}
	}()

	id := s.Database[0].Id
	pw := s.Database[0].Password
	path := s.Database[0].Path

	dbpath := "mongodb+srv://" + id + ":" + pw + "@" + path + "/ballots?retryWrites=true&w=majority"

	log.Println("Connection to database... ", dbpath)
	q, err := nsq.NewConsumer("votes", "counter", nsq.NewConfig())
	q.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		countsLock.Lock()
		defer countsLock.Unlock()
		if counts == nil {
			counts = make(map[string]int)
		}
		vote := string(message.Body)
		counts[vote]++
		return nil
	}))

	if err := q.ConnectToNSQLookupd("localhost:4161"); err != nil {
		log.Fatalln("nsq connection fail: ", err)
		return
	}
	clientOptions := options.Client().
		ApplyURI(dbpath)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, clientOptions)

	pollData := client.Database("ballots").Collection("polls")

	if err != nil {
		log.Fatalln("db connection fail: ", err)
		return
	}
	defer func() {
		log.Println("Closing database connection...")
		cancel()
	}()

	ticker := time.NewTicker(updateDuration)
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		select {
		case <-ticker.C:
			doCount(&countsLock, &counts, pollData)
		case <-termChan:
			ticker.Stop()
			q.Stop()
		case <-q.StopChan:
			return
		}
	}
}
