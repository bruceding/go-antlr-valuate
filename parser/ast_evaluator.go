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

var _ GovaluateVisitor = &ASTEvaluator{}

type ASTEvaluator struct {
	BaseGovaluateVisitor
	paramsMap          map[string]any
	preDefineFunctions map[string]ExpressionFunction
}

func NewASTEvaluator() *ASTEvaluator {
	return NewASTEvaluatorWithParams(make(map[string]any), make(map[string]ExpressionFunction))
}
func NewASTEvaluatorWithParams(p map[string]any, functions map[string]ExpressionFunction) *ASTEvaluator {
	ast := &ASTEvaluator{
		paramsMap:          make(map[string]any),
		preDefineFunctions: make(map[string]ExpressionFunction),
	}

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
	if functions != nil {
		ast.preDefineFunctions = functions
		for k, v := range buildInFunctionsMap {
			ast.preDefineFunctions[k] = v
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
		case "%":
			switch val := left.(type) {
			case float64:
				if right.(float64) == 0 {
					return fmt.Errorf("right value:%v should not be zero ", right)
				}
				return math.Mod(val, right.(float64))
			default:
				return fmt.Errorf("type:%T not support op:%s ", val, ctx.bop.GetText())
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
		case "&":
			switch val := left.(type) {
			case float64:
				return float64(int(val) & int(right.(float64)))
			default:
				return fmt.Errorf("type:%T not support op:%s ", val, ctx.bop.GetText())
			}
		case "|":
			switch val := left.(type) {
			case float64:
				return float64(int(val) | int(right.(float64)))
			default:
				return fmt.Errorf("type:%T not support op:%s ", val, ctx.bop.GetText())
			}
		case "in":
			arr, ok := right.([]any)
			if !ok {
				return fmt.Errorf("right value:%v should be array", right)
			}
			switch val := left.(type) {
			case float64:
				return utils.InArray(val, arr)
			case string:
				return utils.InArray(val, arr)
			default:
				return fmt.Errorf("type:%T not support op:in array", val)
			}
		case "[":
			index, ok := right.(float64)
			if !ok {
				return fmt.Errorf("right value:%v should be number", right)
			}
			switch arr := left.(type) {
			case []any:
				if len(arr) <= int(index) {
					return fmt.Errorf("index:%d exceed array:%v length:%d", int(index), arr, len(arr))
				}
				return arr[int(index)]
			case []float64:
				if len(arr) <= int(index) {
					return fmt.Errorf("index:%d exceed array:%v length:%d", int(index), arr, len(arr))
				}
				return arr[int(index)]
			case []string:
				if len(arr) <= int(index) {
					return fmt.Errorf("index:%d exceed array:%v length:%d", int(index), arr, len(arr))
				}
				return arr[int(index)]
			case []int:
				if len(arr) <= int(index) {
					return fmt.Errorf("index:%d exceed array:%v length:%d", int(index), arr, len(arr))
				}
				return float64(arr[int(index)])
			case []int64:
				if len(arr) <= int(index) {
					return fmt.Errorf("index:%d exceed array:%v length:%d", int(index), arr, len(arr))
				}
				return float64(arr[int(index)])
			case []int32:
				if len(arr) <= int(index) {
					return fmt.Errorf("index:%d exceed array:%v length:%d", int(index), arr, len(arr))
				}
				return float64(arr[int(index)])
			case []float32:
				if len(arr) <= int(index) {
					return fmt.Errorf("index:%d exceed array:%v length:%d", int(index), arr, len(arr))
				}
				return float64(arr[int(index)])

			default:
				return fmt.Errorf("op:[] not support type:%T", left)

			}

		default:
			return fmt.Errorf("op:%s not support", ctx.bop.GetText())
		}
	} else if len(ctx.AllExpression()) == 3 { // exp ? exp : exp
		condition := v.Visit(ctx.Expression(0))
		left := v.Visit(ctx.Expression(1))
		right := v.Visit(ctx.Expression(2))
		switch ctx.bop.GetText() {
		case "?":
			if val, ok := condition.(bool); !ok {
				return fmt.Errorf("op:%s, condition:%v should be bool", ctx.bop.GetText(), condition)
			} else {
				if val {
					return left
				} else {
					return right
				}
			}
		default:
			return fmt.Errorf("op:%s not support", ctx.bop.GetText())
		}
	} else if ctx.FunctionCall() != nil {
		return v.Visit(ctx.FunctionCall())
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
	str = strings.Trim(str, "[]${}")
	if str != "" {
		if val, ok := v.paramsMap[str]; ok {
			return val
		}
	}
	return fmt.Errorf("param:%s not found", str)
}
func (v *ASTEvaluator) VisitFunctionCall(ctx *FunctionCallContext) interface{} {
	funcName := ctx.IDENTIFIER().GetText()
	params := v.Visit(ctx.ExpressionList())

	if f, exist := v.preDefineFunctions[funcName]; exist {
		if v, err := f((params.([]any))...); err != nil {
			return err
		} else {
			return v
		}
	}

	return fmt.Errorf("not found function:%s", funcName)
}

func (v *ASTEvaluator) VisitExpressionList(ctx *ExpressionListContext) interface{} {
	var params []any
	for _, exprContext := range ctx.AllExpression() {
		params = append(params, v.Visit(exprContext))
	}

	return params
}
