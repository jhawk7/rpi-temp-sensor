package common

import (
	"os"

	"github.com/jhawk7/rpi-thermometer/pkg/opentel"
	log "github.com/sirupsen/logrus"
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
	log.Fatal(mpErr)
	tp, tpErr := opentel.InitTraceProvider(exporterUrl, serviceName, environent)
	log.Fatal(tpErr)

	config := Config{
		MeterProvider: mp,
		TraceProvider: tp,
	}

	return &config
}
