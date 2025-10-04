package db

import (
	"context"
	"database/sql"
	"errors"
)

type InventoryRepository interface {
	GetStock(ctx context.Context, productID string) (int, error)
	UpdateStock(ctx context.Context, productID string, delta int) error
	CreateReservation(ctx context.Context, reservationId, orderID, productID string, qty int) error
	GetReservation(ctx context.Context, reservationID string) (string, int, error)
	CancelReservation(ctx context.Context, reservationID string) error
}

type inventoryRepo struct {
	db *sql.DB
}

func NewInventoryRepository(db *sql.DB) InventoryRepository {
	return &inventoryRepo{db: db}
}

func (r *inventoryRepo) GetStock(ctx context.Context, productID string) (int, error) {
	var stock int
	err := r.db.QueryRowContext(ctx, `SELECT stock FROM products WHERE products_id = $1`, productID).Scan(&stock)
	if err == sql.ErrNoRows {
		return 0, errors.New("product not found")
	}
	return stock, err
}

func (r *inventoryRepo) UpdateStock(ctx context.Context, productID string, delta int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE products SET stock = stock + $1 WHERE products_id = $2`, delta, productID)
	return err
}

func (r *inventoryRepo) CreateReservation(ctx context.Context, reservationId, orderID, productID string, qty int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO reservations (reservations_id, product_id, order_id, qty, status)
		VALUES ($1, $2, $3, $4, 'reserved')
	`, reservationId, productID, orderID, qty)
	return err
}

func (r *inventoryRepo) GetReservation(ctx context.Context, reservationID string) (string, int, error) {
	var productID string
	var qty int
	err := r.db.QueryRowContext(ctx, `
		SELECT product_id, qty FROM reservations WHERE reservations_id=$1 AND status='reserved'
	`, reservationID).Scan(&productID, &qty)
	if err == sql.ErrNoRows {
		return "", 0, errors.New("reservation not found")
	}
	return productID, qty, err
}

func (r *inventoryRepo) CancelReservation(ctx context.Context, reservationID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE reservations SET status='canceled' WHERE reservations_id=$1
	`, reservationID)
	return err
}
