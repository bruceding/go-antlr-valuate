package parser

import (
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
	}
	for _, tcase := range testCases {
		lexer := NewGovaluateLexer(antlr.NewInputStream(tcase.input))

		stream := antlr.NewCommonTokenStream(lexer, antlr.LexerDefaultTokenChannel)

		parser := NewGovaluateParser(stream)

		prog := parser.Prog()

		scan := NewVariableScanListener()
		antlr.ParseTreeWalkerDefault.Walk(scan, prog)
		paramMap := map[string]any{
			"b":      4,
			"values": []int{1, 2, 3},
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

		default:
			t.Fatalf("expect value is not equal, expected %v, %T, got %v, %T", expectVal, expectVal, ast.paramsMap[tcase.name], ast.paramsMap[tcase.name])
		}
	}

}
