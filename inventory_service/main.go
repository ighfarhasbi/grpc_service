package main

import (
	"context"
	"database/sql"
	"log"
	"net"

	inventorypb "github.com/ighfarhasbi/grpc_service/proto/inventory"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

type server struct {
	inventorypb.UnimplementedInventoryServiceServer
	db *sql.DB
}

func (s *server) CheckStock(ctx context.Context, req *inventorypb.CheckStockRequest) (*inventorypb.CheckStockResponse, error) {
	var availableQty int
	err := s.db.QueryRow("SELECT stock FROM products WHERE products_id = $1", req.GetProductId()).Scan(&availableQty)
	if err != nil {
		return &inventorypb.CheckStockResponse{
			Available:    false,
			AvailableQty: 0,
			Error:        &inventorypb.ErrorInfo{Code: "NOT_FOUND", Message: "Product not found"},
		}, nil
	}

	return &inventorypb.CheckStockResponse{
		Available:    availableQty >= int(req.GetQuantity()),
		AvailableQty: int32(availableQty),
	}, nil
}

func main() {
	db, err := sql.Open("postgres", "postgres://ighfarhasbiash:Aa123aA@localhost:5433/inventory_db?sslmode=disable")
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	inventorypb.RegisterInventoryServiceServer(grpcServer, &server{db: db})

	log.Println("Inventory gRPC server running on :9090")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
