package valuate

import (
	"reflect"
	"testing"
	"time"

	"github.com/bruceding/go-antlr-valuate/parser"
)

func TestExpression(t *testing.T) {
	expr := "1+2"
	evaluableExpression, err := NewEvaluableExpression(expr)
	if err != nil {
		t.Fatal(err)
	}
	if evaluableExpression == nil {
		t.Fatal("evaluableExpression is nil")
	}
}

func TestEvaluableExpressionResult(t *testing.T) {

	testCases := []struct {
		input       string
		expectType  string
		expectValue any
		params      map[string]any
	}{
		{
			input:       "foo > 0",
			expectType:  "bool",
			expectValue: false,
			params:      map[string]any{"foo": -1},
		},
		{
			input:       "(requests_made * requests_succeeded / 100) >= 90",
			expectType:  "bool",
			expectValue: false,
			params:      map[string]any{"requests_made": 100, "requests_succeeded": 80},
		},
		{
			input:       "http_response_body == 'service is ok'",
			expectType:  "bool",
			expectValue: true,
			params:      map[string]any{"http_response_body": "service is ok"},
		},
		{
			input:       "${http-response_body} == 'service is ok'",
			expectType:  "bool",
			expectValue: true,
			params:      map[string]any{"http-response_body": "service is ok"},
		},
		{
			input:       "1 + 3 * 5",
			expectType:  "float64",
			expectValue: float64(16),
			params:      map[string]any{},
		},
		{
			input:       "max(3, 5)",
			expectType:  "float64",
			expectValue: float64(5),
			params:      map[string]any{},
		},
		{
			input:       "(mem_used / total_mem) * 100",
			expectType:  "float64",
			expectValue: float64(50),
			params:      map[string]any{"mem_used": 512, "total_mem": 1024},
		},
	}
	for _, tcase := range testCases {
		evaluableExpression, err := NewEvaluableExpression(tcase.input)
		if err != nil {
			t.Fatal(evaluableExpression)
		}
		evaluableExpression.Evaluate(tcase.params)
		result, err := evaluableExpression.Evaluate(tcase.params)

		if err != nil {
			t.Fatal(err)
		}

		switch val := result.(type) {
		case []any:
			if tcase.expectType != "[]any" {
				t.Fatal("expect type is []any")
			}

			if reflect.DeepEqual(val, tcase.expectValue) != true {
				t.Fatal("expect value is not equal")
			}
		case float64:
			if tcase.expectType != "float64" {
				t.Fatal("expect type is float64")
			}
			if val != tcase.expectValue {
				t.Fatal("expect value is not equal")
			}
		case string:
			if tcase.expectType != "string" {
				t.Fatal("expect type is string")
			}
			if val != tcase.expectValue {
				t.Fatal("expect value is not equal")
			}
		case bool:
			if tcase.expectType != "bool" {
				t.Fatal("expect type is bool")
			}
			if val != tcase.expectValue {
				t.Fatalf("expect value is not equal,input:%s, expect:%v, actual:%v", tcase.input, tcase.expectValue, val)
			}
		case time.Time:
			if tcase.expectType != "Time" {
				t.Fatal("expect type is Time")
			}
			if val != tcase.expectValue {
				t.Fatal("expect value is not equal")
			}
		case error:
			t.Fatal(val)
		default:
			t.Fatal("expect result error", result)
		}

	}
}

func TestEvaluableExpressionFunctions(t *testing.T) {
	functions := map[string]parser.ExpressionFunction{
		"strlen": func(args ...interface{}) (interface{}, error) {
			length := len(args[0].(string))
			return (float64)(length), nil
		},
	}

	expString := "strlen('someReallyLongInputString') <= 16"
	expression, _ := NewEvaluableExpressionWithFunctions(expString, functions)

	result, _ := expression.Evaluate(nil)
	if result.(bool) != false {
		t.Fatal("expect result wrong, expected false")
	}

	expString = "len('someReallyLongInputString') <= 16"
	expression, _ = NewEvaluableExpressionWithFunctions(expString, map[string]parser.ExpressionFunction{})

	result, _ = expression.Evaluate(nil)
	if result.(bool) != false {
		t.Fatal("expect result wrong, expected false")
	}
}
