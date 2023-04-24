package sample

import (
	"log"
	"math/rand"
	"time"

	"github.com/spf13/cobra"
)

type Envelope struct {
	Value float64 `json:"value"`
}

func FloatConverter(v float64) any {
	return &Envelope{Value: v}
}

type RandomProducer struct {
	debug bool
}

func (rp *RandomProducer) Setup(*cobra.Command, string) {}
func (rp *RandomProducer) Init(d bool) error {
	rand.Seed(time.Now().Unix())
	rp.debug = d
	return nil
}
func (rp *RandomProducer) Produce() (float64, error) {
	if rp.debug {
		log.Println("Generated random value")
	}
	return rand.Float64(), nil
}
func (rp *RandomProducer) Close() error { return nil }
