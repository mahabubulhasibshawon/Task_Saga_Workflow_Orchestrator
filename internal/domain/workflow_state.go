package domain

import "time"

type WorkflowStatus string

const (
	StatusPending     WorkflowStatus = "pending"
	StatusCompleted   WorkflowStatus = "completed"
	StatusFailed      WorkflowStatus = "failed"
	StatusCompensated WorkflowStatus = "compensated"
)

type Step string

const (
	StepReserveSlot    Step = "reserve_pickup_slot"
	StepAssignAgent    Step = "assign_agent"
	StepNotifyCustomer Step = "notify_customer"
)

type CompensationStep string

const (
	CompReleaseSlot        CompensationStep = "release_pickup_slot"
	CompUnassignAgent      CompensationStep = "unassign_agent"
	CompCancelNotification CompensationStep = "cancel_notification"
)

type WorkflowState struct {
	OrderID     string
	CurrentStep Step
	Status      WorkflowStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
