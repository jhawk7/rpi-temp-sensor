package main

import (
	"context"
	"fmt"

	i2c "github.com/d2r2/go-i2c"
	"github.com/gin-gonic/gin"
	"github.com/jhawk7/go-opentel/opentel"
	"github.com/jhawk7/rpi-thermometer/pkg/common"
	_ "github.com/sirupsen/logrus"
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
	/*
		# SHT31 address, 0x44(68)
		# Read data back from 0x00(00), 6 bytes
		# Temp MSB, Temp LSB, Temp CRC, Humididty MSB, Humidity LSB, Humidity CRC
	*/

	// Create new connection to I2C bus on line 4 with address 0x44
	conn, connErr := i2c.NewI2C(0x44, 4)
	if connErr != nil {
		common.ErrorHandler(fmt.Errorf("failed to connect to i2c peripheral device; %v", connErr), true)
	}
	defer conn.Close()

	//creates meter and counter via opentel meter provider
	thermometer := opentel.GetMeterProvider().Meter("rpi-thermometer")
	tempLogger, ctrErr := thermometer.NewFloat64Counter("rpi-thermometer.temp", metric.WithDescription("logs temperature in F"))
	if ctrErr != nil {
		panic(fmt.Errorf("failed to create temp logger; %v", ctrErr))
	}
	ctx := context.Background()

	for {
		// buffer of len 4
		buf := make([]byte, 4)
		read, readErr := conn.ReadBytes(buf)
		if readErr != nil {
			common.ErrorHandler(fmt.Errorf("failed to read bytes from i2c device; %v", readErr), false)
		}
		tempLogger.Add(ctx, float64(read))
		fmt.Printf("Value read from readbytes: %v", read)
		//ftemp := (buf[0] * uint(256))
	}

	/*
		// Convert the data
		double cTemp = (((data[0] * 256) + data[1]) * 175.0) / 65535.0  - 45.0;
		double fTemp = (((data[0] * 256) + data[1]) * 315.0) / 65535.0 - 49.0;
		double humidity = (((data[3] * 256) + data[4])) * 100.0 / 65535.0;
	*/
}
