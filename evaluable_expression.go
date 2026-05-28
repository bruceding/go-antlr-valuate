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

	expressionContext, err := parseExpression(expr)
	if err != nil {
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

// parseExpression performs the standard ANTLR two-stage parsing strategy:
//
//  1. Stage 1 uses the fast SLL prediction mode with an error listener
//     attached. SLL is significantly faster than full LL but is not
//     guaranteed to match LL's decision on ambiguous inputs; if anything
//     goes wrong the listener captures it and we explicitly retry with LL,
//     instead of silently returning a wrong parse tree.
//
//     We deliberately do NOT use antlr.BailErrorStrategy here: in
//     antlr4-go v4.13.0 it panics from DefaultErrorStrategy.ReportError
//     because ParseCancellationException.GetMessage is unimplemented.
//
//  2. Stage 2 only runs when stage 1 reports an error. It rebuilds a fresh
//     parser using the default LL prediction mode and DefaultErrorStrategy,
//     and attaches our error listener so the caller still receives precise
//     diagnostics for genuine syntax errors.
//
// For the overwhelmingly common case of unambiguous, well-formed expressions
// stage 1 succeeds and stage 2 is skipped, so we keep the SLL speed-up
// without sacrificing correctness on ambiguous or malformed inputs.
func parseExpression(expr string) (parser.IExpressionContext, error) {
	// Stage 1: SLL prediction with an error listener that records but does
	// not abort. If SLL hits a conflict / syntax error we discard the
	// (potentially wrong) tree and retry with full LL.
	lexer := parser.NewGovaluateLexer(antlr.NewInputStream(expr))
	stream := antlr.NewCommonTokenStream(lexer, antlr.LexerDefaultTokenChannel)
	p := parser.NewGovaluateParser(stream)
	p.GetInterpreter().SetPredictionMode(antlr.PredictionModeSLL)
	p.RemoveErrorListeners()
	sllListener := NewEvaluableStatementErrorListener()
	p.AddErrorListener(sllListener)

	ctx := p.Expression()
	if sllListener.Error() == nil && !p.HasError() {
		return ctx, nil
	}

	// Stage 2: rebuild with full LL + default error strategy + our listener.
	lexer = parser.NewGovaluateLexer(antlr.NewInputStream(expr))
	stream = antlr.NewCommonTokenStream(lexer, antlr.LexerDefaultTokenChannel)
	p = parser.NewGovaluateParser(stream)
	p.GetInterpreter().SetPredictionMode(antlr.PredictionModeLL)
	p.RemoveErrorListeners()
	errorListener := NewEvaluableStatementErrorListener()
	p.AddErrorListener(errorListener)

	ctx = p.Expression()
	if err := errorListener.Error(); err != nil {
		return nil, err
	}
	return ctx, nil
}

func (e *EvaluableExpression) Evaluate(params map[string]interface{}) (interface{}, error) {
	var ast *parser.ASTEvaluator
	if v := e.pool.Get(); v != nil {
		ast = v.(*parser.ASTEvaluator)
		ast.Reset(params, e.pre)
	} else {
		ast = parser.NewASTEvaluatorShared(params, e.pre)
	}

	// Use defer so the evaluator is still returned to the pool if a
	// user-supplied function (or anything else on the visit path) panics.
	// We also clear the references the evaluator holds so the pooled object
	// does not pin caller-owned parameter maps or pre-compiled data.
	defer func() {
		ast.Reset(nil, nil)
		e.pool.Put(ast)
	}()

	result := ast.Visit(e.expressionContext)

	if err, ok := result.(error); ok {
		return nil, err
	}
	return result, nil
}
