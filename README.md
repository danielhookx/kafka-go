# concurrency kafka consumer

## Go versions

kafka-go requires Go version 1.15 or later.

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
bc := kafka.NewBatchConsumer(batchConf, kafka.WithHandle(func(key string, data []byte) error {
    log.Info("receive msg:", "value", data)
    time.Sleep(time.Millisecond * 500)
    return nil
}), consumer)

bc.Start()
```

## Test
[[kafka producer and consumer test](test/README.md) | [kafka 消费者和生产者测试](test/README-CN.md)]

### License

kafka-go is under the MIT license. See the [LICENSE](LICENSE) file for details.
