package config

import (
	"flag"
	"fmt"
	"log"
)

type RuntimeConfiguration struct {
	PentairHomeUsername string
	PentairHomePassword string
	MQTTHost            string
	MQTTPort            string
	MQTTUsername        string
	MQTTPassword        string
}

func (config *RuntimeConfiguration) ValidateRuntimeConfiguration() {
	if config.PentairHomeUsername == "" {
		flag.PrintDefaults()
		log.Fatal("Pentair Home username is required")
	}
	if config.PentairHomePassword == "" {
		flag.PrintDefaults()
		log.Fatal("Pentair Home password is required")
	}
	if config.MQTTHost == "" {
		flag.PrintDefaults()
		log.Fatal("MQTT host is required")
	}
	if config.MQTTPort == "" {
		flag.PrintDefaults()
		log.Fatal("MQTT port is required")
	}
	if config.MQTTUsername == "" {
		flag.PrintDefaults()
		log.Fatal("MQTT username is required")
	}
	if config.MQTTPassword == "" {
		flag.PrintDefaults()
		log.Fatal("MQTT password is required")
	}
}

func FetchRuntimeConfiguration() *RuntimeConfiguration {
	pentairHomeUsernamePtr := flag.String("pentairhome_username", "", "Pentair Home username")
	pentairHomePasswordPtr := flag.String("pentairhome_password", "", "Pentair Home password")
	mqttHostPtr := flag.String("mqtt_host", "", "MQTT host")
	mqttPortPtr := flag.String("mqtt_port", "", "MQTT port")
	mqttUsernamePtr := flag.String("mqtt_username", "", "MQTT username")
	mqttPasswordPtr := flag.String("mqtt_password", "", "MQTT password")
	flag.Parse()

	return &RuntimeConfiguration{
		PentairHomeUsername: *pentairHomeUsernamePtr,
		PentairHomePassword: *pentairHomePasswordPtr,
		MQTTHost:            *mqttHostPtr,
		MQTTPort:            *mqttPortPtr,
		MQTTUsername:        *mqttUsernamePtr,
		MQTTPassword:        *mqttPasswordPtr,
	}
}

type Configuration struct {
	AWSRegion         string
	AWSUserPoolID     string
	AWSClientID       string
	AWSIdentityPoolId string
}

func (c Configuration) GetLoginKey() string {
	return fmt.Sprintf("cognito-idp.%s.amazonaws.com/%s", c.AWSRegion, c.AWSUserPoolID)
}

func FetchConfiguration() *Configuration {
	return &Configuration{
		AWSRegion:         "us-west-2",
		AWSUserPoolID:     "us-west-2_lbiduhSwD",
		AWSClientID:       "3de110o697faq7avdchtf07h4v",
		AWSIdentityPoolId: "us-west-2:6f950f85-af44-43d9-b690-a431f753e9aa",
	}
}
