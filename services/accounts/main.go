package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"go-delivery/db"
	"go-delivery/pb"
	"go-delivery/services/accounts/service"
	"go-delivery/services/accounts/store"
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

	flag.IntVar(&port, "port", 7500, "grpc port")

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

	usersStore := store.NewUsersStore(dbConn.DB())
	accountsService := service.NewService(usersStore)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Panicln(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAccountsServiceServer(grpcServer, accountsService)

	defer grpcServer.Stop()

	log.Printf("Accounts service running on: [::]:%d\n", port)

	log.Fatalln(grpcServer.Serve(listener))
}
