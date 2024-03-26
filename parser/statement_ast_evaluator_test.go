package parser

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/antlr4-go/antlr/v4"
)

func TestStatement(t *testing.T) {
	testCases := []struct {
		input       string
		expectValue any
		name        string
	}{
		{
			input:       "a = 1;",
			expectValue: float64(1),
			name:        "a",
		},
		{
			input:       "a = b;",
			expectValue: float64(4),
			name:        "a",
		},
		{
			input:       " c =  4;\n a = c;",
			expectValue: float64(4),
			name:        "a",
		},
		{
			input:       " values[0] = b;",
			expectValue: []int{4, 2, 3},
			name:        "values",
		},
		{
			input:       "output_values = (1,2,3);\n values[0] = b;\n output_values[0] = values[0];",
			expectValue: []any{float64(4), float64(2), float64(3)},
			name:        "output_values",
		},
		{
			input:       "sum = 0;\n sum += b; sum +=b;",
			expectValue: float64(8),
			name:        "sum",
		},
		{
			input:       "values[0] +=b;",
			expectValue: []int{5, 2, 3},
			name:        "values",
		},
		{
			input:       "c = values[2] - values[0];",
			expectValue: float64(2),
			name:        "c",
		},
		{
			input:       "c = b;\n c -= values[1];",
			expectValue: float64(2),
			name:        "c",
		},
		{
			input:       "a = 8;\n  a /= b;",
			expectValue: float64(2),
			name:        "a",
		},
		{
			input:       "sum = 0;\n  a = ++sum;",
			expectValue: float64(1),
			name:        "a",
		},
		{
			input:       "sum = 0;\n  a = ++sum;",
			expectValue: float64(1),
			name:        "sum",
		},
		{
			input:       "sum = 4;\n  a = --sum;",
			expectValue: float64(3),
			name:        "sum",
		},
		{
			input:       "sum = 4;\n  a = sum--;",
			expectValue: float64(4),
			name:        "a",
		},
		{
			input:       "sum = 4;\n  a = sum--;",
			expectValue: float64(3),
			name:        "sum",
		},
		{
			input: `sum = 0;
			for (i = 0;  i < 3; i++) {
				sum += values[i];
			}`,
			expectValue: float64(6),
			name:        "sum",
		},
		{
			input: `a = 0 ;
			if (b > 2) {
				a = 1;
			} else {
				a = 2;
			}`,
			expectValue: float64(1),
			name:        "a",
		},
		{
			input: `a = 0 ;
			if (b > 8) {
				a = 1;
			} else {
				a = 2;
			}`,
			expectValue: float64(2),
			name:        "a",
		},
		{
			input: `a = 0 ;
			if (b > 8) {
				a = 1;
			}`,
			expectValue: float64(0),
			name:        "a",
		},
		{
			input: `a = 0 ;
			if (b > 8) {
				a = 1;
			} else if (b == 4){
					a = 3;
			} else {
				a = 2;
			}`,
			expectValue: float64(3),
			name:        "a",
		},
		{
			input: `out_values = (0,0,0);
			for (i = 0;  i < 3; i++) {
				out_values[i] = values[i];
			}`,
			expectValue: []any{float64(1), float64(2), float64(3)},
			name:        "out_values",
		},
		{
			input: `
			foreach (input_values_map as k => v) {
				out_values_map[k] = v;
			}
			`,
			expectValue: map[string]any{"1": 1, "2": 2, "3": 3},
			name:        "out_values_map",
		},
		{
			input: `
			sum = 0;
			foreach (values as k => v) {
				sum += v; 
			}
			`,
			expectValue: float64(6),
			name:        "sum",
		},
	}
	for _, tcase := range testCases {
		lexer := NewGovaluateLexer(antlr.NewInputStream(tcase.input))

		stream := antlr.NewCommonTokenStream(lexer, antlr.LexerDefaultTokenChannel)

		parser := NewGovaluateParser(stream)

		prog := parser.Prog()

		scan := NewVariableScanListener()
		antlr.ParseTreeWalkerDefault.Walk(scan, prog)
		paramMap := map[string]any{
			"b":                4,
			"values":           []int{1, 2, 3},
			"input_values_map": map[string]any{"1": 1, "2": 2, "3": 3},
			"out_values_map":   map[string]any{},
		}

		ast := NewStatementASTEvaluatorWithParams(paramMap, make(map[string]ExpressionFunction), scan.node2Variables)
		ast.Visit(prog)

		switch expectVal := tcase.expectValue.(type) {
		case float64:
			if expectVal != ast.paramsMap[tcase.name] {
				t.Fatalf("expect value is not equal, expected %v, got %v", expectVal, ast.paramsMap[tcase.name])
			}
		case []int:
			if !reflect.DeepEqual(expectVal, ast.paramsMap[tcase.name]) {
				t.Fatalf("expect value is not equal, expected %v, %T, got %v, %T", expectVal, expectVal, ast.paramsMap[tcase.name], ast.paramsMap[tcase.name])
			}
		case []any:
			if !reflect.DeepEqual(expectVal, ast.paramsMap[tcase.name]) {
				t.Fatalf("expect value is not equal, expected %v, %T, got %v, %T", expectVal, expectVal, ast.paramsMap[tcase.name], ast.paramsMap[tcase.name])
			}
		case map[string]any:
			if !reflect.DeepEqual(expectVal, ast.paramsMap[tcase.name]) {
				t.Fatalf("expect value is not equal, expected %v, %T, got %v, %T", expectVal, expectVal, ast.paramsMap[tcase.name], ast.paramsMap[tcase.name])
			}

		default:
			t.Fatalf("expect value is not equal, expected %v, %T, got %v, %T", expectVal, expectVal, ast.paramsMap[tcase.name], ast.paramsMap[tcase.name])
		}
	}

}

