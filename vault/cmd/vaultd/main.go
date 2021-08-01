package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"vault"
	"vault/pb"

	"google.golang.org/grpc"
)

func main() {
	httpAddr := flag.String("http", ":8080", "http listen address")
	grpcAddr := flag.String("grpc", ":8081", "grpc listen address")

	flag.Parse()
	ctx := context.Background()
	srv := vault.NewService()
	errChan := make(chan error)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	hashEndpoint := vault.MakeHashEndpoint(srv)
	validateEndpoint := vault.MakeValidateEndpoint(srv)
	endpoints := vault.Endpoints{
		HashEndPoint:     hashEndpoint,
		ValidateEndPoint: validateEndpoint,
	}

	go func() {
		log.Println("http: ", *httpAddr)
		handler := vault.NewHTTPServer(ctx, endpoints)
		errChan <- http.ListenAndServe(*httpAddr, handler)
	}()

	go func() {
		listener, err := net.Listen("tcp", *grpcAddr)
		if err != nil {
			errChan <- err
			return
		}
		log.Println("grpc: ", *grpcAddr)
		handler := vault.NewGRPCServer(ctx, endpoints)
		gRPCServer := grpc.NewServer()
		pb.RegisterVaultServer(gRPCServer, handler)
		errChan <- gRPCServer.Serve(listener)
	}()
	log.Fatalln(<-errChan)
}
