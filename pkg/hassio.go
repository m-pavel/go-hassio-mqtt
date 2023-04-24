package ghm

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
)

const (
	online  = "online"
	offline = "offline"
	timeout = time.Second * 5
)

type HassioConsumer[R any] struct {
	host     string
	topic    string
	topicc   string
	topica   string
	user     string
	password string
	cliId    string
	ca       string //file
	name     string

	client MQTT.Client

	Converter func(R) any
}

func (hc *HassioConsumer[R]) Setup(cmd *cobra.Command, name string) {
	// cmd.PersistentFlags().IntVar(&mc.period, "period", 60, "Period minutes")

	cmd.PersistentFlags().StringVar(&hc.host, "mqtt", "tcp://localhost:1883", "MQTT endpoint")
	cmd.PersistentFlags().StringVar(&hc.host, "t", fmt.Sprintf("nn/%s", name), "MQTT topic")
	cmd.PersistentFlags().StringVar(&hc.host, "tc", fmt.Sprintf("nn/%s-control", name), "MQTT control topic")
	cmd.PersistentFlags().StringVar(&hc.host, "ta", fmt.Sprintf("nn/%s-aval", name), "MQTT availability topic")
	cmd.PersistentFlags().StringVar(&hc.host, "mqtt-user", "", "MQTT user")
	cmd.PersistentFlags().StringVar(&hc.host, "mqtt-pass", "", "MQTT password")
	cmd.PersistentFlags().StringVar(&hc.host, "mqtt-client", "", "Overwrite default MQTT client id")
	cmd.PersistentFlags().StringVar(&hc.host, "mqtt-ca", "", "MQTT CA certificate file")
	hc.name = name
}
func (hc *HassioConsumer[R]) Init(debug bool) error {
	if debug {
		MQTT.DEBUG = log.New(os.Stderr, "MQTT DEBUG    ", log.Ltime|log.Lshortfile)
	}
	MQTT.WARN = log.New(os.Stderr, "MQTT WARNING  ", log.Ltime|log.Lshortfile)
	MQTT.CRITICAL = log.New(os.Stderr, "MQTT CRITICAL ", log.Ltime|log.Lshortfile)
	MQTT.ERROR = log.New(os.Stderr, "MQTT ERROR    ", log.Ltime|log.Lshortfile)

	if hc.cliId == "" {
		hc.cliId = fmt.Sprintf("%s-go-cli", hc.name)
	}

	if err := hc.setupMqtt(); err != nil {
		log.Panicf("MQTT Connection error: %v\n", err)
	}
	if hc.Converter == nil {
		hc.Converter = func(r R) any {
			return r
		}
	}
	return nil
}
func (hc *HassioConsumer[R]) Consume(v R) error {

	jpl, err := json.Marshal(hc.Converter(v))
	if err != nil {
		log.Println(err)
	} else {
		// if hmss.trace {
		// 	log.Printf("MQTT Payload: %s\n", jpl)
		// }
		if token := hc.client.Publish(hc.topic, 1, false, jpl); token.WaitTimeout(timeout) && token.Error() != nil {
			log.Println(token.Error())
		}
		if token := hc.client.Publish(hc.topica, 0, false, online); token.WaitTimeout(timeout) && token.Error() != nil {
			log.Println(token.Error())
		}
	}

	return nil
}

func (hc *HassioConsumer[R]) Close() error {
	hc.client.Disconnect(3000)
	return nil
}

func (hc *HassioConsumer[R]) setupMqtt() error {
	// Open MQTT connection
	opts := MQTT.NewClientOptions().AddBroker(hc.host)

	opts.SetClientID(hc.cliId)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.OnConnect = func(c MQTT.Client) {
		// TODO
		//	hc.s.OnConnect(c, hc.topic, hc.topicc, hc.topica)
	}

	if hc.user != "" {
		opts.Username = hc.user
		opts.Password = hc.password
	}

	if hc.ca != "" {
		tlscfg := tls.Config{}
		tlscfg.RootCAs = x509.NewCertPool()
		var b []byte
		var err error
		if b, err = os.ReadFile(hc.ca); err != nil {
			return err
		}
		if ok := tlscfg.RootCAs.AppendCertsFromPEM(b); !ok {
			return errors.New("failed to parse root certificate")
		}
		opts.SetTLSConfig(&tlscfg)
	}

	opts.WillEnabled = true
	opts.WillPayload = []byte(offline)
	opts.WillTopic = hc.topica
	opts.WillRetained = true

	hc.client = MQTT.NewClient(opts)
	if token := hc.client.Connect(); token.WaitTimeout(timeout) && token.Error() != nil {
		return token.Error()
	}
	log.Printf("MQTT Connected to %s. Topic is '%s'. Control topic is '%s'. Availability topic is '%s'\n", hc.host, hc.topic, hc.topicc, hc.topica)
	return nil
}
