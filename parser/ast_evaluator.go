package parser

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

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
	paramsMap  map[string]any
}

func NewASTEvaluator() *ASTEvaluator {
	return &ASTEvaluator{
		//BaseGovaluateVisitor: BaseGovaluateVisitor{
		//BaseParseTreeVisitor: &antlr.BaseParseTreeVisitor{},
		//},
		primaryMap: make(map[antlr.Parser]PrimaryValue),
		paramsMap:  make(map[string]any),
	}
}
func NewASTEvaluatorWithParams(p map[string]any) *ASTEvaluator {
	ast := NewASTEvaluator()
	for k, v := range p {
		switch val := v.(type) {
		case float32:
			ast.paramsMap[k] = float64(val)
		case float64:
			ast.paramsMap[k] = val
		case int:
			ast.paramsMap[k] = float64(val)
		case int8:
			ast.paramsMap[k] = float64(val)
		case int16:
			ast.paramsMap[k] = float64(val)
		case int32:
			ast.paramsMap[k] = float64(val)
		case int64:
			ast.paramsMap[k] = float64(val)
		case uint:
			ast.paramsMap[k] = float64(val)
		case uint8:
			ast.paramsMap[k] = float64(val)
		case uint16:
			ast.paramsMap[k] = float64(val)
		case uint32:
			ast.paramsMap[k] = float64(val)
		case uint64:
			ast.paramsMap[k] = float64(val)
		default:
			ast.paramsMap[k] = val
		}
	}
	return ast
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
				return fmt.Errorf("type:%T not support op:- ", val)
			}
		case "*":
			switch val := left.(type) {
			case float64:
				return val * right.(float64)
			default:
				return fmt.Errorf("type:%T not support op:* ", val)
			}
		case "/":
			switch val := left.(type) {
			case float64:
				if right.(float64) == 0 {
					return fmt.Errorf("right value:%v should not be zero ", right)
				}
				return val / right.(float64)
			default:
				return fmt.Errorf("type:%T not support op:/ ", val)
			}
		case "**":
			switch val := left.(type) {
			case float64:
				return math.Pow(val, right.(float64))
			default:
				return fmt.Errorf("type:%T not support op:** ", val)
			}
		case "^": // like **
			switch val := left.(type) {
			case float64:
				return math.Pow(val, right.(float64))
			default:
				return fmt.Errorf("type:%T not support op:^ ", val)
			}
		case "==":
			return reflect.DeepEqual(left, right)
		case "!=":
			return !reflect.DeepEqual(left, right)
		case "<=":
			switch val := left.(type) {
			case float64:
				return utils.LE[float64](val, right.(float64))
			case string:
				return utils.LE[string](val, right.(string))
			case time.Time:
				return utils.LE[int64](val.UnixMilli(), right.(time.Time).UnixMilli())
			default:
				return fmt.Errorf("type:%T not support op:%s ", val, ctx.bop.GetText())
			}
		case ">=":
			switch val := left.(type) {
			case float64:
				return utils.GE[float64](val, right.(float64))
			case string:
				return utils.GE[string](val, right.(string))
			case time.Time:
				return utils.GE[int64](val.UnixMilli(), right.(time.Time).UnixMilli())
			default:
				return fmt.Errorf("type:%T not support op:%s ", val, ctx.bop.GetText())
			}
		case ">":
			switch val := left.(type) {
			case float64:
				return utils.GT[float64](val, right.(float64))
			case string:
				return utils.GT[string](val, right.(string))
			case time.Time:
				return utils.GT[int64](val.UnixMilli(), right.(time.Time).UnixMilli())
			default:
				return fmt.Errorf("type:%T not support op:%s ", val, ctx.bop.GetText())
			}
		case "<":
			switch val := left.(type) {
			case float64:
				return utils.LT[float64](val, right.(float64))
			case string:
				return utils.LT[string](val, right.(string))
			case time.Time:
				return utils.LT[int64](val.UnixMilli(), right.(time.Time).UnixMilli())
			default:
				return fmt.Errorf("type:%T not support op:%s ", val, ctx.bop.GetText())
			}
		case "<<":
			switch val := left.(type) {
			case float64:
				return float64(int(val) << int(right.(float64)))
			default:
				return fmt.Errorf("type:%T not support op:%s ", val, ctx.bop.GetText())
			}
		case ">>":
			switch val := left.(type) {
			case float64:
				return float64(int(val) >> int(right.(float64)))
			default:
				return fmt.Errorf("type:%T not support op:%s ", val, ctx.bop.GetText())
			}
		case "&&":
			switch val := left.(type) {
			case bool:
				return val && right.(bool)
			default:
				return fmt.Errorf("type:%T not support op:%s ", val, ctx.bop.GetText())
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
func (v *ASTEvaluator) VisitIdentifier(ctx *IdentifierContext) interface{} {
	str := ctx.GetText()
	str = strings.Trim(str, "[]")
	if str != "" {
		if val, ok := v.paramsMap[str]; ok {
			return val
		}
	}
	return fmt.Errorf("param:%s not found", str)
}
