package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	kafka "github.com/danielhookx/kafka-go"
	"github.com/danielhookx/kafka-go/test/mock"
	log "github.com/inconshreveable/log15"
)

var (
	topic, group string
	broker       string
	Consumers    int
	Processors   int
)

func init() {
	flag.IntVar(&Consumers, "cons", 1, "the number of work goroutines, write data to cache")
	flag.IntVar(&Processors, "pros", 1, "the number of work goroutines, read data from cache")
	flag.StringVar(&topic, "topic", "test-mytest-topic", "kafka topic")
	flag.StringVar(&group, "group", "", "kafka consumer group")
	flag.StringVar(&broker, "broker", "127.0.0.1:9092", "kafka broker")
}

func main() {
	flag.Parse()
	log.Info("service start", "broker", broker, "topic", topic, "group", group, "Consumers", Consumers, "Processors", Processors)

	consumer := mock.NewConsumer(time.Millisecond * 500)
	log.Info("dial kafka broker success")
	bc := kafka.NewBatchConsumer(kafka.BatchConsumerConf{
		CacheCapacity: 0,
		Consumers:     Consumers,
		Processors:    Processors,
	}, consumer, kafka.WithHandle(func(ctx context.Context, key string, data []byte) error {
		log.Info("receive msg:", "value", data)
		return nil
	}))

	bc.Start()

	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			bc.Stop()
			time.Sleep(time.Second * 2)
			return
		case syscall.SIGHUP:
			// TODO reload
		default:
			return
		}
	}
}
