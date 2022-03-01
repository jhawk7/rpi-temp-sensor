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
		log.Fatal(collectErr)
	}

	defer func() {
		if stopErr := config.MeterProvider.Stop(ctx); stopErr != nil {
			log.Fatal(stopErr)
		}

		if shutdownErr := config.TraceProvider.Shutdown(ctx); shutdownErr != nil {
			log.Fatal(shutdownErr)
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
	gpioErr := rpio.Open()
	log.Fatal(gpioErr)

	pin := rpio.Pin(2)
	pin.Input() // Input mode

	//create meter from global mp
	thermometer := global.Meter("rpi-thermometer")

	for {
		/*Temperature sensor input voltage relates to actual temp
		Temp in °C = [(Vout in mV) - 500] / 10
		(_pin.read()*3.3)-0.500)*100.0;
		empF=(9.0 * myTMP36.read())/5.0 + 32.0;*/

		//create metric for temp reads
		tempCtr, _ := thermometer.NewInt64Counter("rpi-thermometer.temp", metric.WithDescription("logs temperature in F"))
		voltage := pin.Read() // Read state from pin (High / Low)
		read := ((float64(voltage) * 3.3) - 0.5) * 100.0
		tempF := (9.0*read)/5.0 + 32.0
		tempCtr.Measurement(int64(tempF))
		fmt.Sprintf("TempF: %f", tempF)
		time.Sleep(1 * time.Second)
	}
}
