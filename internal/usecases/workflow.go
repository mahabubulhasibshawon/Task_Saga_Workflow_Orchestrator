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

func NextStep(ctx context.Context, orderID string, currentStep domain.Step) error {
    cfg := config.Load()
	db := conn.ConnectPostgres(cfg.DSN())
	defer db.Close()

    workflowRepo := repositories.NewWorkflowRepo(db)
    state, err := workflowRepo.GetStateByOrderID(ctx, orderID)
    if err != nil {
        return fmt.Errorf("failed to get workflow state: %w", err)
    }
    if state == nil {
        return fmt.Errorf("workflow not found for order %s", orderID)
    }

    var nextStep domain.Step
    switch currentStep {
    case domain.StepReserveSlot:
        nextStep = domain.StepAssignAgent
    case domain.StepAssignAgent:
        nextStep = domain.StepNotifyCustomer
    case domain.StepNotifyCustomer:
        return MarkCompleted(ctx, orderID)
    default:
        return fmt.Errorf("unknown current step: %s", currentStep)
    }

    state.CurrentStep = nextStep
    state.UpdatedAt = time.Now()
    if err := workflowRepo.SaveState(ctx, state); err != nil {
        return fmt.Errorf("failed to update workflow state: %w", err)
    }

    client := queue.NewQueueClient()
    defer client.Close()

    payload, err := json.Marshal(queue.StepPayload{OrderID: orderID, Step: nextStep})
    if err != nil {
        return fmt.Errorf("failed to marshal payload: %w", err)
    }
    task := asynq.NewTask("step", payload)
    if _, err := client.EnqueueContext(ctx, task); err != nil {
        return fmt.Errorf("failed to enqueue step %s: %w", nextStep, err)
    }

    logger.Info("Advanced workflow",
        zap.String("order_id", orderID),
        zap.String("next_step", string(nextStep)))
    return nil
}

func MarkCompleted(ctx context.Context, orderID string) error {
	cfg := config.Load()
	db := conn.ConnectPostgres(cfg.DSN())
	defer db.Close()

	orderRepo := repositories.NewOrderRepo(db)
	workflowRepo := repositories.NewWorkflowRepo(db)

	order, err := orderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w",err)
	}
	if order == nil {
		return fmt.Errorf("order %s not found", orderID)
	}
	order.Status = "fulfilled"
    order.UpdatedAt = time.Now()
    if err := orderRepo.SaveOrder(ctx, order); err != nil {
        return fmt.Errorf("failed to update order status: %w", err)
    }

    workflow, err := workflowRepo.GetStateByOrderID(ctx, orderID)
    if err != nil {
        return fmt.Errorf("failed to get workflow state: %w", err)
    }
    if workflow == nil {
        return fmt.Errorf("workflow not found for order %s", orderID)
    }
    workflow.Status = domain.StatusCompleted
    workflow.UpdatedAt = time.Now()
    if err := workflowRepo.SaveState(ctx, workflow); err != nil {
        return fmt.Errorf("failed to update workflow state: %w", err)
    }

    logger.Info("Workflow completed", zap.String("order_id", orderID))
    return nil
}