package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	StepSuccess = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "workflow-step_success_total",
			Help: "Total number of successful workflow step executions",
		},
		[]string{"step"},
	)
	StepFailure = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "workflow_step_failue_total",
			Help: "Total numebr of failed workflow step executions",
		},
		[]string{"step"},
	)
	CompensationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "workflow_compensation_total",
			Help: "Total number of compensation actions triggered",
		},
		[]string{"compensation_step"},
	)
)

func InitMetrics() {
	prometheus.MustRegister(StepSuccess, StepFailure, CompensationTotal)
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
