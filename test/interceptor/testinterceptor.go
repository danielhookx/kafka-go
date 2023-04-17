package interceptor

import (
	"context"

	"github.com/Shopify/sarama"
	"github.com/oofpgDLD/kafka-go"
)

func TestInterceptor(ctx context.Context, msg *sarama.ProducerMessage, handler kafka.ProducerHandler) (int32, int64, error) {
	//set msg
	return handler(ctx, &sarama.ProducerMessage{
		Key:       msg.Key,
		Topic:     "test-interceptor",
		Value:     msg.Value,
		Timestamp: msg.Timestamp,
	})
}
