package utils

import (
	"encoding/json"
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
	case float64:
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

func ToIntWithDefaultValue(i interface{}, defaultVal int) int {
	switch value := i.(type) {
	case int:
		return value
	case float64:
		return int(value)
	case uint:
		return int(value)
	case uint8:
		return int(value)
	case uint16:
		return int(value)
	case uint32:
		return int(value)
	case uint64:
		return int(value)
	case int16:
		return int(value)
	case int32:
		return int(value)
	case int64:
		return int(value)
	case string:
		if val, err := strconv.Atoi(value); err == nil {
			return val
		} else {
			return defaultVal
		}
	case json.Number:
		if val, err := value.Int64(); err == nil {
			return int(val)
		} else {
			return defaultVal
		}
	default:
		return defaultVal
	}
}

func ToStringWithDefaultValue(i interface{}, defaultVal string) string {
	switch value := i.(type) {
	case int:
		return strconv.Itoa(value)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(value), 'f', -1, 64)
	case int32:
		return strconv.Itoa(int(value))
	case int64:
		return strconv.FormatInt(value, 10)
	case string:
		return value
	case json.Number:
		return value.String()
	default:
		return defaultVal
	}
}
