package parser

import (
	"fmt"
	"math"
)

type ExpressionFunction func(arguments ...interface{}) (interface{}, error)

var (
	buildInFunctionsMap = map[string]ExpressionFunction{
		"len": func(args ...interface{}) (interface{}, error) {
			switch val := args[0].(type) {
			case string:
				return len(val), nil
			case []any:
				return len(val), nil
			default:
				return nil, fmt.Errorf("len() function argument must be string or array, but got %T", val)
			}
		},

		"abs": func(arguments ...interface{}) (interface{}, error) {
			return math.Abs(arguments[0].(float64)), nil
		},
		"exp": func(arguments ...interface{}) (interface{}, error) {
			return math.Exp(arguments[0].(float64)), nil
		},
		"log": func(arguments ...interface{}) (interface{}, error) {
			return math.Log(arguments[0].(float64)), nil
		},
		"log10": func(arguments ...interface{}) (interface{}, error) {
			return math.Log10(arguments[0].(float64)), nil
		},
		"max": func(arguments ...interface{}) (interface{}, error) {
			return math.Max(arguments[0].(float64), arguments[1].(float64)), nil
		},
		"min": func(arguments ...interface{}) (interface{}, error) {
			return math.Min(arguments[0].(float64), arguments[1].(float64)), nil
		},
		"sqrt": func(arguments ...interface{}) (interface{}, error) {
			return math.Sqrt(arguments[0].(float64)), nil
		},
	}
)
