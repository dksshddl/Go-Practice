package main

import (
	"Go-Practice/api/respond"
	"Go-Practice/utils"
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type contextKey struct {
	name string
}

type Server struct {
	db *mongo.Client
}

var contextKeyAPIKey = &contextKey{"apikey"}

// contextAPIKey, contextkey 는 private이지만 해당 func로 export 가능해짐
func APIKey(ctx context.Context) (string, bool) {
	key, ok := ctx.Value(contextKeyAPIKey).(string)
	return key, ok
}

func main() {
	id := utils.Setting.Database[0].Id
	pw := utils.Setting.Database[0].Password
	path := utils.Setting.Database[0].Path
	defaultMongoPath := "mongodb+srv://" + id + ":" + pw + "@" + path

	var (
		addr       = flag.String("addr", ":8080", "endpoint address")
		mongo_addr = flag.String("mongo", defaultMongoPath, "mongodb addrss")
	)

	log.Println("Dialing mongo", mongo_addr)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := mongo.Connect(ctx, options.Client().ApplyURI(*mongo_addr))

	if err != nil {
		log.Fatalln(err)
	}
	defer db.Disconnect(ctx)

	s := &Server{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("/polls/", withCORS(withAPIKey(s.handlePolls)))
	log.Println("Starting web server on,", *addr)
	http.ListenAndServe(":8080", mux)
	log.Println("Stopping...")
}

func withAPIKey(fn http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if !isValidAPIKey(key) {
			respond.RespondErr(rw, r, http.StatusUnauthorized, "invalid API key")
			return
		}
		ctx := context.WithValue(r.Context(), contextKeyAPIKey, key)
		fn(rw, r.WithContext(ctx))
	}
}

// 실제 production 에서는 https://github.com/fasterness/cors 확인해보자
func withCORS(fn http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Access-Control-Allow-Origin", "*")
		rw.Header().Set("Access-Control-Expose-Headers", "Location")
		fn(rw, r)
	}
}

func isValidAPIKey(key string) bool {
	return key == "abc123"
}
