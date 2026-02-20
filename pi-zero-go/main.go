package main

import (
	"fmt"
	"time"

	i2c "github.com/d2r2/go-i2c"
	"github.com/gin-gonic/gin"
	"github.com/jhawk7/rpi-thermometer/internal/pkg/common"
	"github.com/jhawk7/rpi-thermometer/internal/pkg/mqttc"
	log "github.com/sirupsen/logrus"
)

var i2cConn *i2c.I2C

func main() {
	config := common.LoadConfig()

	/*
	* Create new connection to I2C bus on line 1 with address 0x44 (default SHT31-D address location)
	* run `i2cdetect -y 1` to view vtable for specific device addr
	* when loaded, a specific device entry folder /dev/i2c-* will be created; using bus 1 for /dev/i2c-1
	 */
	conn, connErr := i2c.NewI2C(0x44, 1)
	if connErr != nil {
		common.ErrorHandler(fmt.Errorf("failed to connect to i2c peripheral device; %v\n", connErr), true)
	}
	defer conn.Close()
	i2cConn = conn

	// connect to mqtt broker and subscribe to topic for remote temp read requests
	mqttClient := mqttc.InitMQTTClient(config)
	defer mqttClient.Disconnect()

	go readTemperature(mqttClient)

	r := gin.New()
	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 */
}

func readTemperature(mqttClient *mqttc.MQTTClient) {
	for {
		// read temp and humidity from i2c device every 5 minutes
		temp, humidity := getReading()
		// push to mqtt broker
		mqttClient.Publish(temp, humidity)
		time.Sleep(600 * time.Second)
	}
}

func getReading() (float64, float64) {
	// send repeatable measurement command to i2c device to begin reading temp and humidity
	// Command msb and command lsb (0x2C, 0x06)
	wbuf := []byte{0x2C, 0x06}
	wlen, wErr := i2cConn.WriteBytes(wbuf)
	if wErr != nil {
		common.ErrorHandler(fmt.Errorf("failed to write cmd to i2c device; %v", wErr), true)
	}
	log.Infof("writing %v bytes to i2c device\n", wlen)

	// read 6 bytes of data for: temp msb, temp lsb, temp CRC, humidity msb, humidity lsb, humidity CRC
	rbuf := make([]byte, 6)
	rlen, readErr := i2cConn.ReadBytes(rbuf)
	if readErr != nil {
		common.ErrorHandler(fmt.Errorf("failed to read bytes from i2c device; %v\n", readErr), true)
	}
	log.Infof("%v bytes read from i2c device\n", rlen)

	/*
		convert response bytes to readable data
		double cTemp = (((data[0] * 256) + data[1]) * 175.0) / 65535.0  - 45.0;
		double fTemp = (((data[0] * 256) + data[1]) * 315.0) / 65535.0 - 49.0;
		double humidity = (((data[3] * 256) + data[4])) * 100.0 / 65535.0;
	*/

	ftemp := ((float32(rbuf[0])*256+float32(rbuf[1]))*315.0)/65535.0 - 49.0
	humidity := (float32(rbuf[3])*256 + float32(rbuf[4])) * 100.0 / 65535.0
	log.Infof("Temp: %.2f F\n Humidity: %.2f RH\n", ftemp, humidity)
	return float64(ftemp), float64(humidity)
}
