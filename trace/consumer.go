package trace

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/danielhookx/kafka-go"
	semconv "github.com/danielhookx/kafka-go/internal/semconv/v1.18.0"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func ConsumeInterceptor(ctx context.Context, msg *sarama.ConsumerMessage, handler kafka.ConsumeHandler) error {
	ctx, span := startConsumeSpan(ctx, msg)
	defer span.End()

	err := handler(ctx, msg)
	return err
}

func startConsumeSpan(ctx context.Context, msg *sarama.ConsumerMessage) (context.Context, trace.Span) {
	tr := otel.Tracer(Name)
	carrier := NewConsumerMessageCarrier(msg)

	spanCtx := otel.GetTextMapPropagator().Extract(ctx, carrier)
	// span name : https://opentelemetry.io/docs/reference/specification/trace/semantic_conventions/messaging/#span-name
	// The span name SHOULD be set to the message destination name and the operation being performed in the following format:
	//		<destination name> <operation name>
	name, attr := consumeSpanInfo(msg, &carrier)
	ctx, span := tr.Start(spanCtx, name, trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(attr...))
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return ctx, span
}

// consumeSpanInfo returns the span info.
func consumeSpanInfo(msg *sarama.ConsumerMessage, carrier *ConsumerMessageCarrier) (string, []attribute.KeyValue) {
	name := fmt.Sprintf("%s %s", msg.Topic, "receive")

	attrs := []attribute.KeyValue{
		semconv.MessagingSystemKafka,                      // messaging.system
		semconv.MessagingSourceKindTopic,                  // messaging.source.kind
		semconv.MessagingSourceName(msg.Topic),            // messaging.source.name
		semconv.MessagingOperationReceive,                 // messaging.operation
		semconv.MessagingKafkaMessageKey(string(msg.Key)), // messaging.kafka.message.key
		semconv.MessagingKafkaDestinationPartition(int(msg.Partition)),
		semconv.MessagingKafkaMessageOffset(int(msg.Offset)),
		//The identifier for the consumer receiving a message.
		//For Kafka, set it to {messaging.kafka.consumer.group} - {messaging.kafka.client_id}, if both are present, or only messaging.kafka.consumer.group.
		semconv.MessagingConsumerID("TODO {messaging.kafka.consumer.group} - {messaging.kafka.client_id}"),
	}
	return name, attrs
}
