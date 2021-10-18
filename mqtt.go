package mqttGather

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/url"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type Mqtt struct {
	Broker         string
	Topic          string
	TelemetryTopic string
	ClientId       string

	cfg          *RunConfig
	db           DB
	client       MQTT.Client
	statsChannel chan DBAStats
}

func (m *Mqtt) Disconnect() error {
	m.db.Close()
	m.client.Disconnect(1000)
	close(m.statsChannel)
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
func (m *Mqtt) msgHandler(c MQTT.Client, msg MQTT.Message) {
	//fmt.Fprintf(os.Stdout, "%#v : %s -> %s\n", c, msg.Topic(), string(msg.Payload()))

	// /opennoise/c4:dd:57:66:95:60/dba_stats
	producer := retrieveClientId(msg.Topic())

	csv := string(msg.Payload())
	stats, err := DBAStatsFromString(csv, producer)
	if err != nil {
		log.Printf("E: could not parse %s : %v", csv, err)
	} else {
		log.Printf("D: recv %s : %s", msg.Topic(), csv)
		if _, err := m.db.SaveNow(stats); err != nil {
			log.Printf("E: could not save %s (raw:%s) : %v", stats, csv, err)
			if err := m.connectDB(); err != nil {
				log.Printf("E: could not reconnect to db (%v)", err)
			}
		}
		m.statsChannel <- *stats
	}

}

func (m *Mqtt) msgHandlerTelemetry(c MQTT.Client, msg MQTT.Message) {
	//fmt.Fprintf(os.Stdout, "%#v : %s -> %s\n", c, msg.Topic(), string(msg.Payload()))

	// /opennoise/c4:dd:57:66:95:60/telemetry
	producer := retrieveClientId(msg.Topic())

	payload := string(msg.Payload())
	telemetry, err := TelemetryFromPayload(payload, producer)
	if err != nil {
		log.Printf("E: could not parse %s : %v", payload, err)
	} else {
		log.Printf("D: recv %s : %s", msg.Topic(), payload)
		if _, err := m.db.SaveTelemetryNow(telemetry); err != nil {
			log.Printf("E: could not save %s (raw:%s) : %v", telemetry, payload, err)
			if err := m.connectDB(); err != nil {
				log.Printf("E: could not reconnect to db (%v)", err)
			}
		}
	}

}

func (m *Mqtt) connectDB() error {
	if m.db != nil {
		log.Printf("D: closing existing connection")
		m.db.Close()
	}

	if db, err := NewDatabase(m.cfg.SqlLiteConnect); err != nil {
		return err
	} else {
		m.db = db
	}
	return nil
}

func NewMQTT(cfg *RunConfig) (*Mqtt, error) {
	mqtt := Mqtt{
		Broker:         cfg.Host,
		Topic:          cfg.Topic,
		TelemetryTopic: cfg.TelemetryTopic,
		ClientId:       cfg.ClientId,

		cfg: cfg,
	}

	if err := mqtt.connectDB(); err != nil {
		log.Printf("E: could not connect to DB: %v", err)
		return nil, err
	}

	mqtt.statsChannel = make(chan DBAStats)

	opts := MQTT.NewClientOptions()
	opts.AddBroker(mqtt.Broker)
	opts.SetClientID(mqtt.ClientId)

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
		token := mqtt.client.Subscribe(mqtt.Topic, byte(0), mqtt.msgHandler)

		// Stats Topic
		if token.Wait() && token.Error() != nil {
			log.Printf("E: subscribtion failed: %s (%v)", mqtt.Topic, token.Error())
			// TODO figure out what to do here: sit under a tree crying and waiting to die?
		} else {
			log.Printf("D: subscribed to: %s", mqtt.Topic)
		}

		// Telemetry Topic
		token = mqtt.client.Subscribe(mqtt.TelemetryTopic, byte(0), mqtt.msgHandlerTelemetry)
		if token.Wait() && token.Error() != nil {
			log.Printf("E: subscribtion failed: %s (%v)", mqtt.TelemetryTopic, token.Error())
		} else {
			log.Printf("D: subscribed to: %s", mqtt.TelemetryTopic)
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
