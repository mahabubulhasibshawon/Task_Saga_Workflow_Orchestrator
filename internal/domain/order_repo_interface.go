package domain

import "context"

type Orderrepo interface {
	SaveOrder(ctx context.Context, order *Order) error
	GetOrderByID(ctx context.Context, orderID string) (*Order, error)
}