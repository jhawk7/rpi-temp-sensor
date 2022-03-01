package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jhawk7/rpi-thermometer/pkg/common"
	log "github.com/sirupsen/logrus"
	rpio "github.com/stianeikeland/go-rpio/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

var config *common.Config

func main() {
	//init config
	config = common.GetConfig()
	//register meterProvider as global mp for package (meterProvider -> meter -> counter)
	//register traceProvider as global tp for package
	global.SetMeterProvider(config.MeterProvider)
	otel.SetTracerProvider(config.TraceProvider)

	//start metric collection
	ctx := context.Background()
	if collectErr := config.MeterProvider.Start(ctx); collectErr != nil {
		err := fmt.Errorf("failed to start metric collector; emessage: %v", collectErr)
		common.ErrorHandler(err, true)
	}

	defer func() {
		if stopErr := config.MeterProvider.Stop(ctx); stopErr != nil {
			err := fmt.Errorf("failed to stop metric collector; emessage: %v", stopErr)
			common.ErrorHandler(err, true)
		}

		if shutdownErr := config.TraceProvider.Shutdown(ctx); shutdownErr != nil {
			err := fmt.Errorf("failed to shutdown trace provider; emessage: %v", shutdownErr)
			common.ErrorHandler(err, true)
		}

		//close rpio pin addresses
		rpio.Close()
	}()

	//start temp sensor
	go readTemperature()

	r := gin.New()
	otelgin.WithTracerProvider(config.TraceProvider)
	r.Use(otelgin.Middleware("rpi-thermometer"))
	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 */
}

func readTemperature() {
	//Open memory range for GPIO access in /dev/mem
	if gpioErr := rpio.Open(); gpioErr != nil {
		err := fmt.Errorf("failed to open mem range for GPIO access; emessage: %v", gpioErr)
		common.ErrorHandler(err, true)
	}

	pin := rpio.Pin(2)
	pin.Input() // Input mode

	//create meter from global mp
	thermometer := global.Meter("rpi-thermometer")
	log.Info("starting temp reading")

	for {
		/*
			Temperature sensor input voltage relates to actual temp
			reading is in mV; input using output voltage of 3.3v
			voltage at pin in mv = ADC_read * 3300/1024
			tempC = (volts - 0.5) * 100
			tempF = tempc * 9 / 5 +32

			Temp in °C = [(Vout in mV) - 500] / 10
			(_pin.read()*3.3)-0.500)*100.0;
			tempF=(9.0 * myTMP36.read())/5.0 + 32.0;*/

		//create metric for temp reads
		tempCtr, _ := thermometer.NewInt64Counter("rpi-thermometer.temp", metric.WithDescription("logs temperature in F"))
		read := pin.Read() // Read state from pin (High / Low) in miliVolts
		voltage := float64(read) * (3300.0 / 1024.0)
		tempC := (voltage - 500.0) / 10.0
		tempF := (tempC*9.0)/5.0 + 32.0
		tempCtr.Measurement(int64(tempF))
		fmt.Printf("TempF: %v\n", tempF)
		time.Sleep(5 * time.Second)
	}
}
