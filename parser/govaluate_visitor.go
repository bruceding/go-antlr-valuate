// Code generated from Govaluate.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser // Govaluate

import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by GovaluateParser.
type GovaluateVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by GovaluateParser#expression.
	VisitExpression(ctx *ExpressionContext) interface{}

	// Visit a parse tree produced by GovaluateParser#expressionList.
	VisitExpressionList(ctx *ExpressionListContext) interface{}

	// Visit a parse tree produced by GovaluateParser#functionCall.
	VisitFunctionCall(ctx *FunctionCallContext) interface{}

	// Visit a parse tree produced by GovaluateParser#float.
	VisitFloat(ctx *FloatContext) interface{}

	// Visit a parse tree produced by GovaluateParser#string.
	VisitString(ctx *StringContext) interface{}

	// Visit a parse tree produced by GovaluateParser#bool.
	VisitBool(ctx *BoolContext) interface{}

	// Visit a parse tree produced by GovaluateParser#arrays.
	VisitArrays(ctx *ArraysContext) interface{}

	// Visit a parse tree produced by GovaluateParser#identifier.
	VisitIdentifier(ctx *IdentifierContext) interface{}

	// Visit a parse tree produced by GovaluateParser#array.
	VisitArray(ctx *ArrayContext) interface{}

	// Visit a parse tree produced by GovaluateParser#array_value.
	VisitArray_value(ctx *Array_valueContext) interface{}
}
