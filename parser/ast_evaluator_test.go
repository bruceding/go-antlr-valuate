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
