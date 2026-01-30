package tracer

import (
	"context"
	"os"
	"time"

	"github.com/kareemhamed001/e-commerce/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func InitTracer(ctx context.Context, serviceName, otlpEndPoint string) (*trace.TracerProvider, error) {

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(otlpEndPoint), otlptracegrpc.WithInsecure())

	res, err := resource.New(
		ctx,
		resource.WithOS(),
		resource.WithTelemetrySDK(),
		resource.WithProcess(),
		resource.WithHost(),
		resource.WithContainer(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.DeploymentEnvironmentName(os.Getenv("APP_ENV")),
			attribute.String("service.version", os.Getenv("APP_VERSION")),
		),
	)

	if err != nil {
		logger.Errorf("Error While Creating Tracing Resource %s", err.Error())
		return nil, ErrCreatingResource
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter, trace.WithBatchTimeout(5*time.Second), trace.WithMaxExportBatchSize(512)),
		trace.WithResource(res),
		trace.WithSampler(getSampler()),
	)

	otel.SetTracerProvider(tp)
	return tp, nil
}

func getSampler() trace.Sampler {
	env := getEnv("APP_ENV", "development")
	if env == "production" {
		logger.Info("Using TraceIDRatioBased sampler with 20% sampling rate for production environment")
		return trace.ParentBased(trace.TraceIDRatioBased(0.2))
	}
	return trace.AlwaysSample()
}
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Shutdown(ctx context.Context, tp *trace.TracerProvider) error {
	if tp == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return tp.Shutdown(ctx)
}
