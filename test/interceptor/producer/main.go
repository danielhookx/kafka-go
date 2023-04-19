package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/oofpgDLD/kafka-go"
	"github.com/oofpgDLD/kafka-go/test/interceptor"
)

var log = log15.New("service", "test-producer")

var (
	start  int
	topic  string
	broker string
	number int
)

func init() {
	flag.IntVar(&start, "start", 0, "start print flag index")
	flag.IntVar(&number, "number", 1, "product concurrency numbers")
	flag.StringVar(&topic, "topic", "empty-topic", "kafka topic")
	flag.StringVar(&broker, "broker", "127.0.0.1:9092", "kafka brokers")
}

func main() {
	flag.Parse()
	p := kafka.NewProducer(kafka.ProducerConfig{
		Version: "",
		Brokers: []string{broker},
	}, kafka.WithProducerInterceptors(interceptor.TestProducerInterceptor))
	end := start + 1000

	log.Info("service start", "thread numbers", number, "broker", broker, "topic", topic, "start", start, "end", end)

	for i := 0; i < number; i++ {
		go func(index int) {
			for j := start; j < end; j++ {
				_, _, err := p.Publish(context.Background(), topic, "", []byte(fmt.Sprintf("%d-%d", index, j)))
				if err != nil {
					log.Error("publish failed", "err", err)
				}
				time.Sleep(time.Millisecond * 1000)
			}
		}(i)
	}

	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			time.Sleep(time.Second * 2)
			return
		case syscall.SIGHUP:
			// TODO reload
		default:
			return
		}
	}
}
