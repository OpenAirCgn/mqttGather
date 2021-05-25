package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/openaircgn/mqttGather"
)

func main() {
	keepAlive := make(chan os.Signal)
	signal.Notify(keepAlive, os.Interrupt, syscall.SIGTERM)

	db, err := mqttGather.NewDatabase("prod.sqlite3")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	mqtt, err := mqttGather.NewMQTT(db)
	if err != nil {
		panic(err)
	}
	defer mqtt.Disconnect()

	<-keepAlive
}
