package interceptor

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/danielhookx/kafka-go"
)

func TestProducerInterceptor(ctx context.Context, msg *sarama.ProducerMessage, handler kafka.ProducerHandler) (int32, int64, error) {
	//set msg
	return handler(ctx, &sarama.ProducerMessage{
		Key:       msg.Key,
		Topic:     "test-interceptor",
		Value:     msg.Value,
		Timestamp: msg.Timestamp,
	})
}

func TestConsumerInterceptor(ctx context.Context, msg *sarama.ConsumerMessage, handler kafka.ConsumeHandler) error {
	//get msg
	prefix := []byte("from-test-interceptor:")
	val := make([]byte, len(prefix)+len(msg.Value))
	copy(val, prefix)
	copy(val[len(prefix):], msg.Value)
	return handler(ctx, &sarama.ConsumerMessage{
		Key:       msg.Key,
		Topic:     msg.Topic,
		Value:     val,
		Timestamp: msg.Timestamp,
	})
}
