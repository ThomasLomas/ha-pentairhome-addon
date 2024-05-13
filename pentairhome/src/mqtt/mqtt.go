package mqtt

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

type MQTTConfig struct {
	Context  context.Context
	Host     string
	Port     string
	Username string
	Password string
}

type MQTTWrapper struct {
	Client  *autopaho.ConnectionManager
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

	u, err := url.Parse(fmt.Sprintf("mqtt://%s:%s", config.Host, config.Port))
	if err != nil {
		log.Fatalf("failed to parse URL: %s", err)
	}

	cliCfg := autopaho.ClientConfig{
		ServerUrls:                    []*url.URL{u},
		KeepAlive:                     20,
		CleanStartOnInitialConnection: true,
		SessionExpiryInterval:         60,
		OnConnectError:                func(err error) { log.Printf("error whilst attempting connection: %s\n", err) },
		ClientConfig: paho.ClientConfig{
			ClientID: "pentairhome",
		},
		ConnectUsername: config.Username,
		ConnectPassword: []byte(config.Password),
	}

	c, err := autopaho.NewConnection(config.Context, cliCfg)

	if err != nil {
		log.Fatalf("failed to create connection: %s", err)
	}

	if err = c.AwaitConnection(config.Context); err != nil {
		log.Fatalf("failed to connect: %s", err)
	}

	fmt.Printf("Connected to %s\n", u)

	return MQTTWrapper{
		Client:  c,
		Context: config.Context,
	}
}
