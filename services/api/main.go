package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go-delivery/pb"
	"go-delivery/services/api/accounts"
	"go-delivery/services/api/middlewares"
	"go-delivery/services/api/orders"
	"go-delivery/services/api/sellers"
	"go-delivery/services/api/wallets"
	"go-delivery/util"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"os"
)

var (
	port         int
	accountsAddr string
	walletsAddr  string
	sellersAddr  string
	ordersAddr   string
)

func init() {
	err := godotenv.Load(util.GetEnvFile())
	if err != nil {
		log.Panicln(err)
	}

	flag.IntVar(&port, "port", 6000, "api port")
	flag.StringVar(&accountsAddr, "accounts_addr", "localhost:7500", "accounts service address")
	flag.StringVar(&walletsAddr, "wallets_addr", "localhost:7501", "wallets service address")
	flag.StringVar(&sellersAddr, "sellers_addr", "localhost:7502", "sellers service address")
	flag.StringVar(&ordersAddr, "orders_addr", "localhost:7503", "orders service address")

	flag.Parse()
}

func main() {
	accountsConn, err := grpc.Dial(accountsAddr, grpc.WithInsecure())
	if err != nil {
		log.Panicln(err)
	}
	defer util.HandleClose(accountsConn)

	router := mux.NewRouter().StrictSlash(true)

	accountsClient := pb.NewAccountsServiceClient(accountsConn)
	middlewareGroup := middlewares.New(accountsClient)

	accounts.RegisterAccountsHandlers(accountsClient, middlewareGroup, router)

	walletsConn, err := grpc.Dial(walletsAddr, grpc.WithInsecure())
	if err != nil {
		log.Panicln(err)
	}
	defer util.HandleClose(walletsConn)

	walletsClient := pb.NewWalletsServiceClient(walletsConn)
	wallets.RegisterWalletsHandlers(walletsClient, middlewareGroup, router)

	sellersConn, err := grpc.Dial(sellersAddr, grpc.WithInsecure())
	if err != nil {
		log.Panicln(err)
	}
	defer util.HandleClose(sellersConn)

	productsClient := pb.NewProductsServiceClient(sellersConn)
	sellers.RegisterSellersHandlers(productsClient, middlewareGroup, router)

	ordersConn, err := grpc.Dial(ordersAddr, grpc.WithInsecure())
	if err != nil {
		log.Panicln(err)
	}
	defer util.HandleClose(ordersConn)

	ordersClient := pb.NewOrdersServiceClient(ordersConn)
	orders.RegisterOrdersHandlers(ordersClient, middlewareGroup, router)

	headers := handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "X-Requested-with"})
	methods := handlers.AllowedMethods([]string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete})
	origins := handlers.AllowedOrigins([]string{"*"})

	handler := handlers.CORS(headers, methods, origins)(router)
	handler = handlers.LoggingHandler(os.Stdout, handler)

	log.Printf("Api service running on: [::]:%d\n", port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), handler))
}
