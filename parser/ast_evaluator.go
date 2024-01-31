package parser

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/bruceding/go-antlr-valuate/utils"
)

type PrimaryType int

const (
	STRING PrimaryType = iota + 1
	FLOAT
	BOOL
	ARRAY
	TIME
)

type PrimaryValue struct {
	t     PrimaryType
	value any
}

var _ GovaluateVisitor = &ASTEvaluator{}

type ASTEvaluator struct {
	BaseGovaluateVisitor
	primaryMap map[antlr.Parser]PrimaryValue
}

func NewASTEvaluator() *ASTEvaluator {
	return &ASTEvaluator{
		//BaseGovaluateVisitor: BaseGovaluateVisitor{
		//BaseParseTreeVisitor: &antlr.BaseParseTreeVisitor{},
		//},
		primaryMap: make(map[antlr.Parser]PrimaryValue),
	}
}
func (v *ASTEvaluator) Visit(tree antlr.ParseTree) interface{} { return tree.Accept(v) }

func (v *ASTEvaluator) VisitExpression(ctx *ExpressionContext) interface{} {
	if ctx.Primary() != nil {
		return v.Visit(ctx.Primary())
	}
	if len(ctx.AllExpression()) == 1 {
		if ctx.prefix != nil {
			switch ctx.prefix.GetText() {
			case "-":
				return (-1) * (v.Visit(ctx.Expression(0))).(float64)
			case "+":
				return v.Visit(ctx.Expression(0))
			case "!":
				return !v.Visit(ctx.Expression(0)).(bool)
			case "~":
				return float64(^int(v.Visit(ctx.Expression(0)).(float64)))
			case "++":
				return float64(1 + int(v.Visit(ctx.Expression(0)).(float64)))
			case "--":
				return float64(int(v.Visit(ctx.Expression(0)).(float64)) - 1)
			}
		}
		return v.Visit(ctx.Expression(0))
	} else if len(ctx.AllExpression()) == 2 {
		left := v.Visit(ctx.Expression(0))
		right := v.Visit(ctx.Expression(1))
		switch ctx.bop.GetText() {
		case "+":
			switch val := left.(type) {
			case string:
				return val + right.(string)
			case float64:
				return val + right.(float64)
			case []any:
				return append(val, right.([]any)...)
			default:
				return fmt.Errorf("type:%T not support op:+ ", val)
			}
		case "-":
			switch val := left.(type) {
			case float64:
				return val - right.(float64)
			default:
				return fmt.Errorf("type:%T not support op:+ ", val)
			}
		case "*":
			switch val := left.(type) {
			case float64:
				return val * right.(float64)
			default:
				return fmt.Errorf("type:%T not support op:+ ", val)
			}
		case "/":
			switch val := left.(type) {
			case float64:
				if right.(float64) == 0 {
					return fmt.Errorf("right value:%v should not be zero ", right)
				}
				return val / right.(float64)
			default:
				return fmt.Errorf("type:%T not support op:+ ", val)
			}
		case "**":
			switch val := left.(type) {
			case float64:
				return math.Pow(val, right.(float64))
			default:
				return fmt.Errorf("type:%T not support op:+ ", val)
			}

		default:
			return fmt.Errorf("op:%s not support", ctx.bop.GetText())
		}
	}
	return v.VisitChildren(ctx)
}
func (v *ASTEvaluator) VisitFloat(ctx *FloatContext) interface{} {

	str := ctx.GetText()
	if f, err := strconv.ParseFloat(str, 64); err == nil {
		return f
	}

	return nil
}

func (v *ASTEvaluator) VisitString(ctx *StringContext) interface{} {
	str := ctx.GetText()
	str = strings.Trim(str, "'")
	if t, ok := utils.TryParseTime(str); ok {
		return t
	}
	return str
}

func (v *ASTEvaluator) VisitBool(ctx *BoolContext) interface{} {
	str := ctx.GetText()
	return str == "true"
}

func (v *ASTEvaluator) VisitArrays(ctx *ArraysContext) interface{} {
	return v.Visit(ctx.Array())
}

func (v *ASTEvaluator) VisitArray(ctx *ArrayContext) interface{} {
	arr := make([]any, 0, len(ctx.AllArray_value()))

	for i := 0; i < len(ctx.AllArray_value()); i++ {
		val := v.Visit(ctx.Array_value(i))
		arr = append(arr, val)
	}

	return arr
}

func (v *ASTEvaluator) VisitArray_value(ctx *Array_valueContext) interface{} {
	if ctx.STRING_LITERAL() != nil {
		str := ctx.GetText()
		str = strings.Trim(str, "'")
		return str
	} else if ctx.FLOAT_LITERAL() != nil {
		str := ctx.GetText()
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return f
		}
	} else if ctx.BOOL_LITERAL() != nil {
		return ctx.GetText() == "true"
	}
	/**
	if ctx.STRING_LITERAL().GetSymbol().GetTokenIndex() == GovaluateLexerSTRING_LITERAL {
		return ctx.STRING_LITERAL().GetText()
	}
	**/

	return ctx.GetText()
}
