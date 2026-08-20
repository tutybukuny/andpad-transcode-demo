package utils

import (
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
