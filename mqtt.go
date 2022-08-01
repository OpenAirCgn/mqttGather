package opennoise_daemon

// MQTT connection handler. Connects to MQTT, 
// subscribes to given topics, responds to messages

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/url"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type Mqtt struct {
	cfg          *RunConfig
	client       MQTT.Client
	observerList []DeviceObserver

}

func (m *Mqtt) Register(o DeviceObserver) {
	m.observerList = append(m.observerList, o);
}

func (m *Mqtt) Unregister(o DeviceObserver) {
	var newObserverList []DeviceObserver	
	for _, observer := range m.observerList {
		if observer != o {
			newObserverList = append(m.observerList, observer)
		}
	}
	m.observerList = newObserverList
}

func (m *Mqtt) noiseMsgHandler(c MQTT.Client, msg MQTT.Message) {
	// /opennoise/c4:dd:57:66:95:60/dba_stats
	producer := retrieveClientId(msg.Topic())
	payload := string(msg.Payload())
	stats, err := DBAStatsFromString(payload, producer)
	if err != nil {
		log.Printf("E: Could not parse %s : %v", payload, err)
	} else {
		log.Printf("D: recv %s : %s", msg.Topic(), payload)
		for _, observer := range m.observerList {
      observer.SensorDBAStats(stats)
    }
	}
}

func (m *Mqtt) telemetryMsgHandler(c MQTT.Client, msg MQTT.Message) {
	// /opennoise/c4:dd:57:66:95:60/telemetry
	producer := retrieveClientId(msg.Topic())
	payload := string(msg.Payload())
	telemetry, err := TelemetryFromPayload(payload, producer)
	if err != nil {
		log.Printf("E: Could not parse %s : %v", payload, err)
	} else {
		log.Printf("D: recv %s : %s", msg.Topic(), payload)
		for _, observer := range m.observerList {
      observer.SensorTelemetry(telemetry)
    }
	}
}

func (m *Mqtt) Disconnect() error {
	m.client.Disconnect(1000)
	return nil
}

func retrieveClientId(topic string) string {
	// /opennoise/c4:dd:57:66:95:60/dba_stats
	// producer := msg.Topic()[11 : 11+17]
	if len(topic) < 28 {
		log.Printf("E: invalid topic: %s", topic)
		return fmt.Sprintf("?:%s", topic)
	}
	return topic[11 : 11+17]
}

func NewMQTT(cfg *RunConfig) (*Mqtt, error) {
	mqtt := Mqtt {
		cfg: cfg,
	}

	opts := MQTT.NewClientOptions()
	opts.AddBroker(cfg.MqttHost)
	opts.SetClientID(cfg.MqttClientId)

	opts.SetConnectRetryInterval(10 * time.Second)
	opts.SetConnectionAttemptHandler(func(u *url.URL, cfg *tls.Config) *tls.Config {
		log.Printf("D: connection attempt: %v", u)
		return cfg // why!?
	})
	opts.SetConnectionLostHandler(func(c MQTT.Client, err error) {
		log.Printf("E: connection lost: %v", err)
	})
	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		log.Printf("I: unexpected message on topic: %s : %s", msg.Topic(), string(msg.Payload()))
		// TODO log to db
	})
	opts.SetOnConnectHandler(func(c MQTT.Client) {
		opts := c.OptionsReader()
		log.Printf("D: connect: %v", opts.ClientID())

		// Noise Topic
		token := mqtt.client.Subscribe(cfg.MqttNoiseTopic, byte(0), mqtt.noiseMsgHandler)
		if token.Wait() && token.Error() != nil {
			log.Printf("E: subscribtion failed: %s (%v)", cfg.MqttNoiseTopic, token.Error())
			// TODO figure out what to do here: sit under a tree crying and waiting to die?
		} else {
			log.Printf("D: subscribed to: %s", cfg.MqttNoiseTopic)
		}

		// Telemetry Topic
		token = mqtt.client.Subscribe(cfg.MqttTelemetryTopic, byte(0), mqtt.telemetryMsgHandler)
		if token.Wait() && token.Error() != nil {
			log.Printf("E: subscription failed: %s (%v)", cfg.MqttTelemetryTopic, token.Error())
		} else {
			log.Printf("D: subscribed to: %s", cfg.MqttTelemetryTopic)
		}
	})
	opts.SetReconnectingHandler(func(c MQTT.Client, o *MQTT.ClientOptions) {
		opts := c.OptionsReader()
		time.Sleep(2 * time.Second) // don't just hammer away at poor server.
		log.Printf("I: reconnecting: %v", opts.ClientID())
	})

	mqtt.client = MQTT.NewClient(opts)

	token := mqtt.client.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &mqtt, nil
}
