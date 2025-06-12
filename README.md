# RPI Temperature and Humidity Sensor 
![Raspberry Pi](https://img.shields.io/badge/-Raspberry_Pi-C51A4A?style=flat&logo=Raspberry-Pi) ![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=flat&logo=go&logoColor=white) ![Python](https://img.shields.io/badge/python-3670A0?style=flat&logo=python&logoColor=ffdd54) ![Prometheus](https://img.shields.io/badge/Prometheus-E6522C?style=flat&logo=Prometheus&logoColor=white) ![Grafana](https://img.shields.io/badge/grafana-%23F46800.svg?style=flat&logo=grafana&logoColor=white) ![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=flat&logo=docker&logoColor=white) ![OpenTelemetry](https://img.shields.io/badge/OpenTelemetry-FFFFFF?&style=flat&logo=opentelemetry&logoColor=black)

![Screenshot 2025-06-12 at 13-35-37 Pi-Thermo (Crawl Space) - Dashboards - Grafana](https://github.com/user-attachments/assets/7f491919-28d1-481e-a25f-7fc5a070674d)
<img src="https://github.com/user-attachments/assets/58045242-33ae-48c9-8fa0-e59d38f3b733" width="400"/>


### Running on Pi Zero (pi-zero-go)

* Uses an [SHT31-D i2c device](https://www.adafruit.com/product/2857?srsltid=AfmBOoo7lEKOvPWaVatMKvZGXeTqnKE-TkIL2cTMc3QbUb8nyVwLrKQq) (temperature and humidity sensor) on raspberry pi (zero or better) to read temperature and humidity and send telemetry metrics to an opentelenetry collector (running on server with prometheus) to be displayed as graph via grafana + prometheus.

### Running on Pico (picow-mpy)

* Refactored for raspberry pi pico to read temperature and humidity, then publish reads to an MQTT server (running independently) for an [MQTT subscriber client](https://github.com/jhawk7/mqtt-sub-client) to process, alert, and forward to a `prometheus instance`. The data fron prometheus can then be visualized in `grafana`. This project uses micro-python for embedded devices (picow).
**TinyGO (picow-go) does not yet support the wifi chip on the picow**

## Flow
#### Connect to i2c device
* (See Go d2r2/go-i2c pkg documentation)
* Create new connection to I2C bus on line 1 with address `0x44` (default SHT31-D address location)
* run `i2cdetect -y 1` to view vtable for specific device addr
* when loaded, a specific device entry folder /dev/i2c-* will be created; using bus 1 for /dev/i2c-1

#### Setup Opentelemetry Gauge Observer via MeterProvider **(Pi Zero Only)**
* (see opentelemetry gauge documentation)
* the gauge observer will continously trigger the registered callback function and observe/record the results
* the callback function, in this case, runs "getReading()" to get the temperature and humidity readings and make them observable

#### Send command to read temp and humidity and convert response bytes to readable data
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

## Helpful Links Used
* Opentelemetry gauge documentation - (https://opentelemetry.io/docs/reference/specification/metrics/api/#asynchronous-gauge)
* Opentel Go pkg - homemade opentel go pkg for handling the openetelmetry setup (https://github.com/jhawk7/go-opentel)
* SHT31-D documentation - http://www.getmicros.net/raspberry-pi-and-sht31-sensor-example-in-c.php
* Useful i2c doc - (https://dave.cheney.net/tag/i2c)
* Go d2r2/go-i2c pkg documentation - (https://github.com/d2r2/go-i2c)
* List of TinyGo Drivers for various hardware components (for embedded) - (https://github.com/tinygo-org/drivers)
* TinyGo working with i2c - (https://tinygo.org/docs/concepts/peripherals/i2c/)
* MicroPython for the pico - (https://docs.micropython.org/en/latest/rp2/quickref.html)
* Connecting pico to mqtt with micropython - (https://www.instructables.com/Connecting-Raspberry-Pi-Pico-Ws-With-MQTT/)
