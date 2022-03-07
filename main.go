package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jhawk7/go-opentel/opentel"
	"github.com/jhawk7/rpi-thermometer/pkg/common"
	log "github.com/sirupsen/logrus"
	rpio "github.com/stianeikeland/go-rpio/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/metric"
)

func main() {
	// initialize meter and trace proivders
	if opentelErr := opentel.InitOpentelProviders(); opentelErr != nil {
		common.ErrorHandler(fmt.Errorf("failed to initialize opentel providers; %v", opentelErr), true)
	}

	defer func() {
		if shutdownErr := opentel.ShutdownOpentelProviders(); shutdownErr != nil {
			common.ErrorHandler(fmt.Errorf("failed to stop opentel providers; %v", shutdownErr), false)
		}

		//close rpio pin addresses
		rpio.Close()
	}()

	//start temp sensor
	go readTemperature()

	r := gin.New()
	r.Use(otelgin.Middleware("rpi-thermometer", otelgin.WithTracerProvider(opentel.GetTraceProvider())))
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
	log.Info("starting temp reading..")

	//creates meter and counter via opentel meter provider
	thermometer := opentel.GetMeterProvider().Meter("rpi-thermometer")
	tempLogger, ctrErr := thermometer.NewFloat64Counter("rpi-thermometer.temp", metric.WithDescription("logs temperature in F"))
	if ctrErr != nil {
		panic(fmt.Errorf("failed to create temp logger; %v", ctrErr))
	}

	for {
		/*
			Temperature sensor input voltage relates to actual temp
			reading is in mV; input using output voltage of 3.3v
			voltage at pin in mv = ADC_read * 3300/1024
			tempC = (volts - 500) / 10
			tempF = tempc * 9 / 5 +32

			Temp in °C = [(Vout in mV) - 500] / 10
			(_pin.read()*3.3)-0.500)*100.0;
			tempF=(9.0 * myTMP36.read())/5.0 + 32.0;*/

		read := pin.Read() // Read state from pin (High / Low) in miliVolts
		voltage := float64(read) * (3300.0 / 1024.0)
		tempC := (voltage - 500.0) / 10.0
		tempF := (tempC*9.0)/5.0 + 32.0
		tempLogger.Measurement(float64(tempF))
		fmt.Printf("TempF: %v\n", tempF)
		time.Sleep(5 * time.Second)
	}
}
