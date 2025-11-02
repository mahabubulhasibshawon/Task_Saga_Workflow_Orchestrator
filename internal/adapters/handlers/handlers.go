package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync/atomic"

	"github.com/hibiken/asynq"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/config"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/adapters/queue"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/domain"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/repositories"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/usecases"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/pkg/conn"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/pkg/mocks"
	"go.uber.org/zap"
)

var (
	logger      *zap.Logger
	failureProb atomic.Value
)

func init() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to init logger: %v", err))
	}
	failureProb.Store(0.0)
}

func SetFailureProbability(prob float64) {
	failureProb.Store(prob)
}

func HandleTask(ctx context.Context, t *asynq.Task) error {
	var payload queue.StepPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal step payload: %w", err)
	}

	logger.Info("Processing step",
		zap.String("order_id", payload.OrderID),
		zap.String("step", string(payload.Step)),
	)

	cfg := config.Load()
	db := conn.ConnectPostgres(cfg.DSN())
	defer db.Close()

	stepRepo := repositories.NewStepExecutionRepo(db)
	dedupeKey := fmt.Sprintf("%s_%s", payload.OrderID, payload.Step)
	executed, result, err := stepRepo.IsExecuted(ctx, dedupeKey)
	if err != nil {
		return fmt.Errorf("failed to check step execution: %w", err)
	}
	if executed {
		logger.Info("Step already executed",
			zap.String("order_id", payload.OrderID),
			zap.String("step", string(payload.Step)),
			zap.String("result", result))
		if result == "success" {
			return usecases.NextStep(ctx, payload.OrderID, payload.Step)
		}
		return fmt.Errorf("step previously failed: %s", result)
	}

	var chaosErr error
	if rand.Float64() < failureProb.Load().(float64) {
		chaosErr = fmt.Errorf("injected failure for step %s", payload.Step)
	}

	var stepErr error
	switch payload.Step {
	case domain.StepReserveSlot:
		_, stepErr = mocks.ReserveSlot(payload.OrderID)
	case domain.StepAssignAgent:
		_, stepErr = mocks.AssignAgent(ctx, payload.OrderID)
	case domain.StepNotifyCustomer:
		stepErr = mocks.NotifyCustomer(payload.OrderID)
	default:
		stepErr = fmt.Errorf("unknown step: %s", stepErr)
	}

	result = "success"
	if stepErr != nil || chaosErr != nil {
		result = "failed"
		if err := stepRepo.SaveExecution(ctx, dedupeKey, result); err != nil {
			return fmt.Errorf("failed to save step execution: %w", err)
		}
		if err := usecases.Compensate(ctx, payload.OrderID, payload.Step); err != nil {
			return fmt.Errorf("failed to compensate: %w", err)
		}
		return fmt.Errorf("step failed: %w", coalesceErr(stepErr, chaosErr))
	}
	if err := stepRepo.SaveExecution(ctx, dedupeKey, result); err != nil {
		return fmt.Errorf("failed to save step execution: %w", err)
	}
	if err := usecases.NextStep(ctx, payload.OrderID, payload.Step); err != nil {
		return fmt.Errorf("failed to enqueue next step: %w", err)
	}
	return nil
}

func coalesceErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}
