import machine
import network
import config
import json
from mqtt.simple import MQTTClient, MQTTException
from time import sleep

sleep(2)
led = machine.Pin("LED", machine.Pin.OUT)
MAX_RETRIES=3

class mqttClient:
  def __init__(self):
    self.isConnected = False
    client = self.__connectMQTT()
    self.client = client
    self.topic = config.ENV["MQTT_TOPIC"]
  
  def __connectMQTT(self, counter=1):
    client = MQTTClient(client_id=b"picow_thermo",
      server = config.ENV["MQTT_SERVER"],
      port = 1883,
      user = config.ENV["MQTT_USER"],
      password = config.ENV["MQTT_PASS"],
      keepalive=10,
      ssl=False
    )
    
    try:
      client.connect()
    except MQTTException:
      led.value(False)
      sleep(1)
      led.value(True)
      print("failed to connect to mqtt server")
      if counter != MAX_RETRIES:
        counter += 1
        sleep(2)
        return self.__connectMQTT(counter) #retry

      led.value(False)
      print("max mqtt retries reached.. backing off")
      return client
      
    else:
      led.value(False)
      self.isConnected = True
      print("connected to mqtt server")
      return client
  
  def publish(self, temp, humidity):
    obj = {"tempF": temp, "humidity": humidity, "action": "log"}
    msg = json.dumps(obj)
    self.client.publish(self.topic, msg)
    print(f"published values to topic {self.topic}")
  
  def disconnect(self):
    self.isConnected = False
    print('disconnecting from mqtt server')
    self.client.disconnect()
    

class wifi:
  def __init__(self):
    self.wlan = self.__connectWifi()
    
  def __connectWifi(self, counter=1):
    print('Connecting to WiFi Network Name:', config.ENV["SSID"])
    wlan = network.WLAN(network.STA_IF)
    wlan.active(True) # power up the WiFi chip
    print('Waiting for wifi chip to power up...')
    sleep(3) # wait three seconds for the chip to power up and initialize
    wlan.connect(config.ENV["SSID"], config.ENV["WPASS"])
    print('Waiting for access point to log us in.')
    sleep(5)
    
    if wlan.isconnected():
      print('Success! We have connected to your access point!')
      print('Try to ping the device at', wlan.ifconfig()[0])
      led.value(False)
      return wlan
    elif MAX_RETRIES != counter:
      print('Failure! We have not connected to your access point!  Check your config file for errors.')
      led.value(False)
      sleep(1)
      led.value(True)
      counter += 1
      print("reconnecting")
      return self.__connectWifi(counter) #retry
    else:
      led.value(False)
      print("reached max retries for wifi.. backing off")
      return wlan
  
  def disconnect(self):
    print('disconnecting from wifi')
    self.wlan.disconnect()
    self.wlan.active(False) #power down wlan chip


def getReading(i2c):
  i2c.writeto(0x44, bytes([0x2C, 0x06]))
  sleep(1)
  data = i2c.readfrom(0x44, 6)
  
  tempF = ((data[0] << 8) + data[1]) * 315.0 / 65535.0 - 49.0
  humidity = ((data[3] << 8) + data[4]) * 100.0 / 65535.0
  print(tempF, humidity)
  return tempF, humidity
  
  
def main():
  i2c = machine.I2C(1, scl=machine.Pin(15), sda=machine.Pin(14), freq=200_000)
  print(i2c.scan())
  
  while True:
    led.value(True) # LED will be on until wifi is connected successfully
    wconn = wifi()
    if wconn.wlan.isconnected():
      sleep(2)
      led.value(True) # LED will remain on until mqtt is connected successfully
      cMQTT = mqttClient()
      if cMQTT.isConnected:
        temp, humidity = getReading(i2c)
        cMQTT.publish(temp, humidity)
        sleep(2)
        cMQTT.disconnect()
        wconn.disconnect()
    
    print("entering power saver mode..")
    sleep(1800)


if __name__ == "__main__":
  main()
