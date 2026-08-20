package utils

import "github.com/bwmarrin/snowflake"

var generator *snowflake.Node

func init() {
	snowflake.Epoch = 1787184000000 // 2026-08-20
	generator, _ = snowflake.NewNode(1)
}

func NewID() string {
	return generator.Generate().String()
}

func NewIDInt64() int64 {
	return generator.Generate().Int64()
}
