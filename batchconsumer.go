package kafka

import (
	"context"
	"io"
	"sync/atomic"

	"github.com/Shopify/sarama"
	"github.com/gammazero/workerpool"
	"github.com/inconshreveable/log15"
)

type IConsumer interface {
	FetchMessage(ctx context.Context) (message *sarama.ConsumerMessage, err error)
	Close() error
}

type ConsumeHandle func(ctx context.Context, key string, data []byte) error

const (
	defaultQueueCapacity = 1000
)

// funcBatchConsumerOption wraps a function that modifies batchConsumerOptions into an implementation of the BatchConsumerOption interface.
type funcBatchConsumerOption struct {
	f func(options *batchConsumerOptions)
}

func (fdo *funcBatchConsumerOption) apply(do *batchConsumerOptions) {
	fdo.f(do)
}

func newFuncBatchConsumerOption(f func(*batchConsumerOptions)) *funcBatchConsumerOption {
	return &funcBatchConsumerOption{
		f: f,
	}
}

type (
	BatchConsumerConf struct {
		CacheCapacity int `json:",optional"`
		Consumers     int `json:",optional"`
		Processors    int `json:",optional"`
	}

	batchConsumerOptions struct {
		logger       log15.Logger
		interceptors []ConsumeInterceptor
		handler      ConsumeHandler
	}
	// A BatchConsumerOption sets options such as interceptor etc.
	BatchConsumerOption interface {
		apply(*batchConsumerOptions)
	}

	BatchConsumer struct {
		cfg             BatchConsumerConf
		log             log15.Logger
		consumer        IConsumer
		channel         chan *sarama.ConsumerMessage
		handler         ConsumeHandler
		producerWorkers *workerpool.WorkerPool
		consumerWorkers *workerpool.WorkerPool
		isClosed        int32
		opts            batchConsumerOptions
		interceptor     ConsumeInterceptor
	}
)

// WithBatchConsumerInterceptors returns a ServerOption that sets the Interceptor for the producer.
func WithBatchConsumerInterceptors(interceptors ...ConsumeInterceptor) BatchConsumerOption {
	return newFuncBatchConsumerOption(func(o *batchConsumerOptions) {
		o.interceptors = append(o.interceptors, interceptors...)
	})
}

func WithLogger(logger log15.Logger) BatchConsumerOption {
	return newFuncBatchConsumerOption(func(o *batchConsumerOptions) {
		o.logger = logger
	})
}

func WithHandle(handle ConsumeHandle) BatchConsumerOption {
	if handle == nil {
		panic("handle func is nil")
	}
	return newFuncBatchConsumerOption(func(o *batchConsumerOptions) {
		o.handler = func(ctx context.Context, msg *sarama.ConsumerMessage) error {
			return handle(ctx, string(msg.Key), msg.Value)
		}
	})
}

func ensureConfigOptions(cfg *BatchConsumerConf, options *batchConsumerOptions) {
	if options.handler == nil {
		panic("handler is nil")
	}
	if options.logger == nil {
		options.logger = log15.New()
	}
	if cfg.CacheCapacity <= 0 {
		cfg.CacheCapacity = defaultQueueCapacity
	}
	if cfg.Consumers == 0 {
		cfg.Consumers = 8
	}
	if cfg.Processors == 0 {
		cfg.Processors = 8
	}
}

func NewBatchConsumer(cfg BatchConsumerConf, consumer IConsumer, opt ...BatchConsumerOption) *BatchConsumer {
	opts := batchConsumerOptions{}
	for _, o := range opt {
		o.apply(&opts)
	}
	ensureConfigOptions(&cfg, &opts)
	bc := &BatchConsumer{
		cfg:             cfg,
		log:             opts.logger,
		consumer:        consumer,
		channel:         make(chan *sarama.ConsumerMessage, cfg.CacheCapacity),
		handler:         opts.handler,
		producerWorkers: workerpool.New(cfg.Consumers),
		consumerWorkers: workerpool.New(cfg.Processors),
		opts:            opts,
	}
	batchConsumerInterceptors(bc)
	return bc
}

func (bc *BatchConsumer) startProducers() {
	for i := 0; i < bc.cfg.Consumers; i++ {
		bc.producerWorkers.Submit(func() {
			defer bc.close()

			for {
				msg, err := bc.consumer.FetchMessage(context.TODO())
				// io.EOF means consumer closed
				// io.ErrClosedPipe means committing messages on the consumer,
				// kafka will refire the messages on uncommitted messages, ignore
				if err == io.EOF || err == io.ErrClosedPipe {
					bc.log.Info("fetchMessage io.EOF or io.ErrClosedPipe")
					return
				}
				if err != nil {
					bc.log.Error("Error on reading message", "err", err.Error())
					continue
				}
				if msg == nil {
					continue
				}
				bc.log.Debug("fetchMessage", "msg", string(msg.Value))
				bc.channel <- msg
			}
		})
	}
}

func (bc *BatchConsumer) startConsumers() {
	for i := 0; i < bc.cfg.Processors; i++ {
		bc.consumerWorkers.Submit(func() {
			for msg := range bc.channel {
				if err := bc.consumeOne(msg); err != nil {
					bc.log.Error("Error on consuming message", "msg", string(msg.Value), "err", err)
				}
			}
		})
	}
}

func (bc *BatchConsumer) consumeOne(msg *sarama.ConsumerMessage) error {
	if bc.interceptor == nil {
		return bc.handler(context.TODO(), msg)
	}

	return bc.interceptor(context.TODO(), msg, bc.handler)
}

func (bc *BatchConsumer) close() {
	if atomic.CompareAndSwapInt32(&bc.isClosed, 0, 1) {
		bc.consumer.Close()
		close(bc.channel)
	}
}

func (bc *BatchConsumer) Start() {
	bc.startConsumers()
	bc.startProducers()
}

func (bc *BatchConsumer) Stop() {
	bc.close()
}

func (bc *BatchConsumer) GracefulStop(ctx context.Context) {
	down := make(chan struct{})
	go func() {
		bc.close()
		bc.consumerWorkers.StopWait()
		close(down)
	}()

	select {
	case <-ctx.Done():
		return
	case <-down:
		return
	}
}
