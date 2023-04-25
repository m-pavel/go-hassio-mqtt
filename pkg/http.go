package ghm

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

type HttpServer[R any] struct {
	port   int
	engine *gin.Engine

	mem MemoryConsumer[R]

	ToJsonConverter func(R) any
	ToRawConverter  func(R) Entry
	YAxis           []string
	indexContent    *template.Template
	name            string
}

func (ht *HttpServer[R]) Setup(cmd *cobra.Command, name string) {
	cmd.PersistentFlags().IntVar(&ht.port, "port", 2999, "Http port")
	ht.mem = MemoryConsumer[R]{}
	ht.mem.Setup(cmd, name)
	ht.name = name
}

func (ht *HttpServer[R]) Init(d bool) error {
	ht.mem.Init(d)

	if ht.ToJsonConverter == nil {
		ht.ToJsonConverter = func(r R) any {
			return r
		}
	}
	if ht.ToRawConverter == nil {
		ht.ToRawConverter = func(r R) Entry {
			return Entry{"value": r}
		}
	}
	if ht.YAxis == nil {
		ht.YAxis = []string{"value"}
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
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	ht.indexContent = template.New("index")
	var err error
	if ht.indexContent, err = ht.indexContent.Parse(index_html); err != nil {
		return err
	}
	return nil
}
func (ht *HttpServer[R]) Consume(v R) error {
	return ht.mem.Consume(v)
}
func (ht *HttpServer[R]) Close() error {
	return ht.mem.Close()
}

func (ht *HttpServer[R]) current(c *gin.Context) {
	c.Writer.Header().Add("Content-type", "application/json")
	json.NewEncoder(c.Writer).Encode(ht.ToJsonConverter(ht.mem.Last()))
}

func (ht *HttpServer[R]) data(c *gin.Context) {
	c.Writer.Header().Add("Content-type", "application/json")
	json.NewEncoder(c.Writer).Encode(ht.mem.Data(ht.ToRawConverter))
}

func (ht *HttpServer[R]) index(c *gin.Context) {
	ht.indexContent.Execute(c.Writer, &IndexModel{Yaxis: ht.name, Marks: ht.YAxis})
}
