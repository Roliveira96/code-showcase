package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

func initTracer() (func(context.Context), error) {
	exporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("localhost:4317"),
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar o exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("servico-gin-golang"),
			semconv.ServiceVersion("v0.1.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar o resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return func(ctx context.Context) {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("erro ao fazer shutdown do tracer provider: %v", err)
		}
	}, nil
}

func main() {
	shutdown, err := initTracer()
	if err != nil {
		log.Fatalf("falha ao inicializar o tracer: %v", err)
	}
	defer shutdown(context.Background())

	r := gin.Default()
	r.Use(otelgin.Middleware("meu-servico-gin"))

	r.GET("/ping", func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		log.Printf(
			"Trace-ID: %s | Span-ID: %s - Dados de telemetria gerados para a rota /ping",
			span.SpanContext().TraceID().String(),
			span.SpanContext().SpanID().String(),
		)

		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	r.GET("/hello/:name", func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		log.Printf(
			"Trace-ID: %s | Span-ID: %s - Dados de telemetria gerados para a rota /hello/:name",
			span.SpanContext().TraceID().String(),
			span.SpanContext().SpanID().String(),
		)

		name := c.Param("name")
		tracer := otel.Tracer("minha-rota-customizada")
		_, customSpan := tracer.Start(c.Request.Context(), "processamento-interno")
		customSpan.SetAttributes(attribute.String("param.name", name))
		time.Sleep(100 * time.Millisecond)
		customSpan.End()
		message := fmt.Sprintf("Hello, %s!", name)
		c.JSON(http.StatusOK, gin.H{"greeting": message})
	})

	log.Println("Servidor Gin rodando na porta :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("falha ao iniciar o servidor Gin: %v", err)
	}
}
