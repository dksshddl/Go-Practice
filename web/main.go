package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	var addr = flag.String("addr", "localhost:8081", "web addr")
	flag.Parse()
	mux := http.NewServeMux()
	mux.Handle("/", http.StripPrefix("/",
		http.FileServer(http.Dir("pulib"))))
	log.Println("Serving website at: ", *addr)
	http.ListenAndServe(*addr, mux)
}
