package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	"vault"

	grpcclient "vault/client/grpc"

	"google.golang.org/grpc"
)

func main() {
	var (
		grpcAddr = flag.String("addr", ":8081", "grpc address")
	)
	flag.Parse()
	ctx := context.Background()

	conn, err := grpc.Dial(*grpcAddr, grpc.WithInsecure(), grpc.WithTimeout(1*time.Second))

	if err != nil {
		log.Fatalln("gRPC dial: ", err)
	}
	log.Printf("gPRC dial success at: %s\n", *grpcAddr)
	defer conn.Close()
	vaultServer := grpcclient.New(conn)
	args := flag.Args()

	var cmd string
	cmd, args = pop(args)
	switch cmd {
	case "hash":
		var password string
		password, args = pop(args)
		hash(ctx, vaultServer, password)
	case "validate":
		var password, hash string
		password, args = pop(args)
		hash, args = pop(args)
		validate(ctx, vaultServer, password, hash)
	default:
		log.Fatalln("unknown command ", cmd)
	}
}

func pop(s []string) (string, []string) {
	if len(s) == 0 {
		return "", s
	}
	return s[0], s[1:]
}

func hash(ctx context.Context, service vault.Service, password string) {
	h, err := service.Hash(ctx, password)
	if err != nil {
		log.Fatalln(err.Error())
	}
	fmt.Println(h)
}

func validate(ctx context.Context, service vault.Service, password, hash string) {
	valid, err := service.Validate(ctx, password, hash)
	if err != nil {
		log.Fatalln(err.Error())
	}
	if !valid {
		fmt.Println("invalid")
		os.Exit(1)
	}
	fmt.Println("valid")
}
