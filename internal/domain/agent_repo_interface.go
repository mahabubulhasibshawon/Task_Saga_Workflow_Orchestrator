package domain

import "context"

type AgentRepo interface {
	AssignAgent(ctx context.Context, orderID, agentID string) error
	GetAgentsByOrderID(ctx context.Context, orderID string) ([]string, error)
	UnassignAgents(ctx context.Context, orderID string) error
}
