package mqtt

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/eclipse/paho.golang/paho"
)

type MQTTConfig struct {
	Context  context.Context
	Host     string
	Port     string
	Username string
	Password string
}

func (c *MQTTConfig) GetServer() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type MQTTWrapper struct {
	Client  *paho.Client
	Context context.Context
}

func (mqttWrapper *MQTTWrapper) Publish(topic string, payload []byte) {
	log.Printf("publishing data to topic: %s", topic)

	mqttWrapper.Client.Publish(mqttWrapper.Context, &paho.Publish{
		Topic:   topic,
		QoS:     byte(0),
		Payload: payload,
	})
}

func MakeClient(config MQTTConfig) MQTTWrapper {
	log.Printf("MQTT Host: %s; Port: %s; Username: %s", config.Host, config.Port, config.Username)

	conn, err := net.Dial("tcp", config.GetServer())

	if err != nil {
		log.Fatalf("Failed to connect to %s: %s", config.GetServer(), err)
	}

	c := paho.NewClient(paho.ClientConfig{
		Conn: conn,
	})

	cp := &paho.Connect{
		KeepAlive:  30,
		ClientID:   "pentairhome",
		CleanStart: true,
		Username:   config.Username,
		Password:   []byte(config.Password),
	}

	if config.Username != "" {
		cp.UsernameFlag = true
	}

	if config.Password != "" {
		cp.PasswordFlag = true
	}

	ca, err := c.Connect(config.Context, cp)

	if err != nil {
		log.Fatalln(err)
	}

	if ca.ReasonCode != 0 {
		log.Fatalf("Failed to connect to %s : %d - %s", config.GetServer(), ca.ReasonCode, ca.Properties.ReasonString)
	}

	fmt.Printf("Connected to %s\n", config.GetServer())

	return MQTTWrapper{
		Client:  c,
		Context: config.Context,
	}
}
