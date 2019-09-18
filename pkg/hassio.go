package ghm

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/sevlyar/go-daemon"
)

type HassioMqttService interface {
	PrepareCommandLineParams()
	Name() string
	Init(client MQTT.Client, topic, topicc, topica string, debug bool, ss SendState) error
	Do() (interface{}, error)
	Close() error
}

type HassioMqttServiceStub struct {
	s    HassioMqttService
	stop chan struct{}
	done chan struct{}

	//client MQTT.Client
	topic  string
	topica string
	topicc string
}

type SendState func() error

func NewStub(s HassioMqttService) *HassioMqttServiceStub {
	hms := HassioMqttServiceStub{s: s}
	hms.done = make(chan struct{})
	hms.stop = make(chan struct{})
	return &hms
}

func (hmss *HassioMqttServiceStub) sendState() error {
	return nil
}
func (hmss *HassioMqttServiceStub) Main() {
	hmss.s.PrepareCommandLineParams()
	name := hmss.s.Name()
	var logf = flag.String("log", fmt.Sprintf("%s.log", name), "log")
	var pid = flag.String("pid", fmt.Sprintf("%s.pid", name), "pid")
	var notdaemonize = flag.Bool("n", false, "Do not do to background.")
	var signal = flag.String("s", "", `send signal to the daemon stop — shutdown`)
	var mqtt = flag.String("mqtt", "tcp://localhost:1883", "MQTT endpoint")
	var topic = flag.String("t", fmt.Sprintf("nn/%s", name), "MQTT topic")
	var topicc = flag.String("tc", fmt.Sprintf("nn/%s-control", name), "MQTT control topic")
	var topica = flag.String("ta", fmt.Sprintf("nn/%s-aval", name), "MQTT availability topic")
	var user = flag.String("mqtt-user", "", "MQTT user")
	var pass = flag.String("mqtt-pass", "", "MQTT password")
	var mqttcliid = flag.String("mqtt-client", "", "Qoverwrite default MQTT client id")
	var debug = flag.Bool("d", false, "debug")
	var interval = flag.Int("interval", 10, "Interval secons")
	var failcnt = flag.Int("failcnt", 15, "Fail after n errors")
	flag.Parse()
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, hmss.termHandler)
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)

	cntxt := &daemon.Context{
		PidFileName: *pid,
		PidFilePerm: 0644,
		LogFileName: *logf,
		LogFilePerm: 0640,
		WorkDir:     "/tmp",
		Umask:       027,
		Args:        os.Args,
	}

	// Send signal if passed
	if !*notdaemonize && len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			log.Fatalf("Unable send signal to the daemon: %v", err)
		}
		daemon.SendCommands(d)
		return
	}

	// Daemonize
	if !*notdaemonize {
		d, err := cntxt.Reborn()
		if err != nil {
			log.Fatal(err)
		}
		if d != nil {
			return
		}
	}

	// Open MQTT connection
	opts := MQTT.NewClientOptions().AddBroker(*mqtt)
	if *mqttcliid != "" {
		opts.SetClientID(*mqttcliid)
	} else {
		opts.SetClientID(fmt.Sprintf("%s-go-cli", name))
	}
	//opts.SetKeepAlive(time.Duration(5) * time.Second)
	//opts.SetAutoReconnect(true)
	if *user != "" {
		opts.Username = *user
		opts.Password = *pass
	}

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Panicf("MQTT Connection error: %v\n", token.Error())
	}
	log.Printf("MQTT Connected to %s. Topic is '%s'. Control topic is '%s'. Availability topic is '%s'\n", *mqtt, *topic, *topicc, *topica)

	err := hmss.s.Init(client, *topic, *topicc, *topica, *debug, hmss.sendState)
	if err != nil {
		log.Panicf("Service init error: %v\n", err)
	}
	actfail := 0

	log.Printf("Starting main loop with %d s. interval.\n", *interval)
	for {
		select {
		case <-hmss.stop:
			log.Println("Exiting because of signal.")
			break
		case <-time.After(time.Duration(*interval) * time.Second):
			if *failcnt > 0 && actfail >= *failcnt {
				log.Printf("Fail limit reached (%d). Exiting.\n", actfail)
				return
			}
			//err := hmss.sendState()
			//if err == nil {
			//	actfail = 0
			//} else {
			//	actfail++
			//}
			v, err := hmss.s.Do()
			if err != nil {
				if token := client.Publish(hmss.topica, 0, false, "offline");  token.Error() != nil {
					log.Println(token.Error())
				}
			} else {
				jpl, err := json.Marshal(&v)
				if err != nil {
					log.Println(err)
				} else {
					if token := client.Publish(hmss.topic, 0, false, jpl); token.Wait() && token.Error() != nil {
						log.Println(token.Error())
					}
					if token := client.Publish(hmss.topica, 0, false, "online");  token.Error() != nil {
						log.Println(token.Error())
					}
				}
			}
			//return err
		}
	}

	if err := hmss.s.Close(); err != nil {
		log.Println(err)
	}
	log.Println("Disconnecting")
	client.Disconnect(3000)

	hmss.done <- struct{}{}
}

func (hmss HassioMqttServiceStub) termHandler(sig os.Signal) error {
	log.Println("terminating...")
	hmss.stop <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-hmss.done
	}
	return daemon.ErrStop
}
