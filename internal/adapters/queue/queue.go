package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/config"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/domain"
)

type StepPayload struct {
	OrderID string      `json:"order_id"`
	Step    domain.Step `json:"step"`
}

func NewQueueClient() *asynq.Client {
	addr := config.Load().RedisAddr
	return asynq.NewClient(asynq.RedisClientOpt{Addr: addr})
}

func NewQueueServer() *asynq.Server {
	addr := config.Load().RedisAddr
	return asynq.NewServer(asynq.RedisClientOpt{Addr: addr}, asynq.Config{
		Concurrency: 10,
		RetryDelayFunc: func(n int, err error, _ *asynq.Task) time.Duration {
			return time.Duration(1<<n) * time.Second
		},
	})
}

func NewServeMux() *asynq.ServeMux {
	return asynq.NewServeMux()
}

func EnqueueStep(ctx context.Context, client *asynq.Client, taskType string, payload StepPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marsahl paylaod: %w", err)
	}
	task := asynq.NewTask(taskType, data)
	_, err = client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}
	return nil
}