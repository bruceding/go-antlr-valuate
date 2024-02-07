package utils

import (
	"reflect"
	"strconv"
)

func ToInt(x any) int {
	switch x := x.(type) {
	case int:
		return x
	case uint:
		return int(x)
	case int64:
		return int(x)
	case int32:
		return int(x)
	case int16:
		return int(x)
	case int8:
		return int(x)
	case uint16:
		return int(x)
	case uint32:
		return int(x)
	case uint64:
		return int(x)
	}

	return 0
}

func ReflectValueof(in reflect.Type, val any) reflect.Value {
	switch in.Kind() {
	case reflect.Int:
		switch v := val.(type) {
		case int:
			return reflect.ValueOf(v)
		case float64:
			return reflect.ValueOf(int(v))
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return reflect.ValueOf(i)
			}
		}

	case reflect.Float64:
		switch v := val.(type) {
		case float64:
			return reflect.ValueOf(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return reflect.ValueOf(f)
			}
		}
	}

	return reflect.ValueOf(val)
}
