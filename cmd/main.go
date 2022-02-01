package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/a2800276/logrotation"

	"github.com/openaircgn/mqttGather"
)

var (
	version        string /* left for the linker to fill */
	sqliteDBName   = flag.String("sqlite", "", "connect string to use for sqlite, when in doubt: provide a filename")
	topic          = flag.String("topic", "", "topic to subscribe to") // todo, this should later be a plugin for sensors
	telemetryTopic = flag.String("telemetry-topic", "", "topic to subscribe to for telemetry data")
	host           = flag.String("host", "", "host to connect to")
	clientId       = flag.String("clientID", "", "clientId to use for connection")
	silent         = flag.Bool("silent", false, "psssh!")
	config         = flag.String("c", "", "name of (optional) config file")
	logDir         = flag.String("log-dir", "", "where to write logs, writes to stdout if not set")
	smsKey         = flag.String("sms-key", "", "api key for SMS")
	_version       = flag.Bool("version", false, "display version information and exit")
)

func banner(w io.Writer) {
	fmt.Fprintf(w, "%s ver %s\n", os.Args[0], version)
}
func summary(rc mqttGather.RunConfig, w io.Writer) {
	banner(w)
	fmt.Fprintf(w, "sqlite connect: %s\n", rc.SqlLiteConnect)
	fmt.Fprintf(w, "subscribing to: %s\n", rc.Topic)
	fmt.Fprintf(w, "host          : %s\n", rc.Host)
	fmt.Fprintf(w, "clientId      : %s\n", rc.ClientId)
	fmt.Fprintf(w, "logDir        : %s\n", rc.LogDir)
	if rc.SMSKey != "" {
		fmt.Fprintf(w, "smsKey        : %s\n", "***")
	} else {
		fmt.Fprintf(w, "smsKey        : %s\n", "not set!")
	}

	if *config != "" {
		fmt.Fprintf(w, "config file   : %s\n", *config)
	}
}

func startAlert(cfg *mqttGather.RunConfig, mqtt *mqttGather.Mqtt) {
	done := make(chan bool)
	alert := mqttGather.NewAlerter(cfg, mqtt, done)
	alert.Start()
}

func main() {
	keepAlive := make(chan os.Signal)
	signal.Notify(keepAlive, os.Interrupt, syscall.SIGTERM)

	flag.Parse()

	if *_version {
		banner(os.Stderr)
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
		summary(*rc, os.Stderr)
	}

	if *logDir != "" {
		rc.LogDir = *logDir
	}

	if *smsKey != "" {
		rc.SMSKey = *smsKey
	}

	var logWriter io.Writer

	if rc.LogDir != "" {
		logWriter = &logrotation.Logrotation{
			BaseFilename: "opennoise",
			Suffix:       "log",
			BaseDir:      rc.LogDir,
			Interval:     24 * time.Hour,
		}
	} else {
		logWriter = os.Stdout
	}
	log.SetOutput(logWriter)

	summary(*rc, logWriter)

	// Start Collecting
	mqtt, err := mqttGather.NewMQTT(rc)
	if err != nil {
		panic(err)
	}
	defer mqtt.Disconnect()

	// start alerting

	if rc.SMSKey != "" {
		startAlert(rc, mqtt)
	}

	<-keepAlive
}
