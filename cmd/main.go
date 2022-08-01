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
	ond "github.com/presseverykey/opennoise_daemon"
)

var (
	version        string /* left for the linker to fill */
	sqliteDBName   = flag.String("sqlite", "", "connect string to use for sqlite, when in doubt: provide a filename")
	noiseTopic     = flag.String("noise-topic", "", "topic to subscribe to for noise data")
	telemetryTopic = flag.String("telemetry-topic", "", "topic to subscribe to for telemetry data")
	mqttHost       = flag.String("mqtt-host", "", "host to connect to")
	mqttClientId   = flag.String("mqtt-client-id", "", "clientId to use for connection")
	silent         = flag.Bool("silent", false, "psssh!")
	config         = flag.String("c", "", "name of (optional) config file")
	logDir         = flag.String("log-dir", "", "where to write logs, writes to stdout if not set")
	smsKey         = flag.String("sms-key", "", "api key for SMS")
	_version       = flag.Bool("version", false, "display version information and exit")
)

func banner(w io.Writer) {
	fmt.Fprintf(w, "%s ver %s\n", os.Args[0], version)
}

func summary(rc ond.RunConfig, w io.Writer) {
	banner(w)
	fmt.Fprintf(w, "sqlite connect: %s\n", rc.SqlLiteConnect)
	fmt.Fprintf(w, "subscribing to: %s\n", rc.MqttNoiseTopic)
	fmt.Fprintf(w, "host          : %s\n", rc.MqttHost)
	fmt.Fprintf(w, "clientId      : %s\n", rc.MqttClientId)
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

func parseConfig() (*(ond.RunConfig), error) {
	flag.Parse()

	if *_version {
		banner(os.Stderr)
		os.Exit(0)
	}

	var rc *ond.RunConfig
	var err error

	if *config != "" {	// config given: load from file
		rc, err = ond.LoadFromFile(*config);
		if err != nil {
			return rc, err
		}
	} else {	//no config given: use individual args
		rc = &ond.RunConfig{}
	}
	if *sqliteDBName != "" {
		rc.SqlLiteConnect = *sqliteDBName
	}
	if *mqttHost != "" {
		rc.MqttHost = *mqttHost
	}
	if *noiseTopic != "" {
		rc.MqttNoiseTopic = *noiseTopic
	}
	if *telemetryTopic != "" {
		rc.MqttTelemetryTopic = *telemetryTopic
	}
	if *mqttClientId != "" {
		rc.MqttClientId = *mqttClientId
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
	return rc, rc.Check()
}

func main() {
	keepAlive := make(chan os.Signal)
	signal.Notify(keepAlive, os.Interrupt, syscall.SIGTERM)

	//parse args and/or config file
	rc, err := parseConfig()
	if err != nil {
		panic(err)
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

	// Build database connection
	db, err := ond.NewDatabase(rc.SqlLiteConnect)
	if err != nil {
		panic(err)
	}

	// Build mqtt connection
	mqtt, err := ond.NewMQTT(rc)
	if err != nil {
		panic(err)
	}
	defer mqtt.Disconnect()

	// Build SMS notifier
	notifier, err := ond.NewSMSNotifier(rc.SMSKey)
	if err != nil {
		panic(err)
	}

	// Build data gather handler
	gather, err := ond.NewNoiseGather(db)
	if err != nil {
		panic(err)
	}

	// build alert handler
	alerter, err := ond.NewNoiseAlert(db, notifier)
	if err != nil {
		panic(err)
	}

	// start alerting
	mqtt.Register(gather)
	mqtt.Register(alerter)

	<-keepAlive
}
