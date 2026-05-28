package parser

import (
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/bruceding/go-antlr-valuate/utils"
)

// PreCompiled holds parse-time, read-only artifacts that are safe to share
// across many Evaluate calls (and goroutines, given the maps are not mutated
// after construction).
//
// By moving literal parsing, identifier name normalization and function-map
// merging out of the hot Evaluate path, the evaluator only does the work
// proportional to the actual run-time inputs (parameters), not the
// expression structure.
type PreCompiled struct {
	// Literals caches the decoded value of each literal node (and any
	// constant-folded expression node) so the visitor avoids re-parsing or
	// re-evaluating purely-static subtrees on every call. Keys may be
	// *FloatContext / *StringContext / *BoolContext / *Array_valueContext
	// (leaf literals from walkPreCompile) or *ExpressionContext (folded
	// expressions from foldConstants).
	Literals map[antlr.ParserRuleContext]interface{}

	// VarNames caches the trimmed variable name for every identifier node
	// (e.g. "${foo}" -> "foo"), avoiding strings.Trim allocation per Visit.
	VarNames map[*IdentifierContext]string

	// Functions is the merged user + built-in function map. Shared across
	// evaluators; never mutated after PreCompile returns.
	Functions map[string]ExpressionFunction

	// ExprChildren caches AllExpression() per ExpressionContext so the
	// evaluator does not rebuild a []IExpressionContext slice on every
	// Evaluate call. The slice is read-only after PreCompile returns.
	ExprChildren map[*ExpressionContext][]IExpressionContext

	// ExprListChildren is the AllExpression() cache for ExpressionListContext
	// (function-call arg lists).
	ExprListChildren map[*ExpressionListContext][]IExpressionContext

	// ResolvedFuncs caches the function pointer for each FunctionCallContext
	// so VisitFunctionCall avoids the IDENTIFIER().GetText() and map lookup
	// on every call. Missing entries mean the function name was unknown at
	// parse time, in which case the visitor returns the usual error.
	ResolvedFuncs map[*FunctionCallContext]ExpressionFunction

	// ResolvedFuncNames is the parsed function name per call site, used for
	// error messages when ResolvedFuncs has no entry.
	ResolvedFuncNames map[*FunctionCallContext]string
}

// PreCompile walks the parse tree once and pre-computes every immutable
// artifact the evaluator would otherwise recompute on each Evaluate call.
func PreCompile(tree antlr.Tree, functions map[string]ExpressionFunction) *PreCompiled {
	pre := &PreCompiled{
		Literals:          make(map[antlr.ParserRuleContext]interface{}),
		VarNames:          make(map[*IdentifierContext]string),
		Functions:         functions,
		ExprChildren:      make(map[*ExpressionContext][]IExpressionContext),
		ExprListChildren:  make(map[*ExpressionListContext][]IExpressionContext),
		ResolvedFuncs:     make(map[*FunctionCallContext]ExpressionFunction),
		ResolvedFuncNames: make(map[*FunctionCallContext]string),
	}
	walkPreCompile(tree, pre)
	foldConstants(tree, pre)
	return pre
}

func walkPreCompile(node antlr.Tree, pre *PreCompiled) {
	if node == nil {
		return
	}

	switch ctx := node.(type) {
	case *FloatContext:
		if f, err := strconv.ParseFloat(ctx.GetText(), 64); err == nil {
			pre.Literals[ctx] = f
		}
	case *StringContext:
		s := strings.Trim(ctx.GetText(), "'")
		if t, ok := utils.TryParseTime(s); ok {
			pre.Literals[ctx] = t
		} else {
			pre.Literals[ctx] = s
		}
	case *BoolContext:
		pre.Literals[ctx] = ctx.GetText() == "true"
	case *IdentifierContext:
		pre.VarNames[ctx] = strings.Trim(ctx.GetText(), "${}")
	case *Array_valueContext:
		// Array_value also wraps literal tokens; pre-decode them too.
		if ctx.STRING_LITERAL() != nil {
			pre.Literals[ctx] = strings.Trim(ctx.GetText(), "'")
		} else if ctx.FLOAT_LITERAL() != nil {
			if f, err := strconv.ParseFloat(ctx.GetText(), 64); err == nil {
				pre.Literals[ctx] = f
			}
		} else if ctx.BOOL_LITERAL() != nil {
			pre.Literals[ctx] = ctx.GetText() == "true"
		}
	case *ExpressionContext:
		// Cache child expressions so the evaluator does not need to call the
		// generated AllExpression() per Evaluate (which allocates a fresh
		// []IExpressionContext slice each time).
		pre.ExprChildren[ctx] = ctx.AllExpression()
	case *ExpressionListContext:
		pre.ExprListChildren[ctx] = ctx.AllExpression()
	case *FunctionCallContext:
		name := ctx.IDENTIFIER().GetText()
		pre.ResolvedFuncNames[ctx] = name
		if f, ok := pre.Functions[name]; ok {
			pre.ResolvedFuncs[ctx] = f
		}
	}

	n := node.GetChildCount()
	for i := 0; i < n; i++ {
		walkPreCompile(node.GetChild(i), pre)
	}
}