func TestStatement2(t *testing.T) {
	testCases := []struct {
		input       string
		expectValue any
		name        string
	}{
		{
			input: `out_values = (0,0,0);
			for (i = 0;  i < len(values); i++) {
				out_values[i] = values[i];
			}`,
			expectValue: []any{float64(1), float64(2), float64(3)},
			name:        "out_values",
		},
	}
	for _, tcase := range testCases {
		lexer := NewGovaluateLexer(antlr.NewInputStream(tcase.input))

		stream := antlr.NewCommonTokenStream(lexer, antlr.LexerDefaultTokenChannel)

		parser := NewGovaluateParser(stream)

		prog := parser.Prog()

		scan := NewVariableScanListener()
		antlr.ParseTreeWalkerDefault.Walk(scan, prog)
		paramMap := map[string]any{
			"b":                4,
			"values":           []int{1, 2, 3},
			"input_values_map": map[string]any{"1": 1, "2": 2, "3": 3},
			"out_values_map":   map[string]any{},
		}

		ast := NewStatementASTEvaluatorWithParams(paramMap, make(map[string]ExpressionFunction), scan.node2Variables)
		ast.Visit(prog)

		fmt.Println(ast.paramsMap)
		switch expectVal := tcase.expectValue.(type) {
		case float64:
			if expectVal != ast.paramsMap[tcase.name] {
				t.Fatalf("expect value is not equal, expected %v, got %v", expectVal, ast.paramsMap[tcase.name])
			}
		case []int:
			if !reflect.DeepEqual(expectVal, ast.paramsMap[tcase.name]) {
				t.Fatalf("expect value is not equal, expected %v, %T, got %v, %T", expectVal, expectVal, ast.paramsMap[tcase.name], ast.paramsMap[tcase.name])
			}
		case []any:
			if !reflect.DeepEqual(expectVal, ast.paramsMap[tcase.name]) {
				t.Fatalf("expect value is not equal, expected %v, %T, got %v, %T", expectVal, expectVal, ast.paramsMap[tcase.name], ast.paramsMap[tcase.name])
			}
		case map[string]any:
			if !reflect.DeepEqual(expectVal, ast.paramsMap[tcase.name]) {
				t.Fatalf("expect value is not equal, expected %v, %T, got %v, %T", expectVal, expectVal, ast.paramsMap[tcase.name], ast.paramsMap[tcase.name])
			}

		default:
			t.Fatalf("expect value is not equal, expected %v, %T, got %v, %T", expectVal, expectVal, ast.paramsMap[tcase.name], ast.paramsMap[tcase.name])
		}
	}

}
