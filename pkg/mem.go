package ghm

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

type MemoryConsumer[R any] struct {
	period int

	debug bool
	cache map[int64]R
	lock  *sync.Mutex
	last  R

	cleanup context.CancelFunc
}

func (mc *MemoryConsumer[R]) Setup(cmd *cobra.Command, name string) {
	cmd.PersistentFlags().IntVar(&mc.period, "period", 60, "Period minutes")
}

func (mc *MemoryConsumer[R]) Init(d bool) error {
	mc.cache = map[int64]R{}
	mc.debug = d
	mc.lock = &sync.Mutex{}

	var ctx context.Context
	ctx, mc.cleanup = context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(time.Minute * time.Duration(mc.period) / 4)
	fl:
		for {
			select {
			case <-ctx.Done():
				break fl
			case <-ticker.C:
				mc.clearCache()
			}
		}
		ticker.Stop()
	}()
	return nil
}
func (mc *MemoryConsumer[R]) Consume(v R) error {
	mc.last = v
	mc.lock.Lock()
	defer mc.lock.Unlock()
	mc.cache[time.Now().Unix()] = v
	if mc.debug {
		log.Printf("Consumed %v", v)
	}
	return nil
}

func (mc *MemoryConsumer[R]) Close() error { return nil }

func (mc *MemoryConsumer[R]) clearCache() {
	mc.lock.Lock()
	defer mc.lock.Unlock()
	dl := time.Now().Add(-time.Minute * time.Duration(mc.period)).Unix()
	for k := range mc.cache {
		if k < dl {
			delete(mc.cache, k)
		}
	}
}

func (mc MemoryConsumer[R]) Last() R {
	return mc.last
}
