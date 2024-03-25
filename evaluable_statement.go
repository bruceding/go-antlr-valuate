package valuate

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/bruceding/go-antlr-valuate/parser"
)

type EvaluableStatement struct {
	variableScanListener *parser.VariableScanListener
	progContext          parser.IProgContext
	customFunctions      map[string]parser.ExpressionFunction
}

func NewEvaluableStatement(expr string) (*EvaluableStatement, error) {
	m := make(map[string]parser.ExpressionFunction)
	return NewEvaluableStatementWithFunctions(expr, m)
}
func NewEvaluableStatementWithFunctions(expr string, functions map[string]parser.ExpressionFunction) (*EvaluableStatement, error) {

	lexer := parser.NewGovaluateLexer(antlr.NewInputStream(expr))

	stream := antlr.NewCommonTokenStream(lexer, antlr.LexerDefaultTokenChannel)

	p := parser.NewGovaluateParser(stream)

	progContext := p.Prog()

	scan := parser.NewVariableScanListener()
	antlr.ParseTreeWalkerDefault.Walk(scan, progContext)

	return &EvaluableStatement{
		variableScanListener: scan,
		progContext:          progContext,
		customFunctions:      functions,
	}, nil
}

func (e *EvaluableStatement) Evaluate(params map[string]interface{}) (map[string]any, error) {
	ast := parser.NewStatementASTEvaluatorWithParams(params, e.customFunctions, e.variableScanListener.GetNode2Variables())
	ast.Visit(e.progContext)

	return ast.ParamsMap(), nil
}
