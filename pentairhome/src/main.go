package main

import (
	"context"
	"encoding/json"
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
	runtimeConfiguration.ValidateRuntimeConfiguration()

	mqttClient := mqtt.MakeClient(mqtt.MQTTConfig{
		Context:  ctx,
		Host:     runtimeConfiguration.MQTTHost,
		Port:     runtimeConfiguration.MQTTPort,
		Username: runtimeConfiguration.MQTTUsername,
		Password: runtimeConfiguration.MQTTPassword,
	})

	identity := cognito.AuthenticateWithUsernameAndPassword(ctx, runtimeConfiguration.PentairHomeUsername, runtimeConfiguration.PentairHomePassword)
	credentials := cognito.GetCredentialsFromAuthentication(ctx, identity)

	apiClient := api.NewAPIClient(ctx, *identity.IdToken, *credentials.AccessKeyId, *credentials.SecretKey, *credentials.SessionToken)

	devices := apiClient.ListDevices()

	intelliConnectIdx := sort.Search(len(devices), func(i int) bool {
		return devices[i].IsIntelliConnect()
	})

	if intelliConnectIdx < 0 {
		log.Fatal("No IntelliConnect devices found")
	}

	device := apiClient.GetDevice(devices[intelliConnectIdx].DeviceID)
	sendSensorConfig(mqttClient, device)

	sendSensorData(mqttClient, device)
	pollSensorData(ctx, mqttClient, apiClient, device.DeviceID)

	<-mqttClient.Client.Done() // Wait for clean shutdown (cancelling the context triggered the shutdown)
}

func pollSensorData(ctx context.Context, mqttClient mqtt.MQTTWrapper, apiClient *api.APIClient, deviceId string) {
	ticker := time.NewTicker(60 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				device := apiClient.GetDevice(deviceId)
				sendSensorData(mqttClient, device)
			case <-ctx.Done():
				log.Println("Shutting down sensor data polling")
				ticker.Stop()
				return
			}
		}
	}()
}

func sendSensorConfig(mqttClient mqtt.MQTTWrapper, device api.Device) {
	sensorConfigs := []sensor.SensorConfig{
		sensor.GenerateSensorConfig(device, "Actual Power", "power", "power"),
		sensor.GenerateSensorConfig(device, "Actual Speed", "actualspeed", "speed"),
		sensor.GenerateSensorConfig(device, "Actual Flow", "actualflow", "volume_flow_rate"),
		sensor.GenerateSensorConfig(device, "Actual Temp", "actualtemp", "temperature"),
	}

	for _, config := range sensorConfigs {
		message, _ := json.Marshal(config)
		topic := fmt.Sprintf("homeassistant/sensor/%s/config", config.UniqueID)
		mqttClient.Publish(topic, message)
	}
}

func sendSensorData(mqttClient mqtt.MQTTWrapper, device api.Device) {
	sensorData := sensor.SensorData{
		Power:       device.GetActualPower(),
		ActualSpeed: device.GetActualSpeed(),
		ActualFlow:  device.GetActualFlow(),
		ActualTemp:  device.GetActualTemp(),
	}

	sensorDataJSON, _ := json.Marshal(sensorData)
	topic := fmt.Sprintf("pentairhome/%s", device.DeviceID)
	mqttClient.Publish(topic, sensorDataJSON)
}
