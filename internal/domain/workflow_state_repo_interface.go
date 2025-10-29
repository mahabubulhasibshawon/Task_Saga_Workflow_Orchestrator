package domain

import (
	"context"
	"time"
)

type Workflowrepo interface {
	SaveState(ctx context.Context, state *WorkflowState) error
	GetStateByOrderID(ctx context.Context, orderID string) (*WorkflowState, error)
	GetStalledWorkflows(ctx context.Context, timeout time.Duration) ([]*WorkflowState, error)
}