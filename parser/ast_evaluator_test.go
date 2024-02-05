package parser

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/antlr4-go/antlr/v4"
)

func TestPrimaryParse(t *testing.T) {
	expectT, _ := time.Parse("2006-01-02", "2019-01-01")
	testCases := []struct {
		input       string
		expectType  string
		expectValue any
	}{
		{
			input:       "123.45",
			expectType:  "float64",
			expectValue: float64(123.45),
		},
		{
			input:       "(1,2,3,4,5)",
			expectType:  "[]any",
			expectValue: []any{float64(1), float64(2), float64(3), float64(4), float64(5)},
		},
		{
			input:       "'nomalstring'",
			expectType:  "string",
			expectValue: "nomalstring",
		},
		{
			input:       "true",
			expectType:  "bool",
			expectValue: true,
		},
		{
			input:       "'2019-01-01'",
			expectType:  "Time",
			expectValue: expectT,
		},
	}
	for _, tcase := range testCases {
		lexer := NewGovaluateLexer(antlr.NewInputStream(tcase.input))

		stream := antlr.NewCommonTokenStream(lexer, antlr.LexerDefaultTokenChannel)

		parser := NewGovaluateParser(stream)

		expression := parser.Expression()

		ast := NewASTEvaluator()
		result := ast.Visit(expression)

		fmt.Println(result)
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
				t.Fatal("expect value is not equal")
			}
		case time.Time:
			if tcase.expectType != "Time" {
				t.Fatal("expect type is Time")
			}
			if val != tcase.expectValue {
				t.Fatal("expect value is not equal")
			}
		default:
			fmt.Println(t)
		}

	}

}

func TestOperatorParse(t *testing.T) {
	testCases := []struct {
		input       string
		expectType  string
		expectValue any
	}{
		{
			input:       "1+2",
			expectType:  "float64",
			expectValue: float64(3),
		},

		{
			input:       "'hello' + ' world'",
			expectType:  "string",
			expectValue: "hello world",
		},

		{
			input:       "(1,2) + (3,4)",
			expectType:  "[]any",
			expectValue: []any{float64(1), float64(2), float64(3), float64(4)},
		},

		{
			input:       "5 - 2",
			expectType:  "float64",
			expectValue: float64(3),
		},
		{
			input:       "5 * 2",
			expectType:  "float64",
			expectValue: float64(10),
		},
		{
			input:       "4 / 2",
			expectType:  "float64",
			expectValue: float64(2),
		},
		{
			input:       "4 % 2",
			expectType:  "float64",
			expectValue: float64(0),
		},
		{
			input:       "1 + 3 * 5",
			expectType:  "float64",
			expectValue: float64(16),
		},
		{
			input:       "3 ** 4",
			expectType:  "float64",
			expectValue: float64(81),
		},
		{
			input:       "3 ^ 4",
			expectType:  "float64",
			expectValue: float64(81),
		},
		{
			input:       "(1 + 3) * 5",
			expectType:  "float64",
			expectValue: float64(20),
		},
		{
			input:       "~3",
			expectType:  "float64",
			expectValue: float64(-4),
		},
		{
			input:       "-1 - 3",
			expectType:  "float64",
			expectValue: float64(-4),
		},
		{
			input:       "1 == 2",
			expectType:  "bool",
			expectValue: false,
		},
		{
			input:       "!(1 == 2)",
			expectType:  "bool",
			expectValue: true,
		},
		{
			input:       "'hi' == 'hi'",
			expectType:  "bool",
			expectValue: true,
		},
		{
			input:       "1 != 2",
			expectType:  "bool",
			expectValue: true,
		},
		{
			input:       "10 > 0",
			expectType:  "bool",
			expectValue: true,
		},
		{
			input:       "'2014-01-02' > '2014-01-01 23:59:59'",
			expectType:  "bool",
			expectValue: true,
		},
		{
			input:       "1 << 10",
			expectType:  "float64",
			expectValue: float64(1024),
		},
		{
			input:       "2 > 1 && 4 > 2",
			expectType:  "bool",
			expectValue: true,
		},
		{
			input:       "2 > 1 ? 4 : 2",
			expectType:  "float64",
			expectValue: float64(4),
		},
		{
			input:       "12 & 10",
			expectType:  "float64",
			expectValue: float64(8),
		},
		{
			input:       "12 | 10",
			expectType:  "float64",
			expectValue: float64(14),
		},
		{
			input:       "'hi' in (1,2,3,'hi')",
			expectType:  "bool",
			expectValue: true,
		},
		{
			input:       "('h' + 'i') in (1,2,3,'hi')",
			expectType:  "bool",
			expectValue: true,
		},
	}
	for _, tcase := range testCases {
		lexer := NewGovaluateLexer(antlr.NewInputStream(tcase.input))

		stream := antlr.NewCommonTokenStream(lexer, antlr.LexerDefaultTokenChannel)

		parser := NewGovaluateParser(stream)

		expression := parser.Expression()

		ast := NewASTEvaluator()
		result := ast.Visit(expression)

		fmt.Println(result)
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
				t.Fatal("expect value is not equal")
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
			fmt.Println(t)
		}

	}

}

func TestOperatorParseWithParams(t *testing.T) {
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
			input:       "[http-response_body] == 'service is ok'",
			expectType:  "bool",
			expectValue: true,
			params:      map[string]any{"http-response_body": "service is ok"},
		},
		{
			input:       "${http-response_body} == 'service is ok'",
			expectType:  "bool",
			expectValue: true,
			params:      map[string]any{"http-response_body": "service is ok"},
		},
		{
			input:       "arr[1]",
			expectType:  "float64",
			expectValue: float64(2),
			params:      map[string]any{"arr": []float64{1, 2, 3, 4}},
		},
	}
	for _, tcase := range testCases {
		lexer := NewGovaluateLexer(antlr.NewInputStream(tcase.input))

		stream := antlr.NewCommonTokenStream(lexer, antlr.LexerDefaultTokenChannel)

		parser := NewGovaluateParser(stream)

		expression := parser.Expression()

		ast := NewASTEvaluatorWithParams(tcase.params, nil)
		result := ast.Visit(expression)

		fmt.Println(result)
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
				t.Fatal("expect value is not equal")
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
			fmt.Println(t)
		}

	}

}
