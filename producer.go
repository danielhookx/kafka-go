package kafka

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/inconshreveable/log15"
)

type ProducerConfig struct {
	Version string   `json:",optional"`
	Brokers []string `json:",optional"`

	realVersion sarama.KafkaVersion `json:"-,optional"`
}

// funcProducerOption wraps a function that modifies producerOptions into an implementation of the ProducerOption interface.
type funcProducerOption struct {
	f func(*producerOptions)
}

func (fdo *funcProducerOption) apply(do *producerOptions) {
	fdo.f(do)
}

func newFuncProducerOption(f func(*producerOptions)) *funcProducerOption {
	return &funcProducerOption{
		f: f,
	}
}

// A ProducerOption sets options such as interceptor etc.
type ProducerOption interface {
	apply(*producerOptions)
}

type producerOptions struct {
	interceptors []ProducerInterceptor
}

// WithProducerInterceptors returns a ServerOption that sets the Interceptor for the producer.
func WithProducerInterceptors(interceptors ...ProducerInterceptor) ProducerOption {
	return newFuncProducerOption(func(o *producerOptions) {
		o.interceptors = append(o.interceptors, interceptors...)
	})
}

type Producer struct {
	brokers     []string
	conn        sarama.SyncProducer
	opts        producerOptions
	interceptor ProducerInterceptor
}

func ensureProducerConfig(cfg *ProducerConfig) {
	//version
	version, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		cfg.realVersion = sarama.V2_0_0_0
	}
	cfg.realVersion = version

	if len(cfg.Brokers) == 0 {
		panic(errors.New("brokers is empty"))
	}
}

func NewProducer(cfg ProducerConfig, opt ...ProducerOption) *Producer {
	ensureProducerConfig(&cfg)
	kc := sarama.NewConfig()
	kc.Version = cfg.realVersion //current kafka version
	//kc.Producer.Compression = sarama.CompressionSnappy //button of data compress,for improve transmission efficiency
	kc.Producer.RequiredAcks = sarama.WaitForAll //wait for all sync duplicate return ack
	kc.Producer.Retry.Max = 10                   //producer retry times
	kc.Producer.Return.Successes = true

	pub, err := sarama.NewSyncProducer(cfg.Brokers, kc)
	if err != nil {
		panic(err)
	}
	opts := producerOptions{}
	for _, o := range opt {
		o.apply(&opts)
	}
	p := &Producer{
		brokers: cfg.Brokers,
		conn:    pub,
		opts:    opts,
	}
	producerInterceptors(p)
	log.Info("producer dial kafka server", "brokers", cfg.Brokers)
	return p
}

func (p *Producer) Publish(topic, k string, v []byte) (int32, int64, error) {
	if k == "" {
		k = strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	log.Debug("push params", "topic", topic, "brokers", p.brokers)
	m := &sarama.ProducerMessage{
		Key:       sarama.StringEncoder(k),
		Topic:     topic,
		Value:     sarama.ByteEncoder(v),
		Timestamp: time.Now(),
	}
	partition, offset, err := p.conn.SendMessage(m)
	return partition, offset, err
}

func (p *Producer) PublishV2(ctx context.Context, topic, k string, v []byte) (int32, int64, error) {
	if k == "" {
		k = strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	m := &sarama.ProducerMessage{
		Key:       sarama.StringEncoder(k),
		Topic:     topic,
		Value:     sarama.ByteEncoder(v),
		Timestamp: time.Now(),
	}
	if p.interceptor == nil {
		return p.publish(ctx, m)
	}
	return p.interceptor(ctx, m, p.publish)
}

func (p *Producer) publish(ctx context.Context, msg *sarama.ProducerMessage) (int32, int64, error) {
	log.Debug("push params", "topic", msg.Topic, "brokers", p.brokers)
	return p.conn.SendMessage(msg)
}

func (p *Producer) Close() error {
	return p.conn.Close()
}
