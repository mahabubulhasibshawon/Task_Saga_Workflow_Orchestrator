package domain

import "context"

type StepExecutionRepo interface {
	IsExecuted(ctx context.Context, dedupeKey string) (bool, string, error)
	SaveExecution(ctx context.Context, dedupeKey, result string) error
}
