package confluentkafkagoconsumer

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	z "go.uber.org/zap"

	"transcode-demo/pkg/broker"
	kafkaconsumer "transcode-demo/pkg/broker/consumer/kafka"
	"transcode-demo/pkg/broker/publisher"
	"transcode-demo/pkg/json"
)

type consumerImpl struct {
	l *z.Logger `container:"name"`

	config   kafkaconsumer.Config
	consumer *kafka.Consumer
	handler  kafkaconsumer.KafkaHandler
	kafkaPub publisher.IPublisher
	stopChan chan struct{}
}

func New(config kafkaconsumer.Config, handler kafkaconsumer.KafkaHandler, kafkaPub publisher.IPublisher) *consumerImpl {
	c := &consumerImpl{
		config:   config,
		handler:  handler,
		kafkaPub: kafkaPub,
		stopChan: make(chan struct{}),
	}

	cfgMap := kafka.ConfigMap{
		"bootstrap.servers":     strings.Join(config.Brokers, ","),
		"group.id":              config.GroupID,
		"session.timeout.ms":    30000,
		"auto.offset.reset":     string(config.AutoOffsetReset),
		"max.poll.interval.ms":  int(config.MaxPollIntervalMs),
		"heartbeat.interval.ms": 3000,
	}
	if config.CommitType == broker.AfterProcess {
		cfgMap["enable.auto.offset.store"] = false
		cfgMap["enable.auto.commit"] = false
	}

	var err error
	c.consumer, err = kafka.NewConsumer(&cfgMap)
	if err != nil {
		c.l.Fatal("cannot create new consumer", z.Any("config", config), z.Error(err))
	}

	if err = c.consumer.Subscribe(config.Topic, nil); err != nil {
		c.l.Fatal("cannot subscribe to topic", z.Any("config", config), z.Error(err))
	}
	c.l.Info("listening to topic", z.String("topic", config.Topic), z.String("group_id", config.GroupID))

	return c
}

func (c *consumerImpl) Listen() error {
	ctx := context.Background()
	commitAfterProcess := c.config.CommitType == broker.AfterProcess
	for {
		select {
		case <-c.stopChan:
			return nil
		default:
			ev := c.consumer.Poll(100)
			switch e := ev.(type) {
			case *kafka.Message:
				c.processEvent(ctx, e)

				if commitAfterProcess {
					_, err := c.consumer.CommitMessage(e)
					if err != nil {
						c.l.Error(fmt.Sprintf("cannot commit message to kafka, topic: %s, group_id: %s, message_offset: %d", c.config.Topic, c.config.GroupID, e.TopicPartition.Offset), z.Error(err))
						return fmt.Errorf("cannot commit message to kafka: %w", err)
					}
					continue
				}
				_, err := c.consumer.StoreMessage(e)
				if err != nil {
					c.l.Error(fmt.Sprintf("cannot store message, topic: %s, group_id: %s, message_offset: %d", c.config.Topic, c.config.GroupID, e.TopicPartition.Offset), z.Error(err))
					return fmt.Errorf("cannot store message: %w", err)
				}
			case kafka.Error:
				c.l.Error("cannot poll message from kafka, topic: %s, group_id: %s",
					z.String("topic", c.config.Topic), z.String("group_id", c.config.GroupID), z.Error(e))
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (c *consumerImpl) Stop() error {
	c.stopChan <- struct{}{}
	return nil
}

func (c *consumerImpl) processEvent(ctx context.Context, event *kafka.Message) {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			stack := make([]byte, 4<<10)
			length := runtime.Stack(stack, true)
			c.l.Error("panic recovered", z.Error(err), z.ByteString("stack", stack[:length]))
		}
	}()

	var envelop broker.Envelop
	if err := json.Unmarshal(event.Value, &envelop); err != nil {
		c.l.Error("cannot unmarshal envelop", z.ByteString("value", event.Value), z.Error(err))
		return
	}

	info := kafkaconsumer.KafkaEventInfo{
		Topic:     *event.TopicPartition.Topic,
		Partition: int(event.TopicPartition.Partition),
		Offset:    int64(event.TopicPartition.Offset),
		Key:       event.Key,
		Headers:   make([]kafkaconsumer.Header, 0, len(event.Headers)),
		Time:      event.Timestamp,
	}
	for i := range event.Headers {
		info.Headers = append(info.Headers, kafkaconsumer.Header(event.Headers[i]))
	}

	status := c.handler(ctx, info, envelop)
	if status.Err != nil {
		c.l.Error("error when handle event", z.Any("info", info), z.ByteString("message", event.Value), z.Error(status.Err))
	}
	switch status.Status {
	case broker.Republish:
		if c.kafkaPub == nil {
			c.l.Error("cannot republish event because of nil kafkaPub")
			break
		}
		envelop.Retry++
		e := broker.Event{
			Topic:   info.Topic,
			Key:     info.Key,
			Envelop: envelop,
		}
		if err := c.kafkaPub.PublishEvent(ctx, &e); err != nil {
			c.l.Error("cannot publish retry event", z.Any("event", e))
		}
	case broker.Success:
		c.l.Debug("processed event", z.Any("info", info), z.ByteString("message", event.Value))
	}
}
