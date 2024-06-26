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
	Client         *autopaho.ConnectionManager
	Context        context.Context
	StatusMessages chan string
}

func (mqttWrapper *MQTTWrapper) Publish(topic string, payload []byte) (*paho.PublishResponse, error) {
	log.Printf("publishing data to topic: %s", topic)

	resp, err := mqttWrapper.Client.Publish(mqttWrapper.Context, &paho.Publish{
		Topic:   topic,
		QoS:     byte(0),
		Payload: payload,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to publish message: %s", err)
	}

	return resp, nil
}

func MakeClient(config MQTTConfig) (*MQTTWrapper, error) {
	log.Printf("MQTT Host: %s; Port: %s; Username: %s", config.Host, config.Port, config.Username)

	u, err := url.Parse(fmt.Sprintf("mqtt://%s:%s", config.Host, config.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %s", err)
	}

	statusMessages := make(chan string, 1)

	cliCfg := autopaho.ClientConfig{
		ServerUrls:                    []*url.URL{u},
		KeepAlive:                     20,
		CleanStartOnInitialConnection: true,
		SessionExpiryInterval:         60,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			fmt.Println("mqtt connection up")

			if _, err := cm.Subscribe(config.Context, &paho.Subscribe{
				Subscriptions: []paho.SubscribeOptions{
					{
						Topic: "homeassistant/status",
						QoS:   byte(0),
					},
				},
			}); err != nil {
				log.Printf("failed to subscribe (%s). This is likely to mean homeassistant restarts will require manual addon restart.", err)
				return
			}

			log.Println("subscribed to homeassistant/status")
		},
		OnConnectError: func(err error) { log.Printf("error whilst attempting connection: %s\n", err) },
		ClientConfig: paho.ClientConfig{
			ClientID:      "pentairhome",
			OnClientError: func(err error) { fmt.Printf("client error: %s\n", err) },
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					msg := pr.Packet

					if msg.Topic == "homeassistant/status" {
						statusMessages <- string(msg.Payload)
					}

					return true, nil
				},
			},
		},
		ConnectUsername: config.Username,
		ConnectPassword: []byte(config.Password),
	}

	c, err := autopaho.NewConnection(config.Context, cliCfg)

	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %s", err)
	}

	if err = c.AwaitConnection(config.Context); err != nil {
		return nil, fmt.Errorf("failed to connect: %s", err)
	}

	fmt.Printf("Connected to %s\n", u)

	return &MQTTWrapper{
		Client:         c,
		Context:        config.Context,
		StatusMessages: statusMessages,
	}, nil
}
