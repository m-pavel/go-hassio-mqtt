package main

import (
	ghm "github.com/m-pavel/go-hassio-mqtt/pkg"
	"github.com/m-pavel/go-hassio-mqtt/sample/sample"
)

func main() {
	ghm.NewExecutor[float64]("sample", &sample.RandomProducer{}, &ghm.HassioConsumer[float64]{Converter: sample.FloatConverter}).Main()
}
