package valuate

import (
	"sync"

	"github.com/antlr4-go/antlr/v4"
	"github.com/bruceding/go-antlr-valuate/parser"
)

// EvaluableExpression represents a parsed and pre-compiled expression that can
// be evaluated many times. It is safe to call Evaluate concurrently from
// multiple goroutines as long as the parameter map passed in is not mutated
// during the call.
type EvaluableExpression struct {
	expressionContext parser.IExpressionContext
	customFunctions   map[string]parser.ExpressionFunction
	// pre holds parse-time caches (literals / variable names / merged
	// function map) that the evaluator reads on every call.
	pre *parser.PreCompiled
	// pool reuses ASTEvaluator instances across Evaluate calls to keep the
	// hot path allocation-free for the evaluator object itself.
	pool sync.Pool
}

func NewEvaluableExpression(expr string) (*EvaluableExpression, error) {
	return NewEvaluableExpressionWithFunctions(expr, nil)
}

func NewEvaluableExpressionWithFunctions(expr string, functions map[string]parser.ExpressionFunction) (*EvaluableExpression, error) {

	lexer := parser.NewGovaluateLexer(antlr.NewInputStream(expr))

	stream := antlr.NewCommonTokenStream(lexer, antlr.LexerDefaultTokenChannel)

	p := parser.NewGovaluateParser(stream)
	// SLL prediction mode is significantly faster for the unambiguous
	// expression grammar used here. ANTLR will fall back to LL automatically
	// if a parse conflict is detected.
	p.GetInterpreter().SetPredictionMode(antlr.PredictionModeSLL)

	p.RemoveErrorListeners()
	errorListener := NewEvaluableStatementErrorListener()
	p.AddErrorListener(errorListener)

	expressionContext := p.Expression()

	if err := errorListener.Error(); err != nil {
		return nil, err
	}

	// Merge built-in and user functions exactly once so the evaluator does
	// not have to copy them on every Evaluate call. Built-in functions take
	// precedence over user-supplied functions with the same name, matching
	// the prior behaviour of NewASTEvaluatorWithParams (where built-ins were
	// written into the merged map after user functions and therefore won on
	// collisions).
	merged := make(map[string]parser.ExpressionFunction, len(functions)+len(parser.BuildInFunctionsMap))
	for k, v := range functions {
		merged[k] = v
	}
	for k, v := range parser.BuildInFunctionsMap {
		merged[k] = v
	}

	pre := parser.PreCompile(expressionContext, merged)

	return &EvaluableExpression{
		expressionContext: expressionContext,
		customFunctions:   functions,
		pre:               pre,
	}, nil
}

func (e *EvaluableExpression) Evaluate(params map[string]interface{}) (interface{}, error) {
	var ast *parser.ASTEvaluator
	if v := e.pool.Get(); v != nil {
		ast = v.(*parser.ASTEvaluator)
		ast.Reset(params, e.pre)
	} else {
		ast = parser.NewASTEvaluatorShared(params, e.pre)
	}

	result := ast.Visit(e.expressionContext)

	// Clear the references the evaluator holds before returning it to the
	// pool so the pooled object does not pin caller-owned data.
	ast.Reset(nil, nil)
	e.pool.Put(ast)

	if err, ok := result.(error); ok {
		return nil, err
	}

	return result, nil
}
