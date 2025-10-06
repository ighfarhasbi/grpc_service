package db

import (
	"context"
	"database/sql"
	"errors"
)

type OrderRepository interface {
	WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error
	CreateOrder(ctx context.Context, tx *sql.Tx, orderID, userID string, totalAmount float64, status string) error
	CreateOrderItem(ctx context.Context, tx *sql.Tx, orderItemID, orderID, productID string, qty int, price float64) error
	UpdateOrderStatus(ctx context.Context, tx *sql.Tx, orderID, status string) error
	GetOrder(ctx context.Context, orderID string) (Order, []OrderItem, error)
}

type Order struct {
	OrderID     string
	UserID      string
	Status      string
	TotalAmount float64
}

type OrderItem struct {
	OrderItemID string
	ProductID   string
	Qty         int
	UnitPrice   float64
	TotalPrice  float64
}

type orderRepo struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepo{db: db}
}

func (r *orderRepo) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *orderRepo) CreateOrder(ctx context.Context, tx *sql.Tx, orderID, userID string, totalAmount float64, status string) error {
	query := `
		INSERT INTO orders (orders_id, users_id, total_amount, status)
		VALUES ($1, $2, $3, $4)
	`
	_, err := tx.ExecContext(ctx, query, orderID, userID, totalAmount, status)
	return err
}

func (r *orderRepo) CreateOrderItem(ctx context.Context, tx *sql.Tx, orderItemID, orderID, productID string, qty int, unitPrice float64) error {
	query := `
		INSERT INTO order_items (order_items_id, orders_id, products_id, qty, unit_price)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := tx.ExecContext(ctx, query, orderItemID, orderID, productID, qty, unitPrice)
	return err
}

func (r *orderRepo) UpdateOrderStatus(ctx context.Context, tx *sql.Tx, orderID, status string) error {
	query := `
		UPDATE orders 
		SET status = $1
		WHERE orders_id = $2 AND status != 'confirmed'
	`
	_, err := tx.ExecContext(ctx, query, status, orderID)
	return err
}

func (r *orderRepo) GetOrder(ctx context.Context, orderID string) (Order, []OrderItem, error) {
	var o Order
	err := r.db.QueryRowContext(ctx, `
		SELECT orders_id, users_id, total_amount, status 
		FROM orders WHERE orders_id=$1
	`, orderID).Scan(&o.OrderID, &o.UserID, &o.TotalAmount, &o.Status)
	if err == sql.ErrNoRows {
		return o, nil, errors.New("order not found")
	} else if err != nil {
		return o, nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT order_items_id, products_id, qty, unit_price, total_price 
		FROM order_items WHERE orders_id=$1
	`, orderID)
	if err != nil {
		return o, nil, err
	}
	defer rows.Close()

	var items []OrderItem
	for rows.Next() {
		var it OrderItem
		if err := rows.Scan(&it.OrderItemID, &it.ProductID, &it.Qty, &it.UnitPrice, &it.TotalPrice); err != nil {
			return o, nil, err
		}
		items = append(items, it)
	}

	return o, items, nil
}
