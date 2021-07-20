package main

import (
	_ "Go-Practice/utils"
	"context"
	"flag"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

type contextKey struct {
	name string
}

type Server struct {
	db *mongo.Session
}

var contextKeyAPIKey = &contextKey{"apikey"}

// contextAPIKey, contextkey 는 private이지만 해당 func로 export 가능해짐
func APIKey(ctx context.Context) (string, bool) {
	key, ok := ctx.Value(contextKeyAPIKey).(string)
	return key, ok
}

func main() {
	id := utils.Database[0].Id
	pw := setting.Database[0].Password
	path := setting.Database[0].Path
	defaultMongoPath := "mongodb+srv://" + id + ":" + pw + "@" + path

	var (
		addr  = flag.String("addr", ":8080", "endpoint address")
		mongo = flag.String("mongo", defaultMongoPath, "mongodb addrss")
	)
}

func withAPIKey(fn http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if !isValidAPIKey(key) {
			respondErr(rw, r, http.StatusUnauthorized, "invalid API key")
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
