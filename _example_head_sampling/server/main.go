package main

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
	"io"
	"log"
	"net/http"
	"time"

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
			semconv.ServiceNameKey.String("server1"),
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

		//span := trace.SpanFromContext(ctx)
		//span.AddEvent("handling this...")

		ctx := req.Context()
		url := "http://localhost:8888/hello"
		client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

		err = requestToServer2(ctx, tp.Tracer("server1"), url, client)
		if err != nil {
			log.Fatal(err)
		}

		_, _ = io.WriteString(w, "Hello, world!\n")
	}

	otelHandler := otelhttp.NewHandler(http.HandlerFunc(helloHandler), "Hello7777")

	http.Handle("/hello", otelHandler)

	log.Printf("Start server...")
	err = http.ListenAndServe(":7777", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func requestToServer2(ctx context.Context, tr trace.Tracer, url string, client http.Client) error {
	var body []byte
	ctx, span := tr.Start(ctx, "say hello", trace.WithAttributes(semconv.PeerServiceKey.String("ExampleService")))
	defer span.End()
	span.SetAttributes(attribute.Key("method").String("requestToServer"))
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	fmt.Printf("Sending request...\n")
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	body, err = io.ReadAll(res.Body)
	_ = res.Body.Close()

	fmt.Printf("Response Received: %s\n\n\n", body)
	fmt.Printf("Waiting for few seconds to export spans ...\n\n")
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("Inspect traces on stdout\n")

	return err
}
