// Code generated from Govaluate.g4 by ANTLR 4.13.0. DO NOT EDIT.

package parser // Govaluate

import "github.com/antlr4-go/antlr/v4"

// GovaluateListener is a complete listener for a parse tree produced by GovaluateParser.
type GovaluateListener interface {
	antlr.ParseTreeListener

	// EnterExpression is called when entering the expression production.
	EnterExpression(c *ExpressionContext)

	// EnterExpressionList is called when entering the expressionList production.
	EnterExpressionList(c *ExpressionListContext)

	// EnterFunctionCall is called when entering the functionCall production.
	EnterFunctionCall(c *FunctionCallContext)

	// EnterFloat is called when entering the float production.
	EnterFloat(c *FloatContext)

	// EnterString is called when entering the string production.
	EnterString(c *StringContext)

	// EnterBool is called when entering the bool production.
	EnterBool(c *BoolContext)

	// EnterArrays is called when entering the arrays production.
	EnterArrays(c *ArraysContext)

	// EnterArray is called when entering the array production.
	EnterArray(c *ArrayContext)

	// EnterArray_value is called when entering the array_value production.
	EnterArray_value(c *Array_valueContext)

	// ExitExpression is called when exiting the expression production.
	ExitExpression(c *ExpressionContext)

	// ExitExpressionList is called when exiting the expressionList production.
	ExitExpressionList(c *ExpressionListContext)

	// ExitFunctionCall is called when exiting the functionCall production.
	ExitFunctionCall(c *FunctionCallContext)

	// ExitFloat is called when exiting the float production.
	ExitFloat(c *FloatContext)

	// ExitString is called when exiting the string production.
	ExitString(c *StringContext)

	// ExitBool is called when exiting the bool production.
	ExitBool(c *BoolContext)

	// ExitArrays is called when exiting the arrays production.
	ExitArrays(c *ArraysContext)

	// ExitArray is called when exiting the array production.
	ExitArray(c *ArrayContext)

	// ExitArray_value is called when exiting the array_value production.
	ExitArray_value(c *Array_valueContext)
}
