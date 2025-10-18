package telemetry

import (
	"context"
	"log"
	"os"
	"strconv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// Init sets up telemetry. If OTEL_EXPORTER=otlp is set, it attempts to create an OTLP gRPC exporter
func InitForConsole() (trace.Tracer, *sdktrace.TracerProvider) {
	var exp sdktrace.SpanExporter
	var err error
	if os.Getenv("OTEL_EXPORTER") == "otlp" {
		// OTLP via gRPC to localhost:4317
		client := otlptracegrpc.NewClient()
		exp, err = otlptracegrpc.New(context.Background(), otlptracegrpc.WithInsecure())
		if err != nil {
			log.Printf("otel otlp init failed: %v, falling back to console", err)
			exp, _ = stdouttrace.New(stdouttrace.WithPrettyPrint())
		}
	} else {
		exp, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil { log.Fatalf("otel export: %v", err) }
	}
	res, _ := resource.New(context.Background(), resource.WithAttributes(semconv.ServiceNameKey.String("aegis-gateway")))
	// configure sampler from env var SAMPLED_RATE (0.0 - 1.0)
	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(1.0))
	if v := os.Getenv("SAMPLED_RATE"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(f))
		}
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp), sdktrace.WithResource(res), sdktrace.WithSampler(sampler))
	otel.SetTracerProvider(tp)
	tr := tp.Tracer("aegis")
	return tr, tp
}

// TraceIDFromContext extracts the trace id string from context if available
func TraceIDFromContext(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()
	if sc.IsValid() {
		return sc.TraceID().String()
	}
	return ""
}
