package common

import (
	"fmt"
	"os"

	"github.com/jhawk7/rpi-thermometer/pkg/opentel"
	"go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/trace"
)

var exporterUrl string = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
var environent string = os.Getenv("environment")
var serviceName string = "rpi-thermometer"

type Config struct {
	MeterProvider *basic.Controller
	TraceProvider *trace.TracerProvider
}

func GetConfig() *Config {
	mp, mpErr := opentel.InitMeterProvider(exporterUrl, serviceName, environent)
	mpErr = fmt.Errorf("failed to initialize meter provider; [mpErr: %v]", mpErr)
	ErrorHandler(mpErr, true)

	tp, tpErr := opentel.InitTraceProvider(exporterUrl, serviceName, environent)
	tpErr = fmt.Errorf("failed to initialize trace provider; [tpErr: %v]", tpErr)
	ErrorHandler(tpErr, true)

	config := Config{
		MeterProvider: mp,
		TraceProvider: tp,
	}

	return &config
}
