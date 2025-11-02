package tracing

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var Tracer = otel.Tracer("workflow-orchestrator")

func InitTracing() func() {
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("workflow-orchestrator"),
			attribute.String("environment", getEnv("ENVIRONMENT", "development")),
		),
	)
	if err != nil {
		panic("failed to create OTel resource: %w" + err.Error())
	}

	var exporter sdktrace.SpanExporter
	exporterType := getEnv("TRACING_EXPORTER", "stdout")
	switch exporterType {
	case "otlp":
		otlpEndpoint := getEnv("OTLP_ENDPOINT", "http://localhost:4318/v1/traces")
		exporter, err = otlptracehttp.New(context.Background(),
			otlptracehttp.WithEndpointURL(otlpEndpoint),
			otlptracehttp.WithInsecure(),
		)
		if err != nil {
			panic("failed to create OTLP HTTP exporter: "+ err.Error())
		}
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			panic("failed to shutdown tracer: "+ err.Error())
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
