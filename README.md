# RPI - Thermometer - Pi Zero

* Uses SHT31-D i2c device (temperature and humidity sensor) on raspberry pi (zero or better) to read temperature and humidity, and opentelemetry to send telemetry metrics to a collector (running on separate machine with prometheus backend) to be displayed as graph via prometheus.

# RPI - Thermometer - Embedded

* Refactored for raspberry pi pico to read temperature and humidity, then publish reads to an MQTT server (running independently) for subscribers (such as prometheus, which can use grafana to graph the reads). This project uses tiny-go for compilation on embedded devices (pico)

# Flow
## Connect to i2c device
* (See Go d2r2/go-i2c pkg documentation)
* Create new connection to I2C bus on line 1 with address `0x44` (default SHT31-D address location)
* run `i2cdetect -y 1` to view vtable for specific device addr
* when loaded, a specific device entry folder /dev/i2c-* will be created; using bus 1 for /dev/i2c-1

## Setup Opentelemetry Gauge Observer via MeterProvider **(Pi Zero Only)**
* (see opentelemetry gauge documentation)
* the gauge observer will continously trigger the registered callback function and observe/record the results
* the callback function, in this case, runs "getReading()" to get the temperature and humidity readings and make them observable

## Send command to read temp and humidity and convert response bytes to readable data
* send repeatable measurement command to i2c device to begin reading temp and humidity (command - (0x2C, 0x06) given lsb addressing)
* convert response bytes to redable readings (c++ code snippet)
```
double cTemp = (((data[0] * 256) + data[1]) * 175.0) / 65535.0  - 45.0;
double fTemp = (((data[0] * 256) + data[1]) * 315.0) / 65535.0 - 49.0;
double humidity = (((data[3] * 256) + data[4])) * 100.0 / 65535.0;
```

## Healthcheck **(Pi Zero Only)**
* a healthcheck endpoint is setup on `port 8080` using gin to remotely verify that everything is running smoothly
* `curl http://<device_ip>:8080/healthcheck`

## Dockerization **(Pi Zero Only)**
* the dockerfile builds the go binary in the build stage (GOOS set to linux and GOARCH set to arm for raspberry pi zero w), and executes the binary in the 2nd stage
* port 8080 is exposed in the image for the healthcheck endpoint
* the i2c device entry folder (SHT31-D) is mounted in the container using the `device` flag in the docker-compose file
* with dockerization, we are able to make changes to the code on our local device, build and update the image, then pull the new image down to the raspberry pi from docker hub
* the application can now be started via docker from a cron job or service in linux

# Helpful Links Used
* Opentelemetry gauge documentation - (https://opentelemetry.io/docs/reference/specification/metrics/api/#asynchronous-gauge)
* Opentel Go pkg - homemade opentel go pkg for handling the openetelmetry setup (https://github.com/jhawk7/go-opentel)
* SHT31-D documentation - http://www.getmicros.net/raspberry-pi-and-sht31-sensor-example-in-c.php
* Useful i2c doc - (https://dave.cheney.net/tag/i2c)
* Go d2r2/go-i2c pkg documentation - (https://github.com/d2r2/go-i2c)
* List of TinyGo Drivers for various hardware components (for embedded) - (https://github.com/tinygo-org/drivers)
* TinyGo working with i2c - (https://tinygo.org/docs/concepts/peripherals/i2c/)