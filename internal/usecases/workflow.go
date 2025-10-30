package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/config"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/adapters/queue"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/domain"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/repositories"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/pkg/conn"
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to init logger: %v", err))
	}
}

func StartWorkflow(ctx context.Context, orderID string) error {
	cfg := config.Load()
	db := conn.ConnectPostgres(cfg.DSN())
	defer db.Close()

	orderRepo := repositories.NewOrderRepo(db)
	workflowRepo := repositories.NewWorkflowRepo(db)

	order := &domain.Order{
		ID:        orderID,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := orderRepo.SaveOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	state := &domain.WorkflowState{
		OrderID:     orderID,
		CurrentStep: domain.StepReserveSlot,
		Status:      domain.StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := workflowRepo.SaveState(ctx, state); err != nil {
		return fmt.Errorf("failed to save workflow state: %w", err)
	}

	client := queue.NewQueueClient()
	defer client.Close()

	payload, err := json.Marshal(queue.StepPayload{OrderID: orderID, Step: domain.StepReserveSlot})
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	task := asynq.NewTask("step", payload)
	if _, err := client.EnqueueContext(ctx, task); err != nil {
		return fmt.Errorf("failed to enqueue first step: %w", err)
	}

	logger.Info("Started workflow",
		zap.String("order_id", orderID),
		zap.String("step", string(domain.StepReserveSlot)))
	return nil
}


