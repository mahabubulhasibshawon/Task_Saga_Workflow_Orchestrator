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

func Compensate(ctx context.Context, orderID string, failedStep domain.Step) error {
	client := queue.NewQueueClient()
	defer client.Close()

	cfg := config.Load()
	db := conn.ConnectPostgres(cfg.DSN())
	defer db.Close()

	orderRepo := repositories.NewOrderRepo(db)
	workflowRepo := repositories.NewWorkflowRepo(db)

	order, err := orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return fmt.Errorf("order %s not found", orderID)
	}

	order.Status = "failed"
	if err := orderRepo.SaveOrder(ctx, order); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	workflow, err := workflowRepo.GetStateByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get workflow state: %w", err)
	}
	if workflow != nil {
		workflow.Status = domain.StatusCompensated
		workflow.UpdatedAt = time.Now()
		if err := workflowRepo.SaveState(ctx, workflow); err != nil {
			return fmt.Errorf("failed to update workflow state: %w", err)
		}
	}

	var compSteps []domain.CompensationStep
	switch failedStep {
	case domain.StepAssignAgent:
		compSteps = []domain.CompensationStep{domain.CompReleaseSlot}
	case domain.StepNotifyCustomer:
		compSteps = []domain.CompensationStep{domain.CompUnassignAgent, domain.CompReleaseSlot}
	case domain.StepReserveSlot:
		return nil
	}

	for _, comp := range compSteps {
		payload, err := json.Marshal(queue.StepPayload{OrderID: orderID, Step: domain.Step(comp)})
		if err != nil {
			return fmt.Errorf("failed to marshal compensation payload: %w", err)
		}
		task := asynq.NewTask("compensation", payload)
		if _, err := client.EnqueueContext(ctx, task); err != nil {
			return fmt.Errorf("failed to enqueue compensation %s: %w", comp, err)
		}
		logger.Info("Enqueued compensation",
			zap.String("order_id", orderID),
			zap.String("compensation", string(comp)),
		)
	}

	return nil
}
