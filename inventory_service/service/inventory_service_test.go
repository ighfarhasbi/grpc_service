package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ighfarhasbi/grpc_service/inventory_service/service"
	"github.com/ighfarhasbi/grpc_service/mocks"
	inventorypb "github.com/ighfarhasbi/grpc_service/proto/inventory"
	"github.com/stretchr/testify/assert"
)

func TestCheckStock_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.InventoryRepository)

	// "pemalsuan" data, seolah ada 5 stok untuk product "prod1" di database
	mockRepo.On("GetStock", ctx, "prod1").Return(5, nil)

	// instance service dengan mock repository
	svc := service.NewInventoryService(mockRepo)

	// memanggil service
	resp, err := svc.CheckStock(ctx, &inventorypb.CheckStockRequest{
		ProductId: "prod1",
		Quantity:  3,
	})
	assert.NoError(t, err)                            // pastikan tidak ada error
	assert.True(t, resp.GetAvailable())               // pastikan tersedia (true)
	assert.Equal(t, int32(5), resp.GetAvailableQty()) // pastikan jumlah stok sesuai dengan mock

	mockRepo.AssertExpectations(t) // pastikan semua ekspektasi pada mock terpenuhi
}

func TestCheckStock_ProductNotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.InventoryRepository)

	// Return error for GetStock
	mockRepo.On("GetStock", ctx, "prod2").Return(0, errors.New("product not found"))

	svc := service.NewInventoryService(mockRepo)

	resp, err := svc.CheckStock(ctx, &inventorypb.CheckStockRequest{
		ProductId: "prod2",
		Quantity:  3,
	})
	assert.NoError(t, err)
	assert.False(t, resp.GetAvailable())
	assert.Contains(t, resp.GetError().GetCode(), "PRODUCT_NOT_FOUND")

	mockRepo.AssertExpectations(t)
}
