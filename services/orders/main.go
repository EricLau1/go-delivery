package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"go-delivery/db"
	"go-delivery/pb"
	"go-delivery/services/orders/service"
	"go-delivery/services/orders/store"
	"go-delivery/services/orders/workflows"
	"go-delivery/util"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

var (
	port         int
	accountsAddr string
	walletsAddr  string
	sellersAddr  string
)

func init() {
	err := godotenv.Load(util.GetEnvFile())
	if err != nil {
		log.Panicln(err)
	}

	flag.StringVar(&accountsAddr, "accounts_addr", "localhost:7500", "accounts service address")
	flag.StringVar(&walletsAddr, "wallets_addr", "localhost:7501", "wallets service address")
	flag.StringVar(&sellersAddr, "sellers_addr", "localhost:7502", "sellers service address")
	flag.IntVar(&port, "port", 7503, "orders service port")

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

	accountsConn, err := grpc.Dial(accountsAddr, grpc.WithInsecure())
	if err != nil {
		log.Panicln(err)
	}
	defer util.HandleClose(accountsConn)

	accountsClient := pb.NewAccountsServiceClient(accountsConn)

	walletsConn, err := grpc.Dial(walletsAddr, grpc.WithInsecure())
	if err != nil {
		log.Panicln(err)
	}
	defer util.HandleClose(walletsConn)

	walletsClient := pb.NewWalletsServiceClient(walletsConn)

	sellersConn, err := grpc.Dial(sellersAddr, grpc.WithInsecure())
	if err != nil {
		log.Panicln(err)
	}
	defer util.HandleClose(sellersConn)

	productsClient := pb.NewProductsServiceClient(sellersConn)

	temporalClient, err := client.NewClient(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer temporalClient.Close()

	ordersStore := store.NewOrdersStore(dbConn.DB())
	ordersService := service.NewService(ordersStore, walletsClient, accountsClient, productsClient)
	ordersWorkflowService := service.NewOrdersWorkflowService(ordersStore, walletsClient, accountsClient, productsClient, temporalClient)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Panicln(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrdersServiceServer(grpcServer, ordersService)
	pb.RegisterOrdersWorkflowServiceServer(grpcServer, ordersWorkflowService)

	defer grpcServer.Stop()

	go workflows.Run(temporalClient)

	log.Printf("Orders service running on: [::]:%d\n", port)

	log.Fatalln(grpcServer.Serve(listener))
}
