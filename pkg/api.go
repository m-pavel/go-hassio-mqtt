package ghm

import "github.com/spf13/cobra"

// Executor to be obtained via NewExecutor
type Executor interface {
	Main()
}

// Consumer is generic consumer
type Consumer[R any] interface {
	Setup(*cobra.Command, string)
	Init(bool) error
	Consume(v R) error
	Close() error
}

// Producer is a generic producer
type Producer[R any] interface {
	Setup(*cobra.Command, string)
	Init(bool) error
	Produce() (R, error)
	Close() error
}
