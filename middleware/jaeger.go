package middleware

import (
	"os"

	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func JaegerTraceProvider() (*sdktrace.TracerProvider, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
	// exp, err := jaeger.New(jaeger.WithAgentEndpoint(jaeger.WithAgentHost(os.Getenv("JAEGER_AGENT_HOST")), jaeger.WithAgentPort(os.Getenv("JAEGER_AGENT_PORT"))))
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(os.Getenv("JAEGER_SERVICE_NAME")),
			semconv.DeploymentEnvironmentKey.String(os.Getenv("ENV")),
		)),
	)
	return tp, nil
}