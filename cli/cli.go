package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"

	"fmt"

	"crypto/tls"
	"crypto/x509"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func logger(c MQTT.Client, m MQTT.Message) {
	log.Println(fmt.Sprintf("[%s] %s", m.Topic(), m.Payload()))
}
func main() {

	MQTT.WARN = log.New(os.Stderr, "MQTT WARNING  ", log.Ltime|log.Lshortfile)
	MQTT.CRITICAL = log.New(os.Stderr, "MQTT CRITICAL ", log.Ltime|log.Lshortfile)
	MQTT.ERROR = log.New(os.Stderr, "MQTT ERROR    ", log.Ltime|log.Lshortfile)

	var mqtt = flag.String("mqtt", "tcp://localhost:1883", "MQTT endpoint")
	var user = flag.String("mqtt-user", "", "MQTT user")
	var pass = flag.String("mqtt-pass", "", "MQTT password")
	var mqttcliid = flag.String("mqtt-client", "", "Qoverwrite default MQTT client id")
	var mqttca = flag.String("mqtt-ca", "", "MQTT CA certificate")
	flag.Parse()
	opts := MQTT.NewClientOptions().AddBroker(*mqtt)
	if *mqttcliid != "" {
		opts.SetClientID(*mqttcliid)
	} else {
		opts.SetClientID("mqtt-go-client-5")
	}

	if *user != "" {
		opts.Username = *user
		opts.Password = *pass
	}
	if *mqttca != "" {
		tlscfg := tls.Config{}
		tlscfg.RootCAs = x509.NewCertPool()
		b, err := ioutil.ReadFile(*mqttca)
		if err != nil {
			log.Fatal(err)
		}
		ok := tlscfg.RootCAs.AppendCertsFromPEM(b)
		if !ok {
			log.Panicln("failed to parse root certificate")
		}

		opts.SetTLSConfig(&tlscfg)
	}
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Panicf("MQTT Connection error: %v\n", token.Error())
	}
	defer client.Disconnect(7)
	client.Subscribe("#", 0, logger)
	waitForCtrlC()

}

func waitForCtrlC() {
	var end_waiter sync.WaitGroup
	end_waiter.Add(1)
	var signal_channel chan os.Signal
	signal_channel = make(chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt)
	go func() {
		<-signal_channel
		end_waiter.Done()
	}()
	end_waiter.Wait()
}
