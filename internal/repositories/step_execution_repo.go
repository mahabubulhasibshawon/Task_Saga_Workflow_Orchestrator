package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/domain"
)

type postgresStepExecutionRepo struct {
	db *sql.DB
}

func NewStepExecutionRepo(db *sql.DB) domain.StepExecutionRepo {
	return &postgresStepExecutionRepo{db: db}
}

func (r *postgresStepExecutionRepo) IsExecuted(ctx context.Context, dedupeKey string) (bool, string, error) {
	var result string
	query := `SELECT result FROM step_executions WHERE dedupe_key = $1`
	err := r.db.QueryRowContext(ctx, query, dedupeKey).Scan(&result)
	if err == sql.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", fmt.Errorf("failed to check execution: %w", err)
	}
	return true, result, nil
}

func (r *postgresStepExecutionRepo) SaveExecution(ctx context.Context, dedupeKey, result string) error {
	query := `INSERT INTO step_executions (dedupe_key, result) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, dedupeKey, result)
	if err != nil {
		return fmt.Errorf("failed to save execution: %w", err)
	}
	return nil
}
