package ghm

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/sevlyar/go-daemon"
	"github.com/spf13/cobra"
)

type executor[R any] struct {
	rootCmd *cobra.Command
	c       []Consumer[R]
	p       Producer[R]

	pid          string
	log          string
	name         string
	notdaemonize bool
	daemonWd     string
	cancel       func()
	debug        bool

	interval      int
	fail          int
	failOnConsume bool
}

func (ex *executor[R]) Main() {
	if err := ex.rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func NewExecutor[R any](name string, p Producer[R], c ...Consumer[R]) Executor {
	e := executor[R]{p: p, c: c, name: name}

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Hugo",
		Long:  `All software has versions. This is Hugo's`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hugo Static Site Generator v0.9 -- HEAD")
		},
	}

	var stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop daemon",
		RunE:  e.stop,
	}

	e.rootCmd = &cobra.Command{
		Short: "hass.io mqtt client",
		RunE:  e.execute,
	}
	e.rootCmd.AddCommand(versionCmd)
	e.rootCmd.AddCommand(stopCmd)
	e.rootCmd.PersistentFlags().StringVar(&e.log, "log", fmt.Sprintf("%s.log", name), "Log file")
	e.rootCmd.PersistentFlags().StringVar(&e.pid, "pid", fmt.Sprintf("%s.pid", name), "Pid file")
	e.rootCmd.PersistentFlags().BoolVarP(&e.notdaemonize, "not-daemon", "n", false, "Do not go to background")
	e.rootCmd.PersistentFlags().StringVar(&e.daemonWd, "daemon-workdir", "/tmp", "Daemon work dir")
	e.rootCmd.PersistentFlags().IntVar(&e.interval, "interval", 15, "Ask interval seconds")
	e.rootCmd.PersistentFlags().IntVar(&e.fail, "fail", 10, "Fail readings count to fail")
	e.rootCmd.PersistentFlags().BoolVar(&e.failOnConsume, "fail-on-consume", false, "Fail on consuming errors")
	e.rootCmd.PersistentFlags().BoolVarP(&e.debug, "debug", "d", false, "Debug")

	p.Setup(e.rootCmd, name)
	for _, cs := range c {
		cs.Setup(e.rootCmd, name)
	}
	return &e
}

func (ex *executor[R]) daemonContext() *daemon.Context {
	return &daemon.Context{
		PidFileName: ex.pid,
		PidFilePerm: 0644,
		LogFileName: ex.log,
		LogFilePerm: 0640,
		WorkDir:     ex.daemonWd,
		Umask:       027,
		Args:        os.Args,
	}
}
func (ex *executor[R]) execute(cmd *cobra.Command, args []string) error {
	if err := ex.p.Init(ex.debug); err != nil {
		return err
	}
	for _, cs := range ex.c {
		if err := cs.Init(ex.debug); err != nil {
			return err
		}
	}

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	go ex.termHandler(ch)

	// Daemonize
	if !ex.notdaemonize {
		cntxt := ex.daemonContext()
		d, err := cntxt.Reborn()
		if err != nil {
			log.Fatal(err)
		}
		if d != nil {
			log.Printf("Created daemon process %d", d.Pid)
			return nil
		}
	}

	var ctx context.Context
	ctx, ex.cancel = context.WithCancel(context.Background())

	ex.loop(ctx)

	if err := ex.p.Close(); err != nil {
		log.Println(err)
	}
	for _, cs := range ex.c {
		if err := cs.Close(); err != nil {
			log.Println(err)
		}
	}
	return nil
}

func (ex *executor[R]) loop(ctx context.Context) {
	actfail := 0
	for exit := true; exit; {
		select {
		case <-ctx.Done():
			log.Println("Exiting because of signal.")
			exit = false
		case <-time.After(time.Duration(ex.interval) * time.Second):
			if ex.fail > 0 && actfail >= ex.fail {
				log.Printf("Fail limit reached (%d). Exiting.\n", actfail)
				exit = false
			}

			v, err := ex.p.Produce()

			if err == nil {
				for i := range ex.c {
					if err := ex.c[i].Consume(v); err != nil {
						if ex.failOnConsume {
							actfail++
						}
						log.Println(err)
					}
				}
			} else {
				log.Printf("[%d] %v\n", actfail, err)
				actfail++
			}
		}
	}
	ex.cancel()
}

func (ex *executor[R]) stop(cmd *cobra.Command, args []string) error {
	d, err := ex.daemonContext().Search()
	if err != nil {
		log.Fatalf("Unable send signal to the daemon: %v", err)
	}
	if d == nil {
		log.Printf("Daemon process already stopped")
		return nil
	} else {
		return d.Kill()
	}
}

func (ex *executor[R]) termHandler(ch chan os.Signal) {
	for {
		select {
		case sig := <-ch:
			if sig == syscall.SIGUSR2 {
				fm, err := os.Create("/tmp/hassio-mem.pprof")
				if err != nil {
					log.Println(err)
				} else {
					pprof.WriteHeapProfile(fm)
				}
				fm.Close()
			} else {
				log.Println("Terminating....")
				pprof.StopCPUProfile()
				ex.cancel()
				time.Sleep(5 * time.Second)
				os.Exit(0)
			}
		}
	}
}
