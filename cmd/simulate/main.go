package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git/internal/usecases"
)

func main() {
	num := flag.Int("num", 10, "Number of orders to simulate")
	delay := flag.Duration("delay", 500*time.Millisecond, "Delay between order creations")
	flag.Parse()

	log.Printf("Simulating %d orders...\n", *num)

	for i := 0; i < *num; i++ {
		orderID := uuid.New().String()
		if err := usecases.StartWorkflow(context.Background(), orderID); err != nil {
			log.Printf("Failed to start workflow for %s: %v", orderID, err)
		} else {
			log.Printf("Started workflow for order %s", orderID)
		}
		time.Sleep(*delay)
	}

	log.Println("Simulation complete.")
}
