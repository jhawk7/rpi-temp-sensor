import machine
import network
import config
import json
from mqtt.simple import MQTTClient, MQTTException
from time import sleep

sleep(2)  # wait for the system to stabilize
print("Starting PicoW Temp Sensor...")
LED = machine.Pin("LED", machine.Pin.OUT)
MAX_RETRIES = 3

class mqttClient:
  def __init__(self):
    self.isConnected = False
    self.topic = config.ENV["MQTT_TOPIC"]
    self.client = self.__connectMQTT()
  
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
      print("Connecting to MQTT Server:", config.ENV["MQTT_SERVER"])
      client.connect()
    except Exception as e:
      LED.value(False)
      sleep(1)
      LED.value(True)
      print("failed to connect to mqtt server:", e)
      if counter < MAX_RETRIES:
        counter += 1
        sleep(1)
        return self.__connectMQTT(counter) #retry

      LED.value(False)
      sleep(0.5)
      doubleBlink()
      print("max mqtt retries reached.. backing off")
      return client
      
    else:
      LED.value(False)
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
      LED.value(False)
      return wlan
    elif counter < MAX_RETRIES:
      print('Failure! We have not connected to your access point!  Check your config file for errors.')
      LED.value(False)
      sleep(1)
      LED.value(True)
      counter += 1
      print("reconnecting")
      return self.__connectWifi(counter) #retry
    else:
      LED.value(False)
      sleep(0.5)
      doubleBlink()
      print("reached max retries for wifi.. backing off")
      return wlan
  
  def disconnect(self):
    print('disconnecting from wifi')
    self.wlan.disconnect()
    self.wlan.active(False) #power down wlan chip

def doubleBlink():
  LED.value(True)
  sleep(0.5)
  LED.value(False)
  sleep(0.5)
  LED.value(True)
  sleep(0.5)
  LED.value(False)

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
    LED.value(True) # LED will be on until wifi is connected successfully
    wconn = wifi()
    if wconn.wlan.isconnected():
      sleep(2)
      LED.value(True) # LED will remain on until mqtt is connected successfully
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
