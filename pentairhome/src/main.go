package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"pentairhome/api"
	"pentairhome/cognito"
	"pentairhome/config"
	"pentairhome/mqtt"
	"pentairhome/sensor"
	"sort"
	"syscall"
	"time"

	"github.com/eclipse/paho.golang/paho"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runtimeConfiguration := config.FetchRuntimeConfiguration()
	runtimeConfigErrors := runtimeConfiguration.ValidateRuntimeConfiguration()

	if len(runtimeConfigErrors) > 0 {
		for _, err := range runtimeConfigErrors {
			log.Println(err)
		}
		flag.PrintDefaults()
		os.Exit(1)
	}

	apiClient := makeApiClient(ctx, runtimeConfiguration)

	devices, err := apiClient.ListDevices()

	if err != nil {
		panic(err)
	}

	intelliConnectIdx := sort.Search(len(devices), func(i int) bool {
		return devices[i].IsIntelliConnect()
	})

	if intelliConnectIdx < 0 {
		panic(errors.New("no IntelliConnect devices found"))
	}

	device, deviceErr := apiClient.GetDevice(devices[intelliConnectIdx].DeviceID)

	if deviceErr != nil {
		log.Panicf("failed to get IntelliConnect device: %s", deviceErr)
	}

	mqttClient, mqttErr := mqtt.MakeClient(mqtt.MQTTConfig{
		Context:  ctx,
		Host:     runtimeConfiguration.MQTTHost,
		Port:     runtimeConfiguration.MQTTPort,
		Username: runtimeConfiguration.MQTTUsername,
		Password: runtimeConfiguration.MQTTPassword,
	})

	if mqttErr != nil {
		log.Panicf("failed to create MQTT client: %s", mqttErr)
	}

	sendSensorConfig(mqttClient, device)
	sendSensorData(mqttClient, device)

	pollSensorData(ctx, mqttClient, apiClient, device, runtimeConfiguration)
	listenForStatusMessages(ctx, mqttClient, apiClient, device, runtimeConfiguration)

	<-mqttClient.Client.Done()
}

func makeApiClient(ctx context.Context, runtimeConfiguration config.RuntimeConfiguration) *api.APIClient {
	identity, err := cognito.AuthenticateWithUsernameAndPassword(ctx, runtimeConfiguration.PentairHomeUsername, runtimeConfiguration.PentairHomePassword)

	if err != nil {
		panic(err)
	}

	credentials, err := cognito.GetCredentialsFromAuthentication(ctx, identity)

	if err != nil {
		panic(err)
	}

	return api.NewAPIClient(ctx, *identity.IdToken, *credentials.AccessKeyId, *credentials.SecretKey, *credentials.SessionToken)
}

func listenForStatusMessages(ctx context.Context, mqttClient *mqtt.MQTTWrapper, apiClient *api.APIClient, device *api.Device, runtimeConfiguration config.RuntimeConfiguration) {
	go func() {
		for {
			select {
			case statusMessage := <-mqttClient.StatusMessages:
				log.Printf("Received status message: %s", statusMessage)
				if statusMessage == "online" {
					log.Println("Home Assistant is online")

					defer func() {
						if r := recover(); r != nil {
							log.Println("Recovered from panic in listening for status messages. Making new API client and listening again.")
							apiClient = makeApiClient(ctx, runtimeConfiguration)
							listenForStatusMessages(ctx, mqttClient, apiClient, device, runtimeConfiguration)
							mqttClient.StatusMessages <- statusMessage
						}
					}()

					log.Printf("Sending sensor config for device: %s", device.DeviceID)
					sendSensorConfig(mqttClient, device)
				}
			case <-ctx.Done():
				log.Println("Shutting down status message listener")
				return
			}
		}
	}()
}

func pollSensorData(ctx context.Context, mqttClient *mqtt.MQTTWrapper, apiClient *api.APIClient, device *api.Device, runtimeConfiguration config.RuntimeConfiguration) {
	ticker := time.NewTicker(60 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				defer func() {
					if r := recover(); r != nil {
						log.Println("Recovered from panic in sensor data polling. Making new API client and restarting polling.")
						apiClient = makeApiClient(ctx, runtimeConfiguration)
						pollSensorData(ctx, mqttClient, apiClient, device, runtimeConfiguration)
					}
				}()

				device, err := apiClient.GetDevice(device.DeviceID)

				if err != nil {
					panic(err)
				}

				sendSensorData(mqttClient, device)
			case <-ctx.Done():
				log.Println("Shutting down sensor data polling")
				ticker.Stop()
				return
			}
		}
	}()
}

func sendSensorConfig(mqttClient *mqtt.MQTTWrapper, device *api.Device) {
	sensorConfigs := []sensor.SensorConfig{
		sensor.GenerateSensorConfig(device, "Pump Power", "power", "power", "W"),
		sensor.GenerateSensorConfig(device, "Pump Speed", "actualspeed", "speed", "rpm"),
		sensor.GenerateSensorConfig(device, "Pump Flow", "actualflow", "volume_flow_rate", "gal/min"),
		sensor.GenerateSensorConfig(device, "Water Temperature", "actualtemp", "temperature", "°F"),
		sensor.GenerateSensorConfig(device, "Outside Temperature", "outsidetemp", "temperature", "°F"),
	}

	for _, config := range sensorConfigs {
		message, err := json.Marshal(config)
		if err != nil {
			panic(err)
		}

		topic := fmt.Sprintf("homeassistant/sensor/%s/config", config.UniqueID)
		if _, err = mqttClient.Publish(topic, message); err != nil {
			panic(err)
		} else {
			log.Printf("Published sensor config to %s", topic)
		}
	}
}

func sendSensorData(mqttClient *mqtt.MQTTWrapper, device *api.Device) (pubResp *paho.PublishResponse) {
	power, err := device.GetActualPower()
	if err != nil {
		panic(err)
	}

	actualSpeed, err := device.GetActualSpeed()
	if err != nil {
		panic(err)
	}

	actualFlow, err := device.GetActualFlow()
	if err != nil {
		panic(err)
	}

	actualTemp, err := device.GetActualTemp()
	if err != nil {
		panic(err)
	}

	outsideTemp, err := device.GetOutsideTemp()
	if err != nil {
		panic(err)
	}

	sensorData := sensor.SensorData{
		Power:       power,
		ActualSpeed: actualSpeed,
		ActualFlow:  actualFlow,
		ActualTemp:  actualTemp,
		OutsideTemp: outsideTemp,
	}

	sensorDataJSON, err := json.Marshal(sensorData)

	if err != nil {
		panic(err)
	}

	topic := fmt.Sprintf("pentairhome/%s", device.DeviceID)
	pubResp, err = mqttClient.Publish(topic, sensorDataJSON)

	if err != nil {
		panic(err)
	}

	return pubResp
}
