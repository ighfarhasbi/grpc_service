package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/ighfarhasbi/grpc_service/inventory_service/db"
	inventorypb "github.com/ighfarhasbi/grpc_service/proto/inventory"
)

type InventoryService struct {
	inventorypb.UnimplementedInventoryServiceServer
	repo db.InventoryRepository
}

func NewInventoryService(repo db.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo}
}

func (s *InventoryService) CheckStock(ctx context.Context, req *inventorypb.CheckStockRequest) (*inventorypb.CheckStockResponse, error) {
	log.Printf("Checking stock for product %s", req.ProductId)

	stock, err := s.repo.GetStock(ctx, req.ProductId)
	if err != nil {
		return &inventorypb.CheckStockResponse{
			Available: false,
			Error: &inventorypb.ErrorInfo{
				Code:    "PRODUCT_NOT_FOUND",
				Message: "Product not found",
			},
		}, nil
	}

	available := stock >= int(req.Quantity)
	return &inventorypb.CheckStockResponse{
		Available:    available,
		AvailableQty: int32(stock),
	}, nil
}

func (s *InventoryService) ReserveStock(ctx context.Context, req *inventorypb.ReserveStockRequest) (*inventorypb.ReserveStockResponse, error) {
	log.Printf("Reserving stock for order %s product %s", req.OrderId, req.ProductId)

	stock, err := s.repo.GetStock(ctx, req.ProductId)
	if err != nil {
		return &inventorypb.ReserveStockResponse{
			Success: false,
			Error: &inventorypb.ErrorInfo{
				Code:    "PRODUCT_NOT_FOUND",
				Message: "Product not found",
			},
		}, nil
	}

	if stock < int(req.Quantity) {
		return &inventorypb.ReserveStockResponse{
			Success: false,
			Error: &inventorypb.ErrorInfo{
				Code:    "OUT_OF_STOCK",
				Message: "Insufficient stock",
			},
		}, nil
	}

	resID := uuid.New().String()

	err = s.repo.WithTx(ctx, func(tx *sql.Tx) error {
		// Kurangi stok
		if err := s.repo.UpdateStock(ctx, tx, req.ProductId, -int(req.Quantity)); err != nil {
			return err
		}

		// Buat reservasi
		if err := s.repo.CreateReservation(ctx, tx, resID, req.OrderId, req.ProductId, int(req.Quantity)); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to reserve stock: %v", err)
	}

	return &inventorypb.ReserveStockResponse{
		Success:       true,
		ReservationId: resID,
	}, nil
}

func (s *InventoryService) ReleaseStock(ctx context.Context, req *inventorypb.ReleaseStockRequest) (*inventorypb.ReleaseStockResponse, error) {
	log.Printf("Releasing stock for reservation %s", req.ReservationId)

	productID, qty, err := s.repo.GetReservation(ctx, req.ReservationId)
	if err != nil {
		return &inventorypb.ReleaseStockResponse{
			Success: false,
			Error: &inventorypb.ErrorInfo{
				Code:    "RESERVATION_NOT_FOUND",
				Message: "Reservation not found",
			},
		}, nil
	}

	err = s.repo.WithTx(ctx, func(tx *sql.Tx) error {
		// Tambah stok kembali
		if err := s.repo.UpdateStock(ctx, tx, productID, qty); err != nil {
			return err
		}

		// Update status jadi canceled
		if err := s.repo.CancelReservation(ctx, tx, req.ReservationId); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to release stock: %v", err)
	}

	return &inventorypb.ReleaseStockResponse{
		Success: true,
	}, nil
}

func (s *InventoryService) GetProduct(ctx context.Context, req *inventorypb.GetProductRequest) (*inventorypb.GetProductResponse, error) {
	sku, name, price, stock, err := s.repo.GetProduct(ctx, req.ProductId)
	if err != nil {
		return &inventorypb.GetProductResponse{
			Error: &inventorypb.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: err.Error(),
			},
		}, nil
	}

	return &inventorypb.GetProductResponse{
		ProductId: req.ProductId,
		Sku:       sku,
		Name:      name,
		Price:     price,
		Stock:     int32(stock),
	}, nil
}
