package main

import (
	"machine"
	sht3x "tinygo.org/x/drivers/sht3x"
)

func main() {
	// i2c connection
	/*
	* Create new connection to I2C bus on line 1 with address 0x44 (default SHT31-D address location)
	* run `i2cdetect -y 1` to view vtable for specific device addr
	* when loaded, a specific device entry folder /dev/i2c-* will be created; using bus 1 for /dev/i2c-1
	 */

	i2c := machine.I2C0
	i2cErr := i2c.Configure(machine.I2CConfig{
		SCL: machine.P0_30,
		SDA: machine.P0_31,
	})

	if i2cErr != nil {
		fmt.Println("could not configure I2C:", i2cErr)

	}

	/* example of sending a cmd and reading output for i2c; sht3x lib will do this for us
	// bytes for register we will write output to
	w := []byte{0x75}

	// we'll read 1 byte
	r := make([]byte, 1)

	//i2c transaction will communicate with reg 0x68 and store byte from reg 0x75 in our read buffer
	txErr := i2c.Tx(0x68, w, r)
	if txErr != nil {
		println("could not interact with I2C device:", txErr)
		return
	}

	println("WHO_AM_I:", r[0]) // prints "WHO_AM_I: 104"
	*/

	sensor := sht3x.New(i2c)
	tempC, relativeH := sensor.ReadTemperatureHumidity
	tempF := (tempC * 9 / 5) + 32
	fmt.Println(tempF, relativeH)
}
