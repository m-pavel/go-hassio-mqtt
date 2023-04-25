package main

import (
	ghm "github.com/m-pavel/go-hassio-mqtt/pkg"
	"github.com/m-pavel/go-hassio-mqtt/sample/sample"
)

func main() {
	ghm.NewExecutor[float64]("sample", &sample.RandomProducer{}, &ghm.HttpServer[float64]{ToJsonConverter: sample.FloatConverter}).Main()
}
