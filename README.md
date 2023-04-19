# concurrency kafka consumer

## Go versions

kafka-go requires Go version 1.15 or later.

## Producer

```go
p := kafka.NewProducer(kafka.ProducerConfig{
    Version: "",
    Brokers: []string{broker},
})

_, _, err := p.Publish(context.Background(), topic, "key", []byte(string("val")))
```

## Consumer

```go
// config
singleConf := kafka.ConsumerConfig{
    Version:        "",
    Brokers:        []string{"127.0.0.1:9092"},
    Group:          "test-group",
    Topic:          "test",
    CacheCapacity:  100,
    ConnectTimeout: time.Millisecond * time.Duration(5000),
}
batchConf := kafka.BatchConsumerConf{
    CacheCapacity: 100,
    Consumers:     4,
    Processors:    4,
}

consumer := kafka.NewConsumer(singleConf, nil)
bc := kafka.NewBatchConsumer(batchConf, consumer, kafka.WithHandle(func(ctx context.Context, key string, data []byte) error {
    log.Info("receive msg:", "value", data)
    time.Sleep(time.Millisecond * 500)
    return nil
}))

bc.Start()
```

## trace interceptor

```go
imports(
	"github.com/oofpgDLD/kafka-go"
    "github.com/oofpgDLD/kafka-go/trace"
)

// new consumer with trace interceptor
func newConsumer(singleConf kafka.ConsumerConfig, batchConf kafka.BatchConsumerConf) {
    consumer := kafka.NewConsumer(singleConf, nil)
    bc := kafka.NewBatchConsumer(batchConf, consumer, kafka.WithHandle(func(ctx context.Context, key string, data []byte) error {
        log.Info("receive msg:", "value", data)
        time.Sleep(time.Millisecond * 500)
        return nil
    }), xkafka.WithBatchConsumerInterceptors(trace.ConsumeInterceptor))
}

// new producer with trace interceptor
func newProducer() {
    p := kafka.NewProducer(kafka.ProducerConfig{
        Version: "",
        Brokers: []string{broker},
    }, xkafka.WithProducerInterceptors(trace.ProducerInterceptor))
}
```

## Test
[[kafka producer and consumer test](test/README.md) | [kafka 消费者和生产者测试](test/README-CN.md)]

### License

kafka-go is under the MIT license. See the [LICENSE](LICENSE) file for details.
