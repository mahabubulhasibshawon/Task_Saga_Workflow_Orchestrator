# Workflow Orchestrator (Saga Pattern)

A **resilient, observable, and fault-tolerant** order processing system built with the **Saga pattern** in Go. Supports **multiple agents per order**, **idempotent step execution**, **automatic compensation on failure**, **distributed tracing**, **metrics**, and **chaos testing**.

---

## Features

| Feature | Description |
|-------|-----------|
| **Saga Orchestration** | Forward steps: `reserve_pickup_slot → assign_agent → notify_customer` |
| **Multiple Agents** | Assign 2+ agents per order (persisted in DB) |
| **Compensation Logic** | Rollback on failure: `unassign_agent`, `release_slot`, `cancel_notification` |
| **Idempotency** | Safe retries using `step_executions` table |
| **Recovery** | Resume stalled workflows after crash |
| **Distributed Tracing** | OpenTelemetry |
| **Metrics** | Prometheus: step success/failure, compensations |
| **Chaos Testing** | Inject failures via `--inject-failure` flag |
| **CLI Tools** | `simulate`, `recover`, `orchestrator` |
| **Clean Architecture** | Domain → Repos → Usecases → Adapters |
| **Go 1.23+ Ready** | Uses latest Go features and patterns |

---

## Workflow
![alt text](saga-workflow.png)

## Project Structure

```bash
workflow-orchestrator/
├── cmd/
│   ├── orchestrator/   # Asynq server + metrics + tracing
│   ├── simulate/       # Generate N orders
│   └── recover/        # Resume stalled workflows
├── internal/
│   ├── domain/         # Order, WorkflowState, Steps
│   ├── repositories/   # DB access (orders, workflows, agents, steps)
│   ├── usecases/       # Business logic (start, next, compensate)
│   └── adapters/
│       ├── handlers/   # Asynq task handlers
│       ├── queue/      # Asynq client/server
│       ├── metrics/    # Prometheus
│       └── tracing/    # OTel 
├── pkg/mocks/          # In-memory services + DB-backed agents
├── migrations/         # Golang-migrate SQL migrations
├── docker-compose.yml  # Postgres, Redis
└── README.md
```

---

## Prerequisites

- [Go 1.23+](https://go.dev/dl/)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- [Golang-migrate](github.com/golang-migrate/migrate/v4) (for migrations)

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

---

## Quick Start

### 1. Clone & Setup

```bash
git clone https://github.com/mahabubulhasibshawon/Task_Saga_Workflow_Orchestrator.git
cd workflow-orchestrator
cp .env.example .env
```

### 2. Start Infrastructure

```bash
docker compose up -d
```

This starts:
- **Postgres**: `localhost:5433`
- **Redis**: `localhost:6380`

### 3. Apply Migrations

```bash
migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5433/workflow_db?sslmode=disable" -verbose up
```

### 4. Run Orchestrator

```bash
go run cmd/orchestrator/main.go --inject-failure=0.2
```

> Injects 20% chance of step failure → triggers compensation.

### 5. Simulate Orders

```bash
go run cmd/simulate/main.go --num=10 --delay=500ms
```

Creates 10 orders, each with:
- 1 pickup slot
- **2 assigned agents**
- 1 customer notification

### 6. Observe

| Tool | URL |
|------|-----|
| **Prometheus Metrics** | `http://localhost:2112/metrics` |
| **Database** | `psql $DB_URL` |

```sql
-- See agents per order
SELECT order_id, COUNT(*) as agents 
FROM agents 
GROUP BY order_id 
ORDER BY agents DESC;
```

---

## CLI Commands

### Orchestrator
```bash
go run cmd/orchestrator/main.go --inject-failure=0.3
```

### Simulate Load
```bash
go run cmd/simulate/main.go --num=50 --delay=200ms
```

### Recover Stalled Workflows
```bash
go run cmd/recover/main.go --timeout=2m
```

---

## Database Schema

```sql
orders          → order status
workflows       → current step & status
step_executions → idempotency key → result
agents          → order_id → agent_id (multiple rows)
```

## DB Diagram
![alt text](image.png)

---

## Observability

### Metrics (Prometheus)

```promql
# Step success rate
rate(workflow_step_success_total[5m])

# Failure rate
rate(workflow_step_failure_total[5m])

# Compensations triggered
workflow_compensation_total
```
---

Covers:
- Agent assignment (multiple, idempotent)
- Mock services
- Repository logic

---

## Development

### Add New Step

1. Add to `domain/workflow_state.go`
2. Implement forward + compensation in `pkg/mocks/services.go`
3. Update `usecases/workflow.go` and `compensation.go`
4. Add metrics labels

---

## Production Tips

- Replace `pkg/mocks` with real microservices
- Add retry policies in Asynq
- Use connection pooling
- Add health checks
- Deploy with Kubernetes + Helm

---

## Contributing

1. Fork it
2. Create your feature branch
3. Write tests
4. Commit with clear messages
5. Open a Pull Request

---



---

**Built with Go, love, and resilience.**

Let it fail. Let it recover. Let it trace. Let it **scale**.

--- 

*Happy orchestrating!*