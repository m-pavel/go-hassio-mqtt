package main

import (
	ghm "github.com/m-pavel/go-hassio-mqtt/pkg"
	"github.com/m-pavel/go-hassio-mqtt/sample/sample"
)

func ToRaw(v *sample.Envelope) ghm.Entry {
	return ghm.Entry{"value": v.Value, "const": 3.14}
}

func main() {
	ghm.NewExecutor[*sample.Envelope]("sample", &sample.RandomStructProducer{}, &ghm.HttpServer[*sample.Envelope]{
		ToRawConverter: ToRaw,
		YAxis:          []string{"value", "const"},
	}).Main()
}
