package valuate

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/bruceding/go-antlr-valuate/parser"
)

type EvaluableExpression struct {
	//ast        *parser.ASTEvaluator
	expressionContext parser.IExpressionContext
	customFunctions   map[string]parser.ExpressionFunction
}

func NewEvaluableExpression(expr string) (*EvaluableExpression, error) {
	m := make(map[string]parser.ExpressionFunction)
	return NewEvaluableExpressionWithFunctions(expr, m)
}
func NewEvaluableExpressionWithFunctions(expr string, functions map[string]parser.ExpressionFunction) (*EvaluableExpression, error) {

	lexer := parser.NewGovaluateLexer(antlr.NewInputStream(expr))

	stream := antlr.NewCommonTokenStream(lexer, antlr.LexerDefaultTokenChannel)

	p := parser.NewGovaluateParser(stream)

	p.RemoveErrorListeners()
	errorListener := NewEvaluableStatementErrorListener()
	p.AddErrorListener(errorListener)

	expressionContext := p.Expression()

	//ast := parser.NewASTEvaluator()

	return &EvaluableExpression{
		//ast:        ast,
		expressionContext: expressionContext,
		customFunctions:   functions,
	}, errorListener.Error()
}

func (e *EvaluableExpression) Evaluate(params map[string]interface{}) (interface{}, error) {
	ast := parser.NewASTEvaluatorWithParams(params, e.customFunctions)
	result := ast.Visit(e.expressionContext)
	if e, ok := result.(error); ok {
		return nil, e
	}

	return result, nil
}
