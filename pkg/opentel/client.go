package opentel

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	_ "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func InitTraceProvider(exporterUrl string, serviceName string, environment string) (tp *sdktrace.TracerProvider, tpErr error) {
	//configure grpc exporter
	if exporterUrl == "" {
		exporterUrl = "localhost:4317"
	}

	exporter, expErr := otlptracegrpc.New(context.Background(), otlptracegrpc.WithEndpoint(exporterUrl))
	if expErr != nil {
		tpErr = expErr
		return
	}

	//configure trace provider resource to describe this application
	r := getAppResource(serviceName, environment)

	//register exporter with new trace provider
	tp = sdktrace.NewTracerProvider(
		//register exporter with trace provider using BatchSpanProcessor
		sdktrace.WithBatcher(exporter),
		//configure resource to be used in all traces from trace provider
		sdktrace.WithResource(r),
		//setup sampler to always sample traces
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	return
}

func InitMeterProvider(exporterUrl string, serviceName string, environment string) (mp *controller.Controller, mpErr error) {
	if exporterUrl == "" {
		exporterUrl = "localhost:4317"
	}

	exporter, expErr := otlpmetricgrpc.New(context.Background(), otlpmetricgrpc.WithEndpoint(exporterUrl))
	if expErr != nil {
		mpErr = expErr
		return
	}

	r := getAppResource(serviceName, environment)

	mp = controller.New(
		processor.NewFactory(
			simple.NewWithHistogramDistribution(),
			exporter,
		),
		//configure exporter for metrics
		controller.WithExporter(exporter),
		//configure resource for metrics
		controller.WithResource(r),
		controller.WithCollectPeriod(10*time.Second),
	)

	return
}

func getAppResource(serviceName string, environment string) *resource.Resource {
	//configures resource to describe this application
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("environment", environment),
		),
	)
	return r
}

// returns a standard console exporter.
/*func newStdExporter(w io.Writer) (sdktrace.SpanExporter, error) {
	// Write telemetry data to a file.
	os.Remove("traces.txt")
	f, err := os.Create("traces.txt")
	if err != nil {
		log.Fatal(err)
	}
	outfile = f

	return stdout.New(
		stdout.WithWriter(w),
		// Use human-readable output.
		stdout.WithPrettyPrint(),
		// Do not print timestamps for the demo.
		stdout.WithoutTimestamps(),
	)
}*/

//returns http exporter
/*
func newHttpExporter() (exporter *otlptrace.Exporter, err error) {
	exporter, err = otlptrace.New(context.Background(), otlptracehttp.NewClient(otlptracehttp.WithEndpoint(COLLECTER_URL)))
	if err != nil {
		fmt.Errorf("error initializing exporter [error: %v]", err)
		//log.Fatal(err)
		return
	}

	return
}
*/
