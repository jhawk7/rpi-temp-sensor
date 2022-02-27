package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jhawk7/rpi-thermostat/pkg/opentel"
	log "github.com/sirupsen/logrus"
	rpio "github.com/stianeikeland/go-rpio"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

//pins correspond to GPIO pin #s and not physical pin #s
const pin = rpio.Pin(2)

var thermometer metric.Meter

func main() {
	//setup gin gonic with otel middleware
	//setup health check endpoint
	//setup endpoint to change voltage read intervals?

	//init global meter provider
	mp, mpErr := opentel.InitMeterProvider()
	if mpErr != nil {
		log.Fatal(mpErr)
	}

	tp, tpErr := opentel.InitTraceProvider()
	if tpErr != nil {
		log.Fatal(tpErr)
	}

	//register meterProvider as global mp for package (meterProvider -> meter -> counter)
	//register traceProvider as global tp for package
	global.SetMeterProvider(mp)
	otel.SetTracerProvider(tp)

	//start metric collection
	ctx := context.Background()
	if collectErr := mp.Start(ctx); collectErr != nil {
		log.Fatal(collectErr)
	}

	//create meter from meter provider (set to global variable)
	thermometer = global.Meter("thermometer")

	defer func() {
		if stopErr := mp.Stop(ctx); stopErr != nil {
			log.Fatal(stopErr)
		}

		if shutdownErr := tp.Shutdown(ctx); shutdownErr != nil {
			log.Fatal(shutdownErr)
		}
	}()

	//create meter from meter provider (set to global variable)
	//ds_meter = global.Meter("deathstar_meter")
	go readTemperature()

	r := gin.New()
	r.Use(otelgin.Middleware("rpi-thermostat"))
	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}

func readTemperature() {
	for {
		// Temperature sensor input voltage relates to actual temp
		//Temp in °C = [(Vout in mV) - 500] / 10
		//((_pin.read()*3.3)-0.500)*100.0;
		// tempF=(9.0 * myTMP36.read())/5.0 + 32.0;
		//create metric for temp reads
		statCtr, _ := thermometer.NewInt64Counter("thermostat.temp", metric.WithDescription("logs temperature in F"))
		pin.Input()          // Input mode
		vstate := pin.Read() // Read state from pin (High / Low)
		read := ((float64(vstate) * 3.3) - 0.5) * 100.0
		tempF := (9.0*read)/5.0 + 32.0
		statCtr.Measurement(int64(tempF))
		fmt.Sprintf("TempF: %f", tempF)
		time.Sleep(1 * time.Second)
	}
}
