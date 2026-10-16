package sync

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTPublisher implements Publisher using the Eclipse Paho MQTT client.
type MQTTPublisher struct {
	client  mqtt.Client
	opts    MQTTOptions
	qos     byte
}

// MQTTOptions configures the MQTT connection.
type MQTTOptions struct {
	Broker     string
	ClientID   string
	Username   string
	Password   string
	CACertPath string
	Insecure   bool
	QoS        byte // 0=at-most-once, 1=at-least-once, 2=exactly-once
}

// NewMQTTPublisher creates a new MQTTPublisher.
func NewMQTTPublisher(opts MQTTOptions) *MQTTPublisher {
	if opts.QoS == 0 {
		opts.QoS = 1 // default: at-least-once
	}

	clientOpts := mqtt.NewClientOptions().
		AddBroker(opts.Broker).
		SetClientID(opts.ClientID).
		SetCleanSession(false).            // durable session for QoS 1+
		SetAutoReconnect(true).            // paho handles reconnection
		SetMaxReconnectInterval(5 * time.Minute).
		SetConnectRetry(true).
		SetConnectRetryInterval(10 * time.Second).
		SetKeepAlive(60 * time.Second).
		SetPingTimeout(10 * time.Second).
		SetWriteTimeout(10 * time.Second).
		SetOrderMatters(false)

	if opts.Username != "" {
		clientOpts.SetUsername(opts.Username)
		clientOpts.SetPassword(opts.Password)
	}

	// TLS configuration
	if opts.CACertPath != "" || opts.Insecure {
		tlsCfg := &tls.Config{
			InsecureSkipVerify: opts.Insecure,
		}
		if opts.CACertPath != "" {
			caCert, err := os.ReadFile(opts.CACertPath)
			if err != nil {
				log.Printf("mqtt: failed to read CA cert %s: %v", opts.CACertPath, err)
			} else {
				pool := x509.NewCertPool()
				pool.AppendCertsFromPEM(caCert)
				tlsCfg.RootCAs = pool
			}
		}
		clientOpts.SetTLSConfig(tlsCfg)
	}

	clientOpts.SetOnConnectHandler(func(c mqtt.Client) {
		log.Printf("mqtt: connected to %s (client=%s)", opts.Broker, opts.ClientID)
	})
	clientOpts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		log.Printf("mqtt: connection lost: %v (will auto-reconnect)", err)
	})
	clientOpts.SetReconnectingHandler(func(c mqtt.Client, co *mqtt.ClientOptions) {
		log.Printf("mqtt: reconnecting to %s...", opts.Broker)
	})

	return &MQTTPublisher{
		client: mqtt.NewClient(clientOpts),
		opts:   opts,
		qos:    opts.QoS,
	}
}

// Connect implements Publisher.
func (p *MQTTPublisher) Connect() error {
	token := p.client.Connect()
	if !token.WaitTimeout(15 * time.Second) {
		return fmt.Errorf("mqtt connect timeout after 15s")
	}
	if token.Error() != nil {
		return fmt.Errorf("mqtt connect: %w", token.Error())
	}
	return nil
}

// Publish implements Publisher.
func (p *MQTTPublisher) Publish(topic string, payload []byte) error {
	token := p.client.Publish(topic, p.qos, false, payload)
	if !token.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("mqtt publish timeout on topic %s", topic)
	}
	if token.Error() != nil {
		return fmt.Errorf("mqtt publish %s: %w", topic, token.Error())
	}
	return nil
}

// IsConnected implements Publisher.
func (p *MQTTPublisher) IsConnected() bool {
	return p.client.IsConnected()
}

// Close implements Publisher.
func (p *MQTTPublisher) Close() error {
	p.client.Disconnect(1000) // 1s quiesce
	return nil
}

// Subscribe subscribes to a topic and calls the handler for each message.
// This is used for receiving commands from the cloud.
func (p *MQTTPublisher) Subscribe(topic string, handler func(topic string, payload []byte)) error {
	token := p.client.Subscribe(topic, p.qos, func(c mqtt.Client, msg mqtt.Message) {
		handler(msg.Topic(), msg.Payload())
	})
	if !token.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("mqtt subscribe timeout on topic %s", topic)
	}
	if token.Error() != nil {
		return fmt.Errorf("mqtt subscribe %s: %w", topic, token.Error())
	}
	log.Printf("mqtt: subscribed to %s", topic)
	return nil
}
