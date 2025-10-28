# Task — Workflow Orchestrator (Saga) with Compensations + Metrics

## Scenario
Model a **three-step order fulfillment workflow** with compensations:
1. `reserve_pickup_slot`
2. `assign_agent`
3. `notify_customer`

If a step fails, execute compensating actions (e.g., `release_pickup_slot`, `unassign_agent`, `cancel_notification`).

## Requirements
1. **Engine**: Use Asynq or NATS JetStream for steps & retries with exponential backoff.
2. **State Machine**: Persist per-order workflow state in Postgres with transitions and timestamps.
3. **Idempotency**: Steps must be idempotent (retries safe). Store step execution results and dedupe keys.
4. **Recovery**: Provide `cmd/recover` to scan stalled workflows and resume from the correct step.
5. **Observability**: `/metrics` with counts per step, success/failure/compensation, and OTel traces per order.
6. **Chaos**: Optional `--inject-failure` to randomly fail steps to validate compensation.
7. **Docker Compose**: Postgres + Redis (if Asynq) + service(s). Include mock services for slot, agent, notification.

## Deliverables
- Orchestrator service, step handlers, compensation logic.
- DB schema & migrations, seed mocks.
- `README.md` (end-to-end demo), `DESIGN.md` (state model, retry, compensation), `devlog.txt`.
- Tests for at least one happy path and one failure path with compensation.

## Acceptance
- End-to-end run with 100 simulated orders produces successful workflows, with some compensated due to injected failures.
- Crash mid-run then `cmd/recover` continues correctly without double effects.
