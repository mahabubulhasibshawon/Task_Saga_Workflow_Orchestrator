package domain

import "context"

type OrderRepo interface {
	SaveOrder(ctx context.Context, order *Order) error
	GetOrderByID(ctx context.Context, orderID string) (*Order, error)
}