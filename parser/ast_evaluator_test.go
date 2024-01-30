package parser

import (
	"fmt"
	"testing"

	"github.com/antlr4-go/antlr/v4"
)

func TestParse(t *testing.T) {
	input := "(1,2,3)"

	lexer := NewGovaluateLexer(antlr.NewInputStream(input))

	stream := antlr.NewCommonTokenStream(lexer, antlr.LexerDefaultTokenChannel)

	parser := NewGovaluateParser(stream)

	expression := parser.Expression()

	ast := NewASTEvaluator()

	result := ast.Visit(expression)

	fmt.Println(result)
}
