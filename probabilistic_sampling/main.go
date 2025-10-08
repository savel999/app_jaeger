package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

var urls = []string{
	"https://opentelemetry.io",
	"https://www.google.com/",
	"http://localhost:8080",
	"https://2gis.ru",
	"https://www.wikipedia.org/",
	"https://archive.org/",
	"https://www.bbc.com/news",
	"https://www.khanacademy.org/",
	"https://unsplash.com/",
	"https://www.duolingo.com/",
	"https://zenhabits.net/",
	"https://www.howstuffworks.com/",
	"https://www.ted.com/",
	"https://artsandculture.google.com/",
	"http://localhost:6789",
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	tp, err := initTracer()
	if err != nil {
		logger.Error(err.Error())

		return
	}

	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error("Error shutting down tracer provider: %v", err)
		}
	}()

	wg := sync.WaitGroup{}

	wg.Add(len(urls))

	logger.Info(fmt.Sprintf("необходимо обойти %d урлов", len(urls)))

	for _, url := range urls {
		go func(url string) {
			defer wg.Done()

			client := http.Client{
				Transport: otelhttp.NewTransport(
					http.DefaultTransport,
					otelhttp.WithTracerProvider(tp),
					otelhttp.WithSpanNameFormatter(spanNameFormatter),
				),
				Timeout: 2 * time.Second,
			}

			logger.Debug("Open url:", url)

			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				logger.Error("Request error", err.Error())

				return
			}

			resp, err := client.Do(req)
			if err != nil {
				logger.Error("Response error", err.Error())

				return
			}

			defer resp.Body.Close()

		}(url)

	}

	wg.Wait()
}

func initTracer() (*sdktrace.TracerProvider, error) {
	// Create stdout exporter to be able to retrieve
	// the collected spans.
	exp, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint("localhost:4317"),
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithCompressor("gzip"),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("client"),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, err
}

func spanNameFormatter(_ string, r *http.Request) string {
	return fmt.Sprintf("%s %s", r.Method, r.URL.String())
}
