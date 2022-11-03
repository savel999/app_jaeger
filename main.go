// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Command jaeger is an example program that creates spans
// and uploads to Jaeger.
package main

import (
	"context"
	"flag"
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"log"
	"net/http"
	"time"
)

const (
	service     = "trace-client"
	environment = "debug"
	id          = 1
)

// tracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func tracerProvider(url string) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	//exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))

	//grpc
	exp, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint("localhost:4317"),
			otlptracegrpc.WithInsecure(),
		),
	)

	//http
	//exp, err := otlptrace.New(
	//	context.Background(),
	//	otlptracehttp.NewClient(
	//		otlptracehttp.WithEndpoint("localhost:4318"),
	//		otlptracehttp.WithInsecure(),
	//	),
	//)

	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		//tracesdk.WithSampler(tracesdk.AlwaysSample()),
		// Record information about this application in a Resource.
		//tracesdk.WithResource(resource.NewWithAttributes(
		//	semconv.SchemaURL,
		//	semconv.ServiceNameKey.String(service),
		//	attribute.String("environment", environment),
		//)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}

func main() {
	//mClient := createMultimediaClient()
	tp, err := tracerProvider("localhost:4318")
	if err != nil {
		log.Fatal(err)
	}

	// Cleanly shutdown and flush telemetry when the application exits.
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	ctx := context.Background()

	for i := 0; i < 100; i++ {
		goodTrace(ctx)
		badTrace(ctx)
	}

	//bar(ctx)
}

func goodTrace(ctx context.Context) {
	tr := otel.Tracer("component-main")
	ctx, span := tr.Start(ctx, "foo-good")
	defer span.End()

	span.SetStatus(codes.Ok, "Что-то пошло не так!!")
	time.Sleep(10 * time.Millisecond)
}

func badTrace(ctx context.Context) {
	tr := otel.Tracer("component-main")
	ctx, span := tr.Start(ctx, "foo-bad")
	defer span.End()

	span.SetStatus(codes.Error, "Что-то пошло не так!!")
	time.Sleep(10 * time.Millisecond)
}

func bar2(ctx context.Context) {
	// Use the global TracerProvider.
	tr := otel.Tracer("component-bar2")
	ctx, span := tr.Start(ctx, "bar2")
	defer span.End()

	span.SetAttributes(attribute.Key("testset").String("value"))

	client := http.Client{
		Transport: http.DefaultTransport,
	}

	fmt.Printf("SpanID: %+v\n", span.SpanContext().SpanID())
	fmt.Printf("TraceID: %+v\n", span.SpanContext().TraceID())
	fmt.Printf("SpanContext: %+v\n", span.SpanContext())

	req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:8181/hello", nil)
	req.Header.Set("traceparent", fmt.Sprintf("00-%s-%s-01", span.SpanContext().TraceID(), span.SpanContext().SpanID()))
	_, err := client.Do(req)
	if err != nil {
		fmt.Print(err)
	}
}

func bar(ctx context.Context) {
	tr := otel.Tracer("component-bar")
	ctx, span := tr.Start(ctx, "bar")
	defer span.End()

	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	fmt.Printf("SpanID: %+v\n", span.SpanContext().SpanID())
	fmt.Printf("TraceID: %+v\n", span.SpanContext().TraceID())
	fmt.Printf("SpanContext: %+v\n", span.SpanContext())

	url := flag.String("server", "http://localhost:7777/hello", "server url")
	flag.Parse()

	req, _ := http.NewRequestWithContext(ctx, "GET", *url, nil)
	fmt.Printf("Sending request...\n")
	_, err := client.Do(req)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Response Received!\n")
}

/*
func createMultimediaClient() *multimedia.Client {
	token := "v2.local.qBoKOeogx9bIxIyCly4VWOMt49V8B6MUcveV5CTEKcZ4SPNcgHEHIHipMLJjAq6brIZqa4sifO9rpVp8F0XRWPjQ6oolCbs2ox5q0yftcatp21k29FbLKwp28Q18TOOfQ_IIldYQ-X9xTR3tAEyhtPxiU3GjslfN2cg5cIqeodo.eyJpZCI6MjU4NDA3fQ"

	//defaultTransport := &http.Transport{
	//	DialContext: (&net.Dialer{
	//		Timeout:   2 * time.Second,
	//		KeepAlive: 1 * time.Second,
	//	}).DialContext,
	//	TLSHandshakeTimeout:   10 * time.Second,
	//	IdleConnTimeout:       10 * time.Second,
	//	MaxIdleConns:          10,
	//	MaxIdleConnsPerHost:   1,
	//	ExpectContinueTimeout: 1 * time.Second,
	//}
	transport := otelhttp.NewTransport(nil)

	authMiddleware := func(token string) clientv2.RequestInterceptor {
		return func(ctx context.Context,
			req *http.Request,
			gqlInfo *clientv2.GQLRequestInfo,
			res interface{},
			next clientv2.RequestInterceptorFunc) error {
			req.Header.Set("Authorization", token)

			return next(ctx, req, gqlInfo, res)
		}
	}

	multimediaClient := multimedia.NewClient(
		&http.Client{
			Transport: transport,
		},
		"http://localhost:8092/graphql",
		authMiddleware(fmt.Sprintf("Bearer %s", token)),
	)

	return multimediaClient
}
*/
