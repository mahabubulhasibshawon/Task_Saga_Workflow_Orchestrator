package mocks

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/config"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/repositories"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/pkg/conn"
)

var (
	slots = sync.Map{}
	notifications = sync.Map{}
)

func ReserveSlot(orderID string) (string, error) {
	if _, exists := slots.Load(orderID); exists {
		return "", fmt.Errorf("slots already reserver for order %s", orderID)
	}
	slotID := "slot-"+orderID
	slots.Store(orderID, slotID)
	return slotID, nil
}

func ReleaseSlot(orderID string) (error) {
	if _, exists := slots.Load(orderID); !exists {
		return fmt.Errorf("no slot exist for order %s", orderID)
	}
	slots.Delete(orderID)
	return nil
}

func AssignAgent(ctx context.Context, orderID string) ([]string, error) {
	cfg := config.Load()
	db := conn.ConnectPostgres(cfg.DSN())
	defer db.Close()

	repo := repositories.NewAgentRepo(db)
	agentIDs := []string{uuid.NewString(), uuid.NewString()} // Assign two agents
    for _, agentID := range agentIDs {
        if err := repo.AssignAgent(ctx, orderID, agentID); err != nil {
            return nil, fmt.Errorf("failed to assign agent %s: %w", agentID, err)
        }
    }
	return agentIDs, nil
}

func UnassignAgent(ctx context.Context, orderID string) error {
    cfg := config.Load()
	db := conn.ConnectPostgres(cfg.DSN())
	defer db.Close()

    repo := repositories.NewAgentRepo(db)
    return repo.UnassignAgents(ctx, orderID)
}

func NotifyCustomer(orderID string) error {
    if _, exists := notifications.Load(orderID); exists {
        return fmt.Errorf("notification already sent for order %s", orderID)
    }
    message := "Order " + orderID + " is ready for pickup"
    notifications.Store(orderID, message)
    return nil
}

func CancelNotification(orderID string) error {
    if _, exists := notifications.Load(orderID); !exists {
        return fmt.Errorf("no notification found for order %s", orderID)
    }
    notifications.Delete(orderID)
    return nil
}