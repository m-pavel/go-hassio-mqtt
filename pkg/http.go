package ghm

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

type HttpServer[R any] struct {
	port   int
	engine *gin.Engine

	mem MemoryConsumer[R]

	Converter func(R) any
}

func (ht *HttpServer[R]) Setup(cmd *cobra.Command, name string) {
	cmd.PersistentFlags().IntVar(&ht.port, "port", 2999, "Http port")
	ht.mem = MemoryConsumer[R]{}
	ht.mem.Setup(cmd, name)
}
func (ht *HttpServer[R]) Init(d bool) error {

	ht.mem.Init(d)

	if ht.Converter == nil {
		ht.Converter = func(r R) any {
			return r
		}
	}
	ht.engine = gin.Default()
	ht.engine.GET("/api/v1/current", ht.current)
	ht.engine.GET("/api/v1/data", ht.data)
	ht.engine.GET("/", ht.index)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", ht.port),
		Handler: ht.engine,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	return nil
}
func (ht *HttpServer[R]) Consume(v R) error {
	return ht.mem.Consume(v)
}
func (ht *HttpServer[R]) Close() error {
	return ht.mem.Close()
}

func (ls *HttpServer[R]) current(c *gin.Context) {
	c.Writer.Header().Add("Content-type", "application/json")
	json.NewEncoder(c.Writer).Encode(ls.Converter(ls.mem.Last()))
}

func (ls *HttpServer[R]) data(c *gin.Context) {

}

func (ls *HttpServer[R]) index(c *gin.Context) {

}
