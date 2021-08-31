package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"go-delivery/db"
	"go-delivery/pb"
	"go-delivery/services/wallets/service"
	"go-delivery/services/wallets/store"
	"go-delivery/util"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

var port int

func init() {
	err := godotenv.Load(util.GetEnvFile())
	if err != nil {
		log.Panicln(err)
	}

	flag.IntVar(&port, "port", 7501, "grpc port")

	flag.Parse()
}

func main() {
	cfg := db.NewConfig()

	log.Println("loading db configs: ", cfg.URI())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbConn := db.New(ctx, cfg)
	defer dbConn.Close(ctx)

	err := dbConn.Ping(ctx)
	if err != nil {
		log.Panicln(err)
	}
	log.Println("database connected successfully")

	walletsStore := store.NewWalletsStore(dbConn.DB())
	walletsService := service.NewService(walletsStore)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Panicln(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWalletsServiceServer(grpcServer, walletsService)

	defer grpcServer.Stop()

	log.Printf("Authentication service running on: [::]:%d\n", port)

	log.Fatalln(grpcServer.Serve(listener))
}
