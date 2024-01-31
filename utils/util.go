package utils

import "golang.org/x/exp/constraints"

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