// foldConstants performs constant folding on the parse tree. Any
// ExpressionContext whose entire subtree is composed of literals (no
// identifiers, no function calls, no member access) is evaluated once and
// the result stored in pre.Literals so the visitor's fast path returns it
// without walking the subtree at all on every Evaluate call.
//
// foldConstants must run after walkPreCompile because it relies on the
// ExprChildren / Literals caches populated there, and uses the visitor
// itself to compute folded values.
func foldConstants(tree antlr.Tree, pre *PreCompiled) {
	eval := NewASTEvaluatorShared(nil, pre)
	foldRec(tree, pre, eval)
}

// foldRec returns true iff the subtree rooted at node is "pure" — it
// contains no identifiers, no function calls and no operators that depend
// on runtime data (member access, indexing, assignment). For every pure
// *ExpressionContext encountered, the evaluator is invoked once to compute
// its value and the result is recorded in pre.Literals.
func foldRec(node antlr.Tree, pre *PreCompiled, eval *ASTEvaluator) bool {
	if node == nil {
		return false
	}
	switch ctx := node.(type) {
	case *FloatContext, *StringContext, *BoolContext, *Array_valueContext:
		return true
	case *IdentifierContext:
		return false
	case *FunctionCallContext:
		// We do not fold function calls — user-supplied functions may have
		// side effects. Recurse so any constant arguments inside still get
		// folded for marginal benefit, but this subtree is not pure.
		recurseChildren(node, pre, eval)
		return false
	case *ArraysContext, *ArrayContext:
		// Arrays are skipped: VisitArray returns a fresh []any per call to
		// avoid sharing a mutable slice between evaluations, so caching the
		// folded value would be unsafe.
		recurseChildren(node, pre, eval)
		return false
	case *ExpressionContext:
		if ctx.bop != nil {
			switch ctx.bop.GetText() {
			case ".", "[", "=", "+=", "-=", "*=", "/=", "&=", "|=", "^=",
				">>=", ">>>=", "<<=", "%=":
				recurseChildren(node, pre, eval)
				return false
			}
		}
		allPure := true
		if p := ctx.Primary(); p != nil {
			if !foldRec(p, pre, eval) {
				allPure = false
			}
		}
		if fc := ctx.FunctionCall(); fc != nil {
			foldRec(fc, pre, eval)
			allPure = false
		}
		for _, e := range pre.ExprChildren[ctx] {
			if !foldRec(e, pre, eval) {
				allPure = false
			}
		}
		if !allPure {
			return false
		}
		// Subtree is pure → safely evaluate once. Any panic during
		// evaluation (type mismatch on mixed-literal arithmetic, etc.) is
		// treated as "not foldable" — the runtime path will surface the
		// real error to the caller.
		func() {
			defer func() { _ = recover() }()
			val := eval.Visit(ctx)
			if _, isErr := val.(error); !isErr {
				pre.Literals[ctx] = val
			}
		}()
		return true
	}
	recurseChildren(node, pre, eval)
	return false
}

func recurseChildren(node antlr.Tree, pre *PreCompiled, eval *ASTEvaluator) {
	n := node.GetChildCount()
	for i := 0; i < n; i++ {
		foldRec(node.GetChild(i), pre, eval)
	}
}
