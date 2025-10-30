package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/domain"
)

type postgresAgentRepo struct {
	db *sql.DB
}

func NewAgentRepo(db *sql.DB) domain.AgentRepo {
	return &postgresAgentRepo{db: db}
}

func (r *postgresAgentRepo) AssignAgent(ctx context.Context, orderID, agentID string) error {
	query := `
        INSERT INTO agents (order_id, agent_id, assigned_at)
        VALUES ($1, $2, $3)
        ON CONFLICT (order_id, agent_id) DO NOTHING
    `
	_, err := r.db.ExecContext(ctx, query, orderID, agentID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to assign agent %s to order %s: %w", agentID, orderID, err)
	}
	return nil
}

func (r *postgresAgentRepo) GetAgentsByOrderID(ctx context.Context, orderID string) ([]string, error) {
	query := `SELECT agent_id FROM agents WHERE order_id = $1`
	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agents for order %s: %w", orderID, err)
	}
	defer rows.Close()

	var agentIDs []string
	for rows.Next() {
		var agentID string
		if err := rows.Scan(&agentID); err != nil {
			return nil, fmt.Errorf("failed to scan agent ID: %w", err)
		}
		agentIDs = append(agentIDs, agentID)
	}
	return agentIDs, rows.Err()
}

func (r *postgresAgentRepo) UnassignAgents(ctx context.Context, orderID string) error {
	query := `DELETE FROM agents WHERE order_id = $1`
	_, err := r.db.ExecContext(ctx, query, orderID)
	if err != nil {
		return fmt.Errorf("failed to unassign agents for order %s: %w", orderID, err)
	}
	return nil
}
