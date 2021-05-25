package mqttGather

import (
	"fmt"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type Mqtt struct {
	Broker   string
	ClientId string
	Topic    string

	db     DB
	client MQTT.Client
}

func (m *Mqtt) Disconnect() error {
	m.client.Disconnect(1000)
	return nil
}

func (m *Mqtt) msgHandler(c MQTT.Client, msg MQTT.Message) {
	fmt.Fprintf(os.Stdout, "%#v : %s -> %s\n", c, msg.Topic(), string(msg.Payload()))

	// /opennoise/c4:dd:57:66:95:60/dba_stats
	client := msg.Topic()[11 : 11+17]
	stats, err := DBAStatsFromString(string(msg.Payload()), client)
	if err != nil {
		println("here1")
		fmt.Printf("%v", err)
	} else {
		if _, err := m.db.SaveNow(stats); err != nil {
			println("here2")
			println(err)
		}
	}

}

func NewMQTT(db DB) (*Mqtt, error) {
	mqtt := Mqtt{
		"tcp://test.mosquitto.org:1883",
		"mqttGather",
		"/opennoise/#",
		db,
		nil,
	}

	opts := MQTT.NewClientOptions()
	opts.AddBroker(mqtt.Broker)
	opts.SetClientID(mqtt.ClientId)

	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		fmt.Printf("unexpected: %s -> %s", msg.Topic(), string(msg.Payload()))
	})

	mqtt.client = MQTT.NewClient(opts)
	if token := mqtt.client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	if token := mqtt.client.Subscribe(mqtt.Topic, byte(0), mqtt.msgHandler); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &mqtt, nil
}
