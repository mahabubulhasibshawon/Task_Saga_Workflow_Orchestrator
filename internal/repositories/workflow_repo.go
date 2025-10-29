package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/domain"
)

type postgresWorkflowRepo struct {
	db *sql.DB
}

func NewWorkflowRepo(db *sql.DB) domain.WorkflowRepo {
	return &postgresWorkflowRepo{db: db}
}

func (r *postgresWorkflowRepo) SaveState(ctx context.Context, state *domain.WorkflowState) error {
	query := `
		INSERT INTO workflows (order_id, current_step, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (order_id) DO UPDATE SET
			current_step = EXCLUDED.current_step,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(ctx, query, state.OrderID, state.CurrentStep, state.Status, state.CreatedAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save workflow state: %w", err)
	}
	return nil
}

func (r *postgresWorkflowRepo) GetStateByOrderID(ctx context.Context, orderID string) (*domain.WorkflowState, error) {
	state := &domain.WorkflowState{}
	query := `SELECT order_id, current_step, status, created_at, updated_at FROM workflows WHERE order_id = $1`
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(&state.OrderID, &state.CurrentStep, &state.Status, &state.CreatedAt, &state.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow state: %w", err)
	}
	return state, nil
}

func (r *postgresWorkflowRepo) GetStalledWorkflows(ctx context.Context, timeout time.Duration) ([]*domain.WorkflowState, error) {
	query := `SELECT order_id, current_step, status, created_at, updated_at FROM workflows WHERE status = 'pending' AND updated_at < $1`
	rows, err := r.db.QueryContext(ctx, query, time.Now().Add(-timeout))
	if err != nil {
		return nil, fmt.Errorf("failed to query stalled workflows: %w", err)
	}
	defer rows.Close()

	var states []*domain.WorkflowState
	for rows.Next() {
		state := &domain.WorkflowState{}
		if err := rows.Scan(&state.OrderID, &state.CurrentStep, &state.Status, &state.CreatedAt, &state.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan stalled workflow: %w", err)
		}
		states = append(states, state)
	}
	return states, rows.Err()
}
