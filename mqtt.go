package mqttGather

import (
	"log"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type Mqtt struct {
	Broker   string
	Topic    string
	ClientId string

	db     DB
	client MQTT.Client
}

func (m *Mqtt) Disconnect() error {
	m.client.Disconnect(1000)
	return nil
}

func (m *Mqtt) msgHandler(c MQTT.Client, msg MQTT.Message) {
	//fmt.Fprintf(os.Stdout, "%#v : %s -> %s\n", c, msg.Topic(), string(msg.Payload()))

	// /opennoise/c4:dd:57:66:95:60/dba_stats
	producer := msg.Topic()[11 : 11+17]

	csv := string(msg.Payload())
	stats, err := DBAStatsFromString(csv, producer)
	if err != nil {
		log.Printf("E: could not parse %s : %v", csv, err)
	} else {
		log.Printf("D: recv %s : %s", msg.Topic(), csv)
		if _, err := m.db.SaveNow(stats); err != nil {
			log.Printf("E: could not save %s (raw:%s) : %v", stats, csv, err)
		}
	}

}

func NewMQTT(cfg *RunConfig, db DB) (*Mqtt, error) {
	mqtt := Mqtt{
		Broker:   cfg.Host,
		Topic:    cfg.Topic,
		ClientId: cfg.ClientId,

		db: db,
	}

	opts := MQTT.NewClientOptions()
	opts.AddBroker(mqtt.Broker)
	opts.SetClientID(mqtt.ClientId)

	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		log.Printf("I: unexpected message on topic: %s : %s", msg.Topic(), string(msg.Payload()))
		// TODO log to db
	})

	mqtt.client = MQTT.NewClient(opts)
	token := mqtt.client.Connect()

	if token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	token = mqtt.client.Subscribe(mqtt.Topic, byte(0), mqtt.msgHandler)

	if token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &mqtt, nil
}
