package kafka

import (
	"context"

	"github.com/Shopify/sarama"
)

type ProducerHandler func(ctx context.Context, msg *sarama.ProducerMessage) (int32, int64, error)
type ProducerInterceptor func(ctx context.Context, msg *sarama.ProducerMessage, handler ProducerHandler) (int32, int64, error)

type ConsumeHandler func(ctx context.Context, msg *sarama.ConsumerMessage) error
type ConsumeInterceptor func(ctx context.Context, msg *sarama.ConsumerMessage, handler ConsumeHandler) error

func producerInterceptors(p *Producer) {
	// Prepend opts.unaryInt to the chaining interceptors if it exists, since unaryInt will
	// be executed before any other chained interceptors.
	interceptors := p.opts.interceptors
	var chainedInt ProducerInterceptor
	if len(interceptors) == 0 {
		chainedInt = nil
	} else if len(interceptors) == 1 {
		chainedInt = interceptors[0]
	} else {
		chainedInt = chainProducerInterceptors(interceptors)
	}
	p.interceptor = chainedInt
}

func chainProducerInterceptors(interceptors []ProducerInterceptor) ProducerInterceptor {
	return func(ctx context.Context, msg *sarama.ProducerMessage, handler ProducerHandler) (int32, int64, error) {
		// the struct ensures the variables are allocated together, rather than separately, since we
		// know they should be garbage collected together. This saves 1 allocation and decreases
		// time/call by about 10% on the microbenchmark.
		var state struct {
			i    int
			next ProducerHandler
		}
		state.next = func(ctx context.Context, msg *sarama.ProducerMessage) (int32, int64, error) {
			if state.i == len(interceptors)-1 {
				return interceptors[state.i](ctx, msg, handler)
			}
			state.i++
			return interceptors[state.i-1](ctx, msg, state.next)
		}
		return state.next(ctx, msg)
	}
}

func batchConsumerInterceptors(c *BatchConsumer) {
	// Prepend opts.unaryInt to the chaining interceptors if it exists, since unaryInt will
	// be executed before any other chained interceptors.
	interceptors := c.opts.interceptors
	var chainedInt ConsumeInterceptor
	if len(interceptors) == 0 {
		chainedInt = nil
	} else if len(interceptors) == 1 {
		chainedInt = interceptors[0]
	} else {
		chainedInt = chainBatchConsumerInterceptors(interceptors)
	}
	c.interceptor = chainedInt
}

func chainBatchConsumerInterceptors(interceptors []ConsumeInterceptor) ConsumeInterceptor {
	return func(ctx context.Context, msg *sarama.ConsumerMessage, handler ConsumeHandler) error {
		// the struct ensures the variables are allocated together, rather than separately, since we
		// know they should be garbage collected together. This saves 1 allocation and decreases
		// time/call by about 10% on the microbenchmark.
		var state struct {
			i    int
			next ConsumeHandler
		}
		state.next = func(ctx context.Context, msg *sarama.ConsumerMessage) error {
			if state.i == len(interceptors)-1 {
				return interceptors[state.i](ctx, msg, handler)
			}
			state.i++
			return interceptors[state.i-1](ctx, msg, state.next)
		}
		return state.next(ctx, msg)
	}
}
