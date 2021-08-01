package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/openaircgn/mqttGather"
)

var (
	version        string /* left for the linker to fill */
	sqliteDBName   = flag.String("sqlite", "", "connect string to use for sqlite, when in doubt: provide a filename")
	topic          = flag.String("topic", "", "topic to subscribe to") // todo, this should later be a plugin for sensors
	telemetryTopic = flag.String("telemetry-topic", "", "topic to subscribe to")
	host           = flag.String("host", "", "host to connect to")
	clientId       = flag.String("clientID", "", "clientId to use for connection")
	silent         = flag.Bool("silent", false, "psssh!")
	config         = flag.String("c", "", "name of (optional) config file")
	_version       = flag.Bool("version", false, "display version information and exit")
)

func banner() {
	fmt.Fprintf(os.Stderr, "%s ver %s\n", os.Args[0], version)
}
func summary(rc mqttGather.RunConfig) {
	banner()
	fmt.Fprintf(os.Stderr, "sqlite connect: %s\n", rc.SqlLiteConnect)
	fmt.Fprintf(os.Stderr, "subscribing to: %s\n", rc.Topic)
	fmt.Fprintf(os.Stderr, "host          : %s\n", rc.Host)
	fmt.Fprintf(os.Stderr, "clientId      : %s\n", rc.ClientId)
	if *config != "" {
		fmt.Fprintf(os.Stderr, "config file   : %s\n", *config)

	}

}
func main() {
	keepAlive := make(chan os.Signal)
	signal.Notify(keepAlive, os.Interrupt, syscall.SIGTERM)

	flag.Parse()

	if *_version {
		banner()
		os.Exit(0)
	}

	var rc *mqttGather.RunConfig
	var err error
	if *config != "" {
		if rc, err = mqttGather.LoadFromFile(*config); err != nil {
			panic(err)
		}

	} else {
		rc = &mqttGather.RunConfig{}
	}
	if *sqliteDBName != "" {
		rc.SqlLiteConnect = *sqliteDBName
	}
	if *host != "" {
		rc.Host = *host
	}
	if *topic != "" {
		rc.Topic = *topic
	}
	if *telemetryTopic != "" {
		rc.TelemetryTopic = *telemetryTopic
	}
	if *clientId != "" {
		rc.ClientId = *clientId
	}

	if !*silent {
		summary(*rc)
	}

	mqtt, err := mqttGather.NewMQTT(rc)

	if err != nil {
		panic(err)
	}
	defer mqtt.Disconnect()

	<-keepAlive
}
