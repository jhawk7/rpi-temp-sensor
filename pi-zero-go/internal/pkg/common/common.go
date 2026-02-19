package common

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func ErrorHandler(err error, fatal bool) {
	if err != nil {
		log.Errorf("error: %v", err)

		if fatal {
			panic(err)
		}
	}
}

type Config struct {
	MQTTServer string
	MQTTPort   string
	MQTTUser   string
	MQTTPass   string
	MQTTTopic  string
}

func LoadConfig() Config {
	server, isValid := os.LookupEnv("MQTT_SERVER")
	if !isValid {
		log.Fatal("MQTT_SERVER env var not set")
	}

	port, isValid := os.LookupEnv("MQTT_PORT")
	if !isValid {
		log.Fatal("MQTT_PORT env var not set")
	}

	user, isValid := os.LookupEnv("MQTT_USER")
	if !isValid {
		log.Fatal("MQTT_USER env var not set")
	}

	pass, isValid := os.LookupEnv("MQTT_PASS")
	if !isValid {
		log.Fatal("MQTT_PASS env var not set")
	}

	topic, isValid := os.LookupEnv("MQTT_TOPIC")
	if !isValid {
		log.Fatal("MQTT_TOPIC env var not set")
	}

	return Config{
		MQTTServer: server,
		MQTTPort:   port,
		MQTTUser:   user,
		MQTTPass:   pass,
		MQTTTopic:  topic,
	}
}
