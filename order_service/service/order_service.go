package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/ighfarhasbi/grpc_service/order_service/db"
	inventorypb "github.com/ighfarhasbi/grpc_service/proto/inventory"
	orderpb "github.com/ighfarhasbi/grpc_service/proto/order"
)

type OrderService struct {
	orderpb.UnimplementedOrderServiceServer
	repo      db.OrderRepository
	invClient inventorypb.InventoryServiceClient
}

func NewOrderService(repo db.OrderRepository, invClient inventorypb.InventoryServiceClient) *OrderService {
	return &OrderService{
		repo:      repo,
		invClient: invClient,
	}
}

// CreateOrder: insert order + call ReserveStock di inventory
func (s *OrderService) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	orderID := uuid.New().String()
	var totalAmount float64

	err := s.repo.WithTx(ctx, func(tx *sql.Tx) error {
		//  hitung total amount
		for _, item := range req.Items {
			productResp, err := s.invClient.GetProduct(ctx, &inventorypb.GetProductRequest{
				ProductId: item.ProductId,
			})
			if err != nil {
				return fmt.Errorf("failed to call GetProduct for product_id=%s: %w", item.ProductId, err)
			}
			if productResp.Error != nil {
				return fmt.Errorf("inventory error: %s", productResp.Error.Message)
			}
			totalAmount += productResp.Price * float64(item.Quantity)
		}
		if err := s.repo.CreateOrder(ctx, tx, orderID, req.UserId, totalAmount, "pending"); err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}
		// Loop semua item, ambil harga aktual dari inventory service
		for _, item := range req.Items {
			productResp, err := s.invClient.GetProduct(ctx, &inventorypb.GetProductRequest{
				ProductId: item.ProductId,
			})
			if err != nil {
				return fmt.Errorf("failed to call GetProduct for product_id=%s: %w", item.ProductId, err)
			}
			if productResp.Error != nil {
				return fmt.Errorf("inventory error: %s", productResp.Error.Message)
			}

			unitPrice := productResp.Price

			orderItemID := uuid.NewString()

			// Insert order item
			if err := s.repo.CreateOrderItem(ctx, tx, orderItemID, orderID, item.ProductId, int(item.Quantity), unitPrice); err != nil {
				return fmt.Errorf("failed to create order item: %w", err)
			}

			// Reserve stok di inventory
			_, err = s.invClient.ReserveStock(ctx, &inventorypb.ReserveStockRequest{
				OrderId:   orderID,
				ProductId: item.ProductId,
				Quantity:  item.Quantity,
			})
			if err != nil {
				return fmt.Errorf("failed to reserve stock for product %s: %w", item.ProductId, err)
			}
		}

		// Update status jadi reserved (untuk sekarang tetap pending, karena enum status hanya pending, confirmed, canceled)
		if err := s.repo.UpdateOrderStatus(ctx, tx, orderID, "pending"); err != nil {
			return fmt.Errorf("failed to update status: %w", err)
		}

		return nil
	})

	if err != nil {
		return &orderpb.CreateOrderResponse{
			Success: false,
			Error: &orderpb.ErrorInfo{
				Code:    "CREATE_FAILED",
				Message: err.Error(),
			},
		}, nil
	}

	return &orderpb.CreateOrderResponse{
		OrderId: orderID,
		Success: true,
	}, nil
}

// CancelOrder: ubah status + kembalikan stok di inventory
func (s *OrderService) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.CancelOrderResponse, error) {
	order, items, err := s.repo.GetOrder(ctx, req.OrderId)
	if err != nil {
		return &orderpb.CancelOrderResponse{
			Success: false,
			Error: &orderpb.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "order not found",
			},
		}, nil
	}

	if order.Status == "canceled" {
		return &orderpb.CancelOrderResponse{
			Success: false,
			Error: &orderpb.ErrorInfo{
				Code:    "ALREADY_CANCELED",
				Message: "order already canceled",
			},
		}, nil
	}

	// Loop semua item dan release stok di inventory
	for _, item := range items {
		resp, err := s.invClient.ReleaseStock(ctx, &inventorypb.ReleaseStockRequest{
			ReservationId: req.OrderId,
		})
		if err != nil {
			return &orderpb.CancelOrderResponse{
				Success: false,
				Error: &orderpb.ErrorInfo{
					Code:    "RELEASE_FAILED",
					Message: fmt.Sprintf("failed to release stock for product %s: %v", item.ProductID, err),
				},
			}, nil
		}
		if resp.Error != nil {
			return &orderpb.CancelOrderResponse{
				Success: false,
				Error: &orderpb.ErrorInfo{
					Code:    "RELEASE_ERROR",
					Message: fmt.Sprintf("inventory error: %s", resp.Error.Message),
				},
			}, nil
		}
	}

	// Update status order di database
	err = s.repo.WithTx(ctx, func(tx *sql.Tx) error {
		return s.repo.UpdateOrderStatus(ctx, tx, req.OrderId, "canceled")
	})
	if err != nil {
		return &orderpb.CancelOrderResponse{
			Success: false,
			Error: &orderpb.ErrorInfo{
				Code:    "DB_UPDATE_FAILED",
				Message: fmt.Sprintf("failed to update order status: %v", err),
			},
		}, nil
	}

	return &orderpb.CancelOrderResponse{
		Success: true,
	}, nil
}

// GetOrder: ambil data order + item
func (s *OrderService) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {
	order, items, err := s.repo.GetOrder(ctx, req.OrderId)
	if err != nil {
		return &orderpb.GetOrderResponse{
			Error: &orderpb.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "order not found",
			},
		}, nil
	}

	var pbItems []*orderpb.OrderItem
	for _, it := range items {
		pbItems = append(pbItems, &orderpb.OrderItem{
			ProductId: it.ProductID,
			Quantity:  int32(it.Qty),
		})
	}

	return &orderpb.GetOrderResponse{
		OrderId: order.OrderID,
		UserId:  order.UserID,
		Status:  order.Status,
		Items:   pbItems,
	}, nil
}
