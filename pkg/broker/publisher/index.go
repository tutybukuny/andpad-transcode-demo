package publisher

import (
	"context"
	"log"

	"transcode-demo/pkg/broker"
	"transcode-demo/pkg/json"
)

type IPublisher interface {
	PublishEvent(ctx context.Context, event *broker.Event) error
	CreateTopic(ctx context.Context, topic string, partitionNumber int) error
}

type PublisherWrapper[T broker.IEvent] struct {
	kafkaPub    IPublisher
	topic       string
	envelopType broker.EnvelopType
}

func New[T broker.IEvent](kafkaPub IPublisher, topic string, autoCreateTopic bool, envelopType broker.EnvelopType) *PublisherWrapper[T] {
	w := &PublisherWrapper[T]{
		kafkaPub:    kafkaPub,
		topic:       topic,
		envelopType: envelopType,
	}
	if autoCreateTopic {
		err := kafkaPub.CreateTopic(context.Background(), topic, 1)
		if err != nil {
			log.Fatalf("cannot create topic: %s", err)
		}
	}

	return w
}

func (w *PublisherWrapper[T]) PublishEvent(ctx context.Context, data T, topics ...string) error {
	topic := w.topic
	if len(topics) > 0 {
		topic = topics[0]
	}

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return w.kafkaPub.PublishEvent(ctx, &broker.Event{
		Topic: topic,
		Key:   data.Key(),
		Envelop: broker.Envelop{
			Type:    w.envelopType,
			Message: b,
		},
	})
}
