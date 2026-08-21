package kafkaconsumer

import (
	"context"
	"fmt"
	"time"

	"transcode-demo/pkg/broker"
)

type Config struct {
	Brokers           []string
	Username          string
	Password          string
	Topic             string
	GroupID           string
	MaxBytes          int64
	CommitInterval    time.Duration
	MaxPollIntervalMs int64
	MaxPollRecords    int
	CommitType        broker.CommitType
	AutoOffsetReset   broker.AutoOffsetResetType
}

type KafkaEventInfo struct {
	Topic         string
	Partition     int
	Offset        int64
	HighWaterMark int64
	Key           []byte
	Headers       []Header
	Time          time.Time
}

type Header struct {
	Key   string // Header name (utf-8 string)
	Value []byte // Header value (nil, empty, or binary)
}

type KafkaHandler func(ctx context.Context, eventInfo KafkaEventInfo, envelop broker.Envelop) broker.ProcessStatus
type HandlerMap map[broker.EnvelopType]KafkaHandler

var (
	ContextCancelErr = fmt.Errorf("context canceled")
)
