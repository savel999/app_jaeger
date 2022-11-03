package main

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"io"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	// Create stdout exporter to be able to retrieve
	// the collected spans.
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

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		//sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("client"),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, err
}

func getNotSampledTracer() (*sdktrace.TracerProvider, error) {
	// Create stdout exporter to be able to retrieve
	// the collected spans.
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
		sdktrace.WithSampler(sdktrace.NeverSample()),
		//sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.NeverSample())),
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("client_never_sampled"),
		)),
	)

	return tp, err
}

func main() {
	tp, err := initTracer()
	if err != nil {
		log.Fatal(err)
	}
	neverTp, err := getNotSampledTracer()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}

		if err := neverTp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	url := "http://localhost:7777/hello"

	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	ctx := context.Background()

	//tr := otel.Tracer("example/client")

	for i := 1; i <= 10; i++ {
		if i%2 == 0 {
			err = requestToServerNever(ctx, neverTp.Tracer("client_never_sampled"), url, client)
		} else {
			err = requestToServer(ctx, tp.Tracer("client"), url, client)
		}

		if err != nil {
			log.Fatal(err)
		}
	}

	time.Sleep(1 * time.Second)
	fmt.Printf("END\n")
}

func requestToServer(ctx context.Context, tr trace.Tracer, url string, client http.Client) error {
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

func requestToServerNever(ctx context.Context, tr trace.Tracer, url string, client http.Client) error {
	var body []byte
	ctx, span := tr.Start(ctx, "say hello", trace.WithAttributes(semconv.PeerServiceKey.String("ExampleService")))
	defer span.End()
	span.SetAttributes(attribute.Key("method").String("requestToServerNever"))
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
