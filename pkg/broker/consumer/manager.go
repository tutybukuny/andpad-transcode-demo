package consumer

import (
	"strings"
	"sync"
	"time"

	z "go.uber.org/zap"

	"transcode-demo/pkg/broker"
	kafkaconsumer "transcode-demo/pkg/broker/consumer/kafka"
	confluentkafkagoconsumer "transcode-demo/pkg/broker/consumer/kafka/confluent-kafka-go"
	"transcode-demo/pkg/broker/publisher"
)

type Manager struct {
	l *z.Logger `container:"name"`

	cfg          broker.KafkaConfig
	kafkaBrokers []string
	consumers    []IConsumer
}

func New(cfg broker.KafkaConfig, l *z.Logger) *Manager {
	consumer := &Manager{
		l:         l,
		cfg:       cfg,
		consumers: make([]IConsumer, 0),
	}

	if cfg.Brokers != "" {
		consumer.kafkaBrokers = strings.Split(cfg.Brokers, ",")
	}

	return consumer
}

func (c *Manager) AddKafkaTopic(
	handler kafkaconsumer.KafkaHandler,
	kafkaPub publisher.IPublisher,
	cfg kafkaconsumer.Config,
) *Manager {
	commitInterval := time.Duration(0)
	if cfg.CommitType == broker.Async {
		commitInterval = 1 * time.Second
	}
	cfg.Brokers = c.kafkaBrokers
	cfg.Username = c.cfg.Username
	cfg.Password = c.cfg.Password
	cfg.CommitInterval = commitInterval
	if cfg.MaxBytes == 0 {
		cfg.MaxBytes = 10 * 1024 * 1024 //10MB
	}
	if cfg.CommitType == "" {
		cfg.CommitType = broker.AfterProcess
	}
	if cfg.AutoOffsetReset == "" {
		cfg.AutoOffsetReset = broker.Earliest
	}
	if cfg.MaxPollIntervalMs == 0 {
		cfg.MaxPollIntervalMs = 10 * 60 * 1000 // 10 minutes
	}
	c.consumers = append(c.consumers, confluentkafkagoconsumer.New(cfg, handler, kafkaPub))
	return c
}

func (c *Manager) Start() {
	for i := range c.consumers {
		consumer := c.consumers[i]
		go func() {
			if err := consumer.Listen(); err != nil {
				c.l.Fatal("error when listen", z.Error(err))
			}
		}()
	}
}

func (c *Manager) Stop() {
	w := sync.WaitGroup{}
	w.Add(len(c.consumers))
	for i := range c.consumers {
		consumer := c.consumers[i]
		go func() {
			defer w.Done()
			if err := consumer.Stop(); err != nil {
				c.l.Error("error when stop consumer", z.Error(err))
			}
		}()
	}
	w.Wait()
}
