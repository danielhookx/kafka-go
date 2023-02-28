# kafka producer and consumer test

### build
```shell
make build
```

### batch produce and consume
```shell
# start consumer
./consumer --topic test-topic --group test-group --broker "127.0.0.1:9092" --cons 8 --pros 8

# start producer
./producer --topic test-topic --broker "127.0.0.1:9092" --number 16
```

### produce faster than consume with cache
```shell
# start consumer
./consumer --topic test-topic --group test-group --broker "127.0.0.1:9092" --cons 1 --pros 1

# start producer
./producer --topic test-topic --broker "127.0.0.1:9092" --number 4

# wait for a while，press command+c to close producer.
# consume from cache, after moment press command+c to close consumer, it will waiting for cache is empty and stop graceful.
```

### produce faster than consume without cache
```shell
# start consumer
./consumer --topic test-topic --group test-group --broker "127.0.0.1:9092" --cons 1 --pros 1 --bcc 1 --ccc 1

# start producer
./producer --topic test-topic --broker "127.0.0.1:9092" --number 4

# wait for a while，press command+c to close producer.
# consume from kafka broker, after moment press command+c to close consumer，service stop immediately.
```

### consumer connection timeout
```shell
./consumer --topic test-topic --group test-group --broker "127.0.0.1:9092" --cons 1 --pros 1 --bcc 1 --ccc 1 --cto 1
```