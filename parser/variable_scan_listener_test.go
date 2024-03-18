package parser

import (
	"testing"

	"github.com/antlr4-go/antlr/v4"
)

func TestScanVariable(t *testing.T) {
	testCases := []struct {
		input       string
		expectValue *Variable
	}{
		{
			input: "a = b;",
			expectValue: &Variable{
				Name:         "a",
				VariableType: UNKNOWN,
				Index:        -1,
			},
		},
		{
			input: "a = 'string';",
			expectValue: &Variable{
				Name:         "a",
				VariableType: UNKNOWN,
				Index:        -1,
			},
		},
		{
			input: "a[1] = 'string';",
			expectValue: &Variable{
				Name:         "a",
				VariableType: ARRAYORMAP,
				Index:        1,
			},
		},
		{
			input: "a['index'] = 'string';",
			expectValue: &Variable{
				Name:         "a",
				VariableType: ARRAYORMAP,
				Index:        -1,
				IndexnName:   "index",
			},
		},
		{
			input: "a[b] = 'string';",
			expectValue: &Variable{
				Name:               "a",
				VariableType:       ARRAYORMAP,
				Index:              -1,
				IndexnVariableName: "b",
			},
		},
	}
	for _, tcase := range testCases {
		lexer := NewGovaluateLexer(antlr.NewInputStream(tcase.input))

		stream := antlr.NewCommonTokenStream(lexer, antlr.LexerDefaultTokenChannel)

		parser := NewGovaluateParser(stream)

		prog := parser.Prog()

		scan := NewVariableScanListener()
		antlr.ParseTreeWalkerDefault.Walk(scan, prog)

		ctx := prog.BlockStatements().Statement(0).GetStatementExpression().Expression(0)

		expectVariable := scan.node2Variables[ctx]
		if expectVariable == nil {
			t.Fatal("expected variable empty")
		} else {
			if !expectVariable.Equal(tcase.expectValue) {
				t.Fatalf("expect value is not equal, expected:%s, got :%s", tcase.expectValue.String(), expectVariable.String())
			}
		}
	}

}
