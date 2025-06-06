package trace

import (
	"context"
	"fmt"
	"strconv"

	semconv "github.com/danielhookx/kafka-go/internal/semconv/v1.18.0"

	"github.com/Shopify/sarama"
	"github.com/danielhookx/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const Name = "kafka-go"

const (
	VersionCodeKey = "VersionCode"
)

func ProducerInterceptor(ctx context.Context, msg *sarama.ProducerMessage, handler kafka.ProducerHandler) (int32, int64, error) {
	ctx, span := startProducerSpan(ctx, msg)
	defer span.End()

	partition, offset, err := handler(ctx, msg)
	finishProducerSpan(span, partition, offset, err)
	return partition, offset, err
}

func startProducerSpan(ctx context.Context, msg *sarama.ProducerMessage) (context.Context, trace.Span) {
	tr := otel.Tracer(Name)
	carrier := NewProducerMessageCarrier(msg)

	spanCtx := otel.GetTextMapPropagator().Extract(ctx, carrier)
	// span name : https://opentelemetry.io/docs/reference/specification/trace/semantic_conventions/messaging/#span-name
	// The span name SHOULD be set to the message destination name and the operation being performed in the following format:
	//		<destination name> <operation name>
	name, attr := producerSpanInfo(msg, &carrier)
	ctx, span := tr.Start(spanCtx, name, trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(attr...))
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return ctx, span
}

func finishProducerSpan(span trace.Span, partition int32, offset int64, err error) {
	span.SetAttributes(
		semconv.MessagingKafkaDestinationPartition(int(partition)),
		semconv.MessagingKafkaMessageOffset(int(offset)),
	)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
}

// producerSpanInfo returns the span info.
func producerSpanInfo(msg *sarama.ProducerMessage, carrier *ProducerMessageCarrier) (string, []attribute.KeyValue) {
	name := fmt.Sprintf("%s %s", msg.Topic, "publish")

	var key string
	if k, ok := msg.Key.(sarama.StringEncoder); ok {
		key = string(k)
	}

	version, err := strconv.ParseInt(carrier.Get(VersionCodeKey), 10, 64)
	if err != nil {
		version = 1
	}

	attrs := []attribute.KeyValue{
		semconv.MessagingSystemKafka,                                         // messaging.system
		semconv.MessagingDestinationKindTopic,                                // messaging.destination.kind
		semconv.MessagingDestinationName(msg.Topic),                          // messaging.destination.name
		semconv.MessagingOperationPublish,                                    // messaging.operation
		semconv.MessagingKafkaMessageKey(key),                                // messaging.kafka.message.key
		semconv.MessagingMessagePayloadSizeBytes(msg.ByteSize(int(version))), // messaging.message.payload_size_bytes
	}
	return name, attrs
}
