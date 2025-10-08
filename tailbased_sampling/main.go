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

type UrlPing struct {
	URL     string
	Timeout time.Duration
}

var urls = []UrlPing{
	{URL: "https://opentelemetry.io", Timeout: time.Second * 2},
	{URL: "https://www.google.com/", Timeout: time.Second * 2},
	{URL: "http://localhost:8080"},
	{URL: "https://2gis.ru", Timeout: time.Second * 2},
	{URL: "https://www.wikipedia.org/"},
	{URL: "https://archive.org/"},
	{URL: "https://www.bbc.com/news"},
	{URL: "https://www.khanacademy.org/"},
	{URL: "https://unsplash.com/"},
	{URL: "https://www.duolingo.com/"},
	{URL: "https://zenhabits.net/"},
	{URL: "https://www.howstuffworks.com/"},
	{URL: "https://www.ted.com/"},
	{URL: "https://artsandculture.google.com/"},
	{URL: "http://localhost:6789"},
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
		go func(url UrlPing) {
			defer wg.Done()

			client := http.Client{
				Transport: otelhttp.NewTransport(
					NewDelayChain(url.Timeout)(http.DefaultTransport),
					otelhttp.WithTracerProvider(tp),
					otelhttp.WithSpanNameFormatter(spanNameFormatter),
				),
				Timeout: 6 * time.Second,
			}

			now := time.Now()

			req, err := http.NewRequest(http.MethodGet, url.URL, nil)
			if err != nil {
				logger.Error("Request error", err.Error())

				return
			}

			resp, err := client.Do(req)
			if err != nil {
				logger.Error("Response error", err.Error())

				return
			}

			logger.Debug(fmt.Sprintf("Open url:%s , duration: %d ms", url.URL, time.Since(now).Milliseconds()))

			defer resp.Body.Close()

		}(url)

	}

	wg.Wait()
}

type Chain func(http.RoundTripper) http.RoundTripper
type ChainFunc func(*http.Request) (*http.Response, error)

func (f ChainFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func NewDelayChain(delay time.Duration) Chain {
	return func(next http.RoundTripper) http.RoundTripper {
		return ChainFunc(func(r *http.Request) (*http.Response, error) {
			resp, err := next.RoundTrip(r)

			time.Sleep(delay)

			return resp, err
		})
	}
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
