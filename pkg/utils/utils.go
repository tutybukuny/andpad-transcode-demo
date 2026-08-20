package utils

import (
	"math/rand"
	"strings"
	"unsafe"
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
