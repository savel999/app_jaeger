package main

import (
	"context"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"io"
	"log"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	// Create stdout exporter to be able to retrieve
	// the collected spans.
	//exp, err := otlptrace.New(
	//	context.Background(),
	//	otlptracehttp.NewClient(
	//		otlptracehttp.WithEndpoint("localhost:4318"),
	//		otlptracehttp.WithInsecure(),
	//	),
	//)

	exp, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint("localhost:4317"),
			otlptracegrpc.WithInsecure(),
		),
	)
	if err != nil {
		return nil, err
	}

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("server2"),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, err
}

func main() {
	tp, err := initTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		//ctx := req.Context()
		//span := trace.SpanFromContext(ctx)
		//span.AddEvent("handling this...")

		_, _ = io.WriteString(w, "Hello, world!\n")
	}

	otelHandler := otelhttp.NewHandler(http.HandlerFunc(helloHandler), "Hello77777")

	http.Handle("/hello", otelHandler)

	log.Printf("Start server...")
	err = http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Fatal(err)
	}
}
