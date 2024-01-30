// Code generated from Govaluate.g4 by ANTLR 4.13.0. DO NOT EDIT.

package parser // Govaluate

import "github.com/antlr4-go/antlr/v4"

type BaseGovaluateVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseGovaluateVisitor) VisitExpression(ctx *ExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGovaluateVisitor) VisitExpressionList(ctx *ExpressionListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGovaluateVisitor) VisitFunctionCall(ctx *FunctionCallContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGovaluateVisitor) VisitFloat(ctx *FloatContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGovaluateVisitor) VisitString(ctx *StringContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGovaluateVisitor) VisitBool(ctx *BoolContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGovaluateVisitor) VisitArrays(ctx *ArraysContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGovaluateVisitor) VisitArray(ctx *ArrayContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseGovaluateVisitor) VisitArray_value(ctx *Array_valueContext) interface{} {
	return v.VisitChildren(ctx)
}