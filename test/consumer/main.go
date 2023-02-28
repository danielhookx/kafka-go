package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/inconshreveable/log15"
	kafka "github.com/oofpgDLD/kafka-go"
)

var (
	topic, group          string
	broker                string
	Consumers             int
	Processors            int
	connTimeout           int
	batchCacheCapacity    int
	consumerCacheCapacity int
	graceful              bool
)

func init() {
	flag.IntVar(&Consumers, "cons", 1, "the number of work goroutines, write data to cache")
	flag.IntVar(&Processors, "pros", 1, "the number of work goroutines, read data from cache")
	flag.IntVar(&connTimeout, "cto", 7000, "connect timeout,ms")
	flag.IntVar(&batchCacheCapacity, "bcc", 0, "batch consumer cache size")
	flag.IntVar(&consumerCacheCapacity, "ccc", 0, "single consumer cache size")
	flag.StringVar(&topic, "topic", "test-mytest-topic", "kafka topic")
	flag.StringVar(&group, "group", "", "kafka consumer group")
	flag.StringVar(&broker, "broker", "127.0.0.1:9092", "kafka broker")
	flag.BoolVar(&graceful, "graceful", true, "stop service graceful")
}

//
func main() {
	flag.Parse()
	log.Info("service start", "broker", broker, "topic", topic, "group", group)
	log.Info("configs",
		"Consumers", Consumers, "Processors", Processors,
		"connTimeout", connTimeout,
		"batchCacheCapacity", batchCacheCapacity,
		"consumerCacheCapacity", consumerCacheCapacity,
		"graceful", graceful,
	)

	consumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Version:        "",
		Brokers:        []string{broker},
		Group:          group,
		Topic:          topic,
		CacheCapacity:  consumerCacheCapacity,
		ConnectTimeout: time.Millisecond * time.Duration(connTimeout),
	}, nil)
	log.Info("dial kafka broker success")
	bc := kafka.NewBatchConsumer(kafka.BatchConsumerConf{
		CacheCapacity: batchCacheCapacity,
		Consumers:     Consumers,
		Processors:    Processors,
	}, kafka.WithHandle(func(key string, data []byte) error {
		log.Info("receive msg:", "value", data)
		time.Sleep(time.Millisecond * 500)
		return nil
	}), consumer)

	bc.Start()

	// init signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			if graceful {
				bc.GracefulStop(context.Background())
			} else {
				bc.Stop()
			}
			time.Sleep(time.Second * 2)
			return
		case syscall.SIGHUP:
			// TODO reload
		default:
			return
		}
	}
}
