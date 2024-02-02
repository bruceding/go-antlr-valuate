// Code generated from Govaluate.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser // Govaluate

import "github.com/antlr4-go/antlr/v4"

// BaseGovaluateListener is a complete listener for a parse tree produced by GovaluateParser.
type BaseGovaluateListener struct{}

var _ GovaluateListener = &BaseGovaluateListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseGovaluateListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseGovaluateListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseGovaluateListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseGovaluateListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterExpression is called when production expression is entered.
func (s *BaseGovaluateListener) EnterExpression(ctx *ExpressionContext) {}

// ExitExpression is called when production expression is exited.
func (s *BaseGovaluateListener) ExitExpression(ctx *ExpressionContext) {}

// EnterExpressionList is called when production expressionList is entered.
func (s *BaseGovaluateListener) EnterExpressionList(ctx *ExpressionListContext) {}

// ExitExpressionList is called when production expressionList is exited.
func (s *BaseGovaluateListener) ExitExpressionList(ctx *ExpressionListContext) {}

// EnterFunctionCall is called when production functionCall is entered.
func (s *BaseGovaluateListener) EnterFunctionCall(ctx *FunctionCallContext) {}

// ExitFunctionCall is called when production functionCall is exited.
func (s *BaseGovaluateListener) ExitFunctionCall(ctx *FunctionCallContext) {}

// EnterFloat is called when production float is entered.
func (s *BaseGovaluateListener) EnterFloat(ctx *FloatContext) {}

// ExitFloat is called when production float is exited.
func (s *BaseGovaluateListener) ExitFloat(ctx *FloatContext) {}

// EnterString is called when production string is entered.
func (s *BaseGovaluateListener) EnterString(ctx *StringContext) {}

// ExitString is called when production string is exited.
func (s *BaseGovaluateListener) ExitString(ctx *StringContext) {}

// EnterBool is called when production bool is entered.
func (s *BaseGovaluateListener) EnterBool(ctx *BoolContext) {}

// ExitBool is called when production bool is exited.
func (s *BaseGovaluateListener) ExitBool(ctx *BoolContext) {}

// EnterArrays is called when production arrays is entered.
func (s *BaseGovaluateListener) EnterArrays(ctx *ArraysContext) {}

// ExitArrays is called when production arrays is exited.
func (s *BaseGovaluateListener) ExitArrays(ctx *ArraysContext) {}

// EnterIdentifier is called when production identifier is entered.
func (s *BaseGovaluateListener) EnterIdentifier(ctx *IdentifierContext) {}

// ExitIdentifier is called when production identifier is exited.
func (s *BaseGovaluateListener) ExitIdentifier(ctx *IdentifierContext) {}

// EnterArray is called when production array is entered.
func (s *BaseGovaluateListener) EnterArray(ctx *ArrayContext) {}

// ExitArray is called when production array is exited.
func (s *BaseGovaluateListener) ExitArray(ctx *ArrayContext) {}

// EnterArray_value is called when production array_value is entered.
func (s *BaseGovaluateListener) EnterArray_value(ctx *Array_valueContext) {}

// ExitArray_value is called when production array_value is exited.
func (s *BaseGovaluateListener) ExitArray_value(ctx *Array_valueContext) {}
