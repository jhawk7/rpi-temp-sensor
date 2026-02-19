package mqttc

import (
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jhawk7/rpi-thermometer/internal/pkg/common"
)

var pubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("message published to mqtt; [topic: %v]", msg.Topic())
}

var connHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("successfully connected to mqtt server")
}

var lostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("lost connection to mqtt server; %v", err)
}

type MQTTClient struct {
	conn  *mqtt.Client
	topic string
}

func InitMQTTClient(config common.Config) *MQTTClient {
	//set client options
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%v:%v", config.MQTTServer, config.MQTTPort))
	opts.SetClientID("pi-thermo")
	opts.SetPassword(config.MQTTPass)
	opts.SetUsername(config.MQTTUser)
	opts.SetDefaultPublishHandler(pubHandler)
	opts.OnConnect = connHandler
	opts.OnConnectionLost = lostHandler
	opts.CleanSession = false
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		err := fmt.Errorf("mqtt connection failed; %v", token.Error())
		common.ErrorHandler(err, true)
	}

	return &MQTTClient{
		conn:  &client,
		topic: config.MQTTTopic,
	}
}

func (c *MQTTClient) Publish(temp float64, humidity float64) {
	datamap := map[string]interface{}{
		"tempF":    fmt.Sprintf("%.2f", temp),
		"humidity": fmt.Sprintf("%.2f", humidity),
		"action":   "log",
	}

	jsonBytes, err := json.Marshal(datamap)
	if err != nil {
		e := fmt.Errorf("error creating json bytes from data; %v", err)
		common.ErrorHandler(e, false)
		return
	}
	token := (*c.conn).Publish(c.topic, 1, false, string(jsonBytes))
	token.Wait()
}

func (c *MQTTClient) Disconnect() {
	(*c.conn).Disconnect(250)
}
