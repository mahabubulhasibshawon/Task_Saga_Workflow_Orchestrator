package domain

import "time"

type Order struct {
	ID        string
	Status    string // pending, fulfilled, failed
	CreatedAt time.Time
	UpdatedAt time.Time
}
