package broker

import (
	"time"

	"transcode-demo/pkg/cerrors"
	"transcode-demo/pkg/json"
)

type Status int
type EnvelopType string
type CommitType string
type AutoOffsetResetType string

const (
	Success Status = iota
	Retry
	Drop
	Republish
)

const (
	AfterProcess CommitType = "after_process"
	Async        CommitType = "async"
)

const (
	Earliest AutoOffsetResetType = "earliest"
	Latest   AutoOffsetResetType = "latest"
)

type ProcessStatus struct {
	Err    *cerrors.CError
	Status Status
}

type IEvent interface {
	Key() []byte
}

type Event struct {
	Topic   string     `json:"topic,omitempty"` // is exchange for rabbitmq
	Key     []byte     `json:"key,omitempty"`
	Envelop Envelop    `json:"envelop,omitempty"`
	Time    *time.Time `json:"time"`
}

type Envelop struct {
	Type    EnvelopType     `json:"type"`
	Retry   int             `json:"retry"`
	Message json.RawMessage `json:"message"`
}
