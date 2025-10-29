package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/domain"
)

type postgresOrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) domain.OrderRepo {
	return &postgresOrderRepo{db: db}
}

func (r *postgresOrderRepo) SaveOrder(ctx context.Context, order *domain.Order) error {
	query := `
		INSERT INTO orders (id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(ctx, query, order.ID, order.Status, order.CreatedAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}
	return nil
}

func (r *postgresOrderRepo) GetOrderByID(ctx context.Context, orderID string) (*domain.Order, error) {
	query := `SELECT id, status, created_at, updated_at FROM orders WHERE id = $1`
	order := &domain.Order{}
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(&order.ID, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil // Not found, return nil order
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order by ID %s: %w", orderID, err)
	}
	return order, nil
}
