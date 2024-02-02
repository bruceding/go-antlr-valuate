package utils

import (
	"reflect"

	"golang.org/x/exp/constraints"
)

func GT[T constraints.Ordered](x, y T) bool {
	return x > y
}

func GE[T constraints.Ordered](x, y T) bool {
	return x >= y
}

func LT[T constraints.Ordered](x, y T) bool {
	return x < y
}
func LE[T constraints.Ordered](x, y T) bool {
	return x <= y
}

func InArray[E constraints.Ordered, T any](x E, arr []T) bool {
	for _, v := range arr {
		if reflect.DeepEqual(x, v) {
			return true
		}
	}

	return false
}
