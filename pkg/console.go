package ghm

import (
	"log"

	"github.com/spf13/cobra"
)

type ConsoleConsumer[R any] struct {
	Converter func(R) any
}

func (cc *ConsoleConsumer[R]) Setup(cmd *cobra.Command, name string) {
}

func (cc *ConsoleConsumer[R]) Init(d bool) error {
	if cc.Converter == nil {
		cc.Converter = func(r R) any { return r }
	}
	return nil
}
func (cc *ConsoleConsumer[R]) Consume(v R) error {
	log.Println(cc.Converter(v))
	return nil
}

func (cc *ConsoleConsumer[R]) Close() error { return nil }
