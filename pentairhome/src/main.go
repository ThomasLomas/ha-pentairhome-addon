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
)

func main() {
	// App will run until cancelled by user (e.g. ctrl-c)
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

	mqttClient, err := mqtt.MakeClient(mqtt.MQTTConfig{
		Context:  ctx,
		Host:     runtimeConfiguration.MQTTHost,
		Port:     runtimeConfiguration.MQTTPort,
		Username: runtimeConfiguration.MQTTUsername,
		Password: runtimeConfiguration.MQTTPassword,
	})

	if err != nil {
		panic(err)
	}

	apiClient := makeApiClient(ctx, *runtimeConfiguration)

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

	device, err := apiClient.GetDevice(devices[intelliConnectIdx].DeviceID)

	if err != nil {
		panic(err)
	}

	sendSensorConfig(mqttClient, device)

	sendSensorData(mqttClient, device)
	pollSensorData(ctx, mqttClient, apiClient, device.DeviceID)

	<-mqttClient.Client.Done() // Wait for clean shutdown (cancelling the context triggered the shutdown)
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

func pollSensorData(ctx context.Context, mqttClient *mqtt.MQTTWrapper, apiClient *api.APIClient, deviceId string) {
	ticker := time.NewTicker(60 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				defer func() {
					if r := recover(); r != nil {
						apiClient = makeApiClient(ctx, *config.FetchRuntimeConfiguration())
						log.Println("Recovered from panic in sensor data polling")
					}
				}()

				device, err := apiClient.GetDevice(deviceId)

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
		sensor.GenerateSensorConfig(device, "Actual Power", "power", "power"),
		sensor.GenerateSensorConfig(device, "Actual Speed", "actualspeed", "speed"),
		sensor.GenerateSensorConfig(device, "Actual Flow", "actualflow", "volume_flow_rate"),
		sensor.GenerateSensorConfig(device, "Actual Temp", "actualtemp", "temperature"),
	}

	for _, config := range sensorConfigs {
		message, err := json.Marshal(config)
		if err != nil {
			panic(err)
		}

		topic := fmt.Sprintf("homeassistant/sensor/%s/config", config.UniqueID)
		_, err = mqttClient.Publish(topic, message)

		if err != nil {
			panic(err)
		}
	}
}

func sendSensorData(mqttClient *mqtt.MQTTWrapper, device *api.Device) {
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

	sensorData := sensor.SensorData{
		Power:       power,
		ActualSpeed: actualSpeed,
		ActualFlow:  actualFlow,
		ActualTemp:  actualTemp,
	}

	sensorDataJSON, err := json.Marshal(sensorData)

	if err != nil {
		panic(err)
	}

	topic := fmt.Sprintf("pentairhome/%s", device.DeviceID)
	_, err = mqttClient.Publish(topic, sensorDataJSON)

	if err != nil {
		panic(err)
	}
}
