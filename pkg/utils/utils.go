package utils

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"time"
	"unsafe"

	z "go.uber.org/zap"
)

const (
	TestBucket    = "local-bucket"
	TestBucketURL = "s3://local-bucket"
	TestMp4File   = "s3://local-bucket/test-data/input.mp4"
)

// JoinConfigKeys joins a list of configuration keys into a single string with dots as separators.
func JoinConfigKeys(keys ...string) string {
	return strings.Join(keys, ".")
}

func GetString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func ChoiceCondition[T any](cond bool, t T, f T) T {
	if cond {
		return t
	}
	return f
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(length ...int) string {
	l := 8
	if len(length) > 0 {
		l = length[0]
	}
	str := make([]rune, l)
	for i := range l {
		str[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(str)
}

func Recover(l *z.Logger, f func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			rErr, ok := r.(error)
			if !ok {
				rErr = fmt.Errorf("%v", r)
			}
			stack := make([]byte, 4<<10)
			length := runtime.Stack(stack, true)
			l.Error("panic recovered", z.Error(rErr), z.ByteString("stack", stack[:length]))
			err = rErr
		}
	}()
	err = f()
	return
}

func SleepWithContext(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(d):
		return
	}
}
