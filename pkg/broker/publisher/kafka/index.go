package kafkapublisher

import (
	"context"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	z "go.uber.org/zap"

	"transcode-demo/pkg/broker"
	"transcode-demo/pkg/json"
)

type PublisherImpl struct {
	l *z.Logger `container:"name"`

	writer *kafka.Writer
}

func New(cfg broker.KafkaConfig, l *z.Logger) *PublisherImpl {
	p := &PublisherImpl{l: l}
	brokers := strings.Split(cfg.Brokers, ",")

	var transport kafka.RoundTripper
	if cfg.Username != "" {
		mechanism, err := scram.Mechanism(scram.SHA512, cfg.Username, cfg.Password)
		if err != nil {
			p.l.Fatal("cannot create mechanism", z.Any("config", cfg), z.Error(err))
		}
		transport = &kafka.Transport{
			SASL: mechanism,
		}
	}

	p.writer = &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		Transport:    transport,
		BatchTimeout: 5 * time.Millisecond,
	}

	return p
}

func (p PublisherImpl) PublishEvent(ctx context.Context, event *broker.Event) error {
	value, err := json.Marshal(event.Envelop)
	if err != nil {
		return err
	}
	message := kafka.Message{
		Topic: event.Topic,
		Key:   event.Key,
		Value: value,
	}
	if event.Time != nil {
		message.Time = *event.Time
	}
	return p.writer.WriteMessages(ctx, message)
}

func (p PublisherImpl) CreateTopic(ctx context.Context, topic string, partitionNumber int) error {
	client := kafka.Client{
		Addr:      p.writer.Addr,
		Timeout:   p.writer.WriteTimeout,
		Transport: p.writer.Transport,
	}

	_, err := client.CreateTopics(ctx, &kafka.CreateTopicsRequest{
		Addr: client.Addr,
		Topics: []kafka.TopicConfig{
			{
				Topic:             topic,
				NumPartitions:     partitionNumber,
				ReplicationFactor: -1,
			},
		},
	})
	return err
}
