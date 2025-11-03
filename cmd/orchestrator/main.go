package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/adapters/handlers"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/adapters/metrics"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/adapters/queue"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/adapters/tracing"
)

func main() {
	injectFailure := flag.Float64("inject-failure", 0.0, "Probability of injected failure (0.0 to 1.0)")
	flag.Parse()

	// initialize tracing
	cleanup := tracing.InitTracing()
	defer cleanup()

	// initialize metrics
	metrics.InitMetrics()
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", metrics.MetricsHandler())
		if err := http.ListenAndServe(":2112", mux); err != nil {
			log.Fatalf("Failed to start metrics server: %w", err)
		}
	}()

	// set chaos
	handlers.SetFailureProbability(*injectFailure)

	// start asynq server
	server := queue.NewQueueServer()
	mux := queue.NewServeMux()
	mux.HandleFunc("step", handlers.HandleStep)
	mux.HandleFunc("compensation", handlers.HandleCompensation)

	go func() {
		if err := server.Run(mux); err != nil {
			log.Fatalf("Asynq sever error ", err)
		}
	}()

	log.Println("Orchestrator running. Press Ctrl+C to stop.")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown()
	log.Println("Orchestrator stopped")
}
