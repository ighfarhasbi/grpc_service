package main

import (
	"log"
	"net"

	"github.com/ighfarhasbi/grpc_service/inventory_service/config"
	"github.com/ighfarhasbi/grpc_service/inventory_service/db"
	"github.com/ighfarhasbi/grpc_service/inventory_service/service"
	inventorypb "github.com/ighfarhasbi/grpc_service/proto/inventory"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func main() {
	// load config
	cfg := config.New()

	// connect to database
	conn, err := db.NewPostgres(cfg.DBUrl) // connect to postgres
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	defer conn.Close()

	// setup gRPC server
	grpcServer := grpc.NewServer()
	repo := db.NewInventoryRepository(conn)
	inventoryServer := service.NewInventoryService(repo)

	inventorypb.RegisterInventoryServiceServer(grpcServer, inventoryServer)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("Inventory gRPC server running on :" + cfg.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
