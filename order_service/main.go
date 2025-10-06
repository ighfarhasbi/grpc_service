package main

import (
	"log"
	"net"
	"os"

	"github.com/ighfarhasbi/grpc_service/order_service/config"
	"github.com/ighfarhasbi/grpc_service/order_service/db"
	"github.com/ighfarhasbi/grpc_service/order_service/service"
	inventorypb "github.com/ighfarhasbi/grpc_service/proto/inventory"
	orderpb "github.com/ighfarhasbi/grpc_service/proto/order"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if os.Getenv("APP_ENV") == "local" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("Error loading .env: %v", err)
		}
	}

	// load config
	cfg := config.New()

	// connect to database
	conn, err := db.NewOrderPostgres(cfg.DBUrl) // connect to postgres
	if err != nil {
		panic("Failed to connect to order database: " + err.Error())
	}
	defer conn.Close()

	// connect ke inventory gRPC
	// invConn, err := grpc.NewClient(cfg.InventoryServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	invConn, err := grpc.Dial(cfg.InventoryServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to inventory service: %v", err)
	}
	defer invConn.Close()

	// buat client untuk inventory service
	invClient := inventorypb.NewInventoryServiceClient(invConn)

	// setup gRPC server
	grpcServer := grpc.NewServer()
	repo := db.NewOrderRepository(conn)

	// 🔹 inject inventory client ke service
	orderServer := service.NewOrderService(repo, invClient)

	orderpb.RegisterOrderServiceServer(grpcServer, orderServer)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("Order gRPC server running on :" + cfg.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
