package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/config"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/adapters/queue"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/repositories"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/pkg/conn"
)

func main() {
	timeout := flag.Duration("timeout", 5*time.Minute, "Consider workflows stalled if not updated for this duration")
	flag.Parse()

	cfg := config.Load()
	db := conn.ConnectPostgres(cfg.DSN())
	defer db.Close()

	workflowRepo := repositories.NewWorkflowRepo(db)
	stalled, err := workflowRepo.GetStalledWorkflows(context.Background(), *timeout)
	if err != nil {
		log.Fatalf("Failed to get stalled workflows: %v", err)
	}

	log.Printf("Found %d stalled workflows\n", len(stalled))

	client := queue.NewQueueClient()
	defer client.Close()

	for _, state := range stalled {
		payload := queue.StepPayload{OrderID: state.OrderID, Step: state.CurrentStep}
		if err := queue.EnqueueStep(context.Background(), client, "step", payload); err != nil {
			log.Printf("Failed to reenqueue %s: %v", state.OrderID, err)
		} else {
			log.Printf("Reenqueued stalled workflow: %s (step: %s)", state.OrderID, state.CurrentStep)
		}
	}
}
