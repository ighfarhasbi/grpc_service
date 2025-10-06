package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/ighfarhasbi/grpc_service/inventory_service/service"
	"github.com/ighfarhasbi/grpc_service/mocks"
	inventorypb "github.com/ighfarhasbi/grpc_service/proto/inventory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestReserveStock_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.InventoryRepository)

	// ada 5 stok, diminta 3
	mockRepo.On("GetStock", ctx, "prod4").Return(5, nil)

	// ekspektasi pemanggilan update stock
	mockRepo.On("UpdateStock", ctx, mock.Anything, "prod4", -3).Return(nil)
	// ekspektasi pemanggilan create reservation
	mockRepo.On("CreateReservation", ctx, mock.Anything, mock.Anything, mock.Anything, "prod4", 3).Return(nil)

	// ekspektasi pemanggilan transaksi
	mockRepo.On("WithTx", ctx, mock.AnythingOfType("func(*sql.Tx) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(*sql.Tx) error)
			_ = fn(nil) // jalankan callback supaya UpdateStock & CreateReservation terpanggil
		}).
		Return(nil)

	svc := service.NewInventoryService(mockRepo)

	resp, err := svc.ReserveStock(ctx, &inventorypb.ReserveStockRequest{
		OrderId:   "order2",
		ProductId: "prod4",
		Quantity:  3,
	})
	assert.NoError(t, err)
	assert.True(t, resp.GetSuccess())
	assert.NotEmpty(t, resp.GetReservationId())

	mockRepo.AssertExpectations(t)
}

func TestReserveStock_InsufficientStock(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.InventoryRepository)

	// ada 2 stok, tapi diminta 3
	mockRepo.On("GetStock", ctx, "prod3").Return(2, nil)

	svc := service.NewInventoryService(mockRepo)

	resp, err := svc.ReserveStock(ctx, &inventorypb.ReserveStockRequest{
		OrderId:   "order1",
		ProductId: "prod3",
		Quantity:  3,
	})
	assert.NoError(t, err)
	assert.False(t, resp.GetSuccess())
	assert.Contains(t, resp.GetError().GetCode(), "OUT_OF_STOCK")

	mockRepo.AssertExpectations(t)
}

func TestReleaseStock_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.InventoryRepository)

	// ada reservasi dengan product "prod5" sejumlah 4
	mockRepo.On("GetReservation", ctx, "res1").Return("prod5", 4, nil)

	// ekspektasi pemanggilan update stock (mengembalikan 4 stok)
	mockRepo.On("UpdateStock", ctx, mock.Anything, "prod5", 4).Return(nil)

	// ekspektasi pemanggilan cancel reservation
	mockRepo.On("CancelReservation", ctx, mock.Anything, "res1").Return(nil)

	// ekspektasi pemanggilan transaksi
	mockRepo.On("WithTx", ctx, mock.AnythingOfType("func(*sql.Tx) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(*sql.Tx) error)
			_ = fn(nil) // jalankan callback supaya UpdateStock terpanggil
		}).
		Return(nil)

	svc := service.NewInventoryService(mockRepo)

	resp, err := svc.ReleaseStock(ctx, &inventorypb.ReleaseStockRequest{
		ReservationId: "res1",
	})
	assert.NoError(t, err)
	assert.True(t, resp.GetSuccess())

	mockRepo.AssertExpectations(t)
}

func TestReleaseStock_ReservationNotFound(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.InventoryRepository)

	// Return error for GetReservation
	mockRepo.On("GetReservation", ctx, "res2").Return("", 0, errors.New("reservation not found"))

	svc := service.NewInventoryService(mockRepo)

	resp, err := svc.ReleaseStock(ctx, &inventorypb.ReleaseStockRequest{
		ReservationId: "res2",
	})
	assert.NoError(t, err)
	assert.False(t, resp.GetSuccess())
	assert.Contains(t, resp.GetError().GetCode(), "RESERVATION_NOT_FOUND")

	mockRepo.AssertExpectations(t)
}
