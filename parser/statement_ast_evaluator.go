package parser

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/antlr4-go/antlr/v4"
	"github.com/bruceding/go-antlr-valuate/utils"
)

var _ GovaluateVisitor = &StatementASTEvaluator{}

type StatementASTEvaluator struct {
	BaseGovaluateVisitor
	paramsMap          map[string]any
	preDefineFunctions map[string]ExpressionFunction
	node2Variables     map[antlr.ParserRuleContext]*Variable
}

func NewStatementASTEvaluator(node2Variables map[antlr.ParserRuleContext]*Variable) *StatementASTEvaluator {
	return NewStatementASTEvaluatorWithParams(make(map[string]any), make(map[string]ExpressionFunction), node2Variables)
}
func NewStatementASTEvaluatorWithParams(p map[string]any, functions map[string]ExpressionFunction, node2Variables map[antlr.ParserRuleContext]*Variable) *StatementASTEvaluator {
	ast := &StatementASTEvaluator{
		paramsMap:          make(map[string]any),
		preDefineFunctions: make(map[string]ExpressionFunction),
		node2Variables:     node2Variables,
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
func (v *StatementASTEvaluator) Visit(tree antlr.ParseTree) interface{} { return tree.Accept(v) }

func (v *StatementASTEvaluator) VisitProg(ctx *ProgContext) interface{} {
	v.Visit(ctx.BlockStatements())

	return nil
}
func (v *StatementASTEvaluator) VisitBlock(ctx *BlockContext) interface{} {
	v.Visit(ctx.BlockStatements())
	return nil
}
func (v *StatementASTEvaluator) VisitBlockStatements(ctx *BlockStatementsContext) interface{} {
	for _, statment := range ctx.AllStatement() {
		v.Visit(statment)
	}

	return nil
}
func (v *StatementASTEvaluator) VisitStatement(ctx *StatementContext) interface{} {
	if ctx.statementExpression != nil {
		v.Visit(ctx.statementExpression)
	} else if ctx.FOR() != nil {
		if ctx.ForControl() != nil {
			if ctx.ForControl().ForInit() != nil {
				for _, expression := range ctx.ForControl().ForInit().ExpressionList().AllExpression() {
					v.Visit(expression)
				}
			}

			i := 0
			start := time.Now()
			for {
				if ctx.ForControl().Expression() != nil {
					flag := v.Visit(ctx.ForControl().Expression())
					if err, ok := flag.(error); ok {
						return err
					} else {
						if boolVal, ok := flag.(bool); ok {
							if !boolVal {
								break
							}
						} else {
							return fmt.Errorf("invalid for statement expression, result:%v, not true or false, expression:%s, at %d,%d", flag,
								ctx.ForControl().Expression().GetText(), ctx.ForControl().Expression().GetStart().GetLine(), ctx.ForControl().Expression().GetStart().GetColumn())
						}
					}
				}
				if ctx.Statement(0) != nil {
					v.Visit(ctx.Statement(0))
				}

				i++
				// for loop is long
				if i >= 10000 {
					// 5 seconds
					if time.Now().Unix()-start.Unix() > 5 {
						return errors.New("for loop excute too long , maybe contains an infinite loop")
					}
				}
				if ctx.ForControl().GetForUpdate() != nil {
					for _, expression := range ctx.ForControl().GetForUpdate().AllExpression() {
						v.Visit(expression)
					}
				}
			}

		}
	} else if ctx.blockLabel != nil {
		v.Visit(ctx.Block())
	} else if ctx.IF() != nil {
		flag := v.Visit(ctx.ParExpression().Expression())
		if err, ok := flag.(error); ok {
			return err
		} else {
			if bFlag, ok := flag.(bool); ok {
				if bFlag {
					v.Visit(ctx.Statement(0))
				} else {
					// if condition is false
					if ctx.Statement(1) != nil {
						v.Visit(ctx.Statement(1))
					}
				}
			} else {
				return fmt.Errorf("invalid for ParExpression, result:%v, not true or false, error expression:%s, at %d,%d", flag,
					ctx.ParExpression().Expression().GetText(), ctx.ParExpression().Expression().GetStart().GetLine(), ctx.ParExpression().Expression().GetStart().GetColumn())

			}
		}
	} else if ctx.FOREACH() != nil {
		eachVariable := v.node2Variables[ctx.ForeachControl().GetFromExpression()]
		if eachVariable == nil || eachVariable.Name == "" {
			return fmt.Errorf("left variable not found, expression:%s, at %d,%d", ctx.ForeachControl().GetFromExpression().GetText(),
				ctx.ForeachControl().GetFromExpression().GetStart().GetLine(), ctx.ForeachControl().GetFromExpression().GetStart().GetColumn())
		}
		eachParam, ok := v.paramsMap[eachVariable.Name]
		if !ok {
			return fmt.Errorf("left variable not found, expression:%s, at %d,%d", ctx.ForeachControl().GetFromExpression().GetText(),
				ctx.ForeachControl().GetFromExpression().GetStart().GetLine(), ctx.ForeachControl().GetFromExpression().GetStart().GetColumn())
		}
		eachParamType := reflect.TypeOf(eachParam)
		keyVariable := v.node2Variables[ctx.ForeachControl().GetKeyExpression()]
		valueVariable := v.node2Variables[ctx.ForeachControl().GetValueExpression()]
		if eachParamType.Kind() == reflect.Slice {
			switch eachParamValue := eachParam.(type) {
			case []any:
				for i, value := range eachParamValue {
					v.paramsMap[keyVariable.Name] = i
					v.paramsMap[valueVariable.Name] = value
					v.Visit(ctx.Statement(0))
				}
			case []string:
				for i, value := range eachParamValue {
					v.paramsMap[keyVariable.Name] = i
					v.paramsMap[valueVariable.Name] = value
					v.Visit(ctx.Statement(0))
				}
			case []int:
				for i, value := range eachParamValue {
					v.paramsMap[keyVariable.Name] = i
					v.paramsMap[valueVariable.Name] = value
					v.Visit(ctx.Statement(0))
				}
			case []float64:
				for i, value := range eachParamValue {
					v.paramsMap[keyVariable.Name] = i
					v.paramsMap[valueVariable.Name] = value
					v.Visit(ctx.Statement(0))
				}
			}

		} else if eachParamType.Kind() == reflect.Map {
			switch eachParamValue := eachParam.(type) {
			case map[string]any:
				for key, value := range eachParamValue {
					v.paramsMap[keyVariable.Name] = key
					v.paramsMap[valueVariable.Name] = value

					v.Visit(ctx.Statement(0))
				}
			case map[string]int:
				for key, value := range eachParamValue {
					v.paramsMap[keyVariable.Name] = key
					v.paramsMap[valueVariable.Name] = value

					v.Visit(ctx.Statement(0))
				}
			case map[string]float64:
				for key, value := range eachParamValue {
					v.paramsMap[keyVariable.Name] = key
					v.paramsMap[valueVariable.Name] = value

					v.Visit(ctx.Statement(0))
				}
			case map[string]string:
				for key, value := range eachParamValue {
					v.paramsMap[keyVariable.Name] = key
					v.paramsMap[valueVariable.Name] = value

					v.Visit(ctx.Statement(0))
				}
			}
		} else {
			return fmt.Errorf("foreach not support variable type:%T, expression:%s, at %d,%d", eachParam, ctx.ForeachControl().GetFromExpression().GetText(),
				ctx.ForeachControl().GetFromExpression().GetStart().GetLine(), ctx.ForeachControl().GetFromExpression().GetStart().GetColumn())

		}

	}
	return nil
}
func (v *StatementASTEvaluator) VisitExpression(ctx *ExpressionContext) interface{} {
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
			case "++", "--":
				leftVariable := v.node2Variables[ctx.Expression(0)]
				if leftVariable == nil || leftVariable.Name == "" {
					return fmt.Errorf("left variable not found")
				}
				paramValue, exist := v.paramsMap[leftVariable.Name]
				if !exist {
					return fmt.Errorf("variable %v not found", leftVariable.Name)
				} else {
					if leftVariable.VariableType == UNKNOWN {
						leftValue, err := utils.ToFloat(v.paramsMap[leftVariable.Name])
						if err != nil {
							return err
						}
						if ctx.prefix.GetText() == "++" {
							v.paramsMap[leftVariable.Name] = leftValue + 1
							return leftValue + 1
						} else if ctx.prefix.GetText() == "--" {
							v.paramsMap[leftVariable.Name] = leftValue - 1
							return leftValue - 1
						}
					} else if leftVariable.VariableType == ARRAYORMAP {
						vType := reflect.TypeOf(paramValue)
						if vType.Kind() == reflect.Slice {
							index := -1
							if leftVariable.Index >= 0 {
								index = leftVariable.Index
							} else if leftVariable.IndexnVariableName != "" {
								index = utils.ToIntWithDefaultValue(v.paramsMap[leftVariable.IndexnVariableName], -1)
							}
							if index == -1 {
								return fmt.Errorf("invalid index, variable:%v", leftVariable.String())
							}
							switch sliceValue := paramValue.(type) {
							case []any:
								leftValue, err := utils.ToFloat(sliceValue[index])
								if err != nil {
									return err
								}
								if ctx.prefix.GetText() == "++" {
									sliceValue[index] = leftValue + 1
								} else if ctx.prefix.GetText() == "--" {
									sliceValue[index] = leftValue - 1
								}
								return sliceValue[index]
							case []float64:
								leftValue, err := utils.ToFloat(sliceValue[index])
								if err != nil {
									return err
								}
								if ctx.prefix.GetText() == "++" {
									sliceValue[index] = leftValue + 1
								} else if ctx.prefix.GetText() == "--" {
									sliceValue[index] = leftValue - 1
								}
								return sliceValue[index]
							case []int:
								leftValue, err := utils.ToFloat(sliceValue[index])
								if err != nil {
									return err
								}
								if ctx.prefix.GetText() == "++" {
									sliceValue[index] = utils.ToInt(leftValue + 1)
								} else if ctx.prefix.GetText() == "--" {
									sliceValue[index] = utils.ToInt(leftValue - 1)
								}
								return sliceValue[index]
							default:
								return fmt.Errorf("unspport variable type, variable:%v", sliceValue)
							}
						} else if vType.Kind() == reflect.Map {
							index := ""
							if leftVariable.IndexnName != "" {
								index = leftVariable.IndexnName
							} else if leftVariable.IndexnVariableName != "" {
								index = utils.ToStringWithDefaultValue(v.paramsMap[leftVariable.IndexnVariableName], "")
							}
							if index == "" {
								return fmt.Errorf("invalid index, variable:%v", leftVariable.String())
							}
							if mapValue, ok := paramValue.(map[string]any); ok {
								leftValue, err := utils.ToFloat(mapValue[index])
								if err != nil {
									return err
								}
								if ctx.prefix.GetText() == "++" {
									mapValue[index] = leftValue + 1
								} else if ctx.prefix.GetText() == "--" {
									mapValue[index] = leftValue - 1
								}
								return mapValue[index]
							}
						}
					}
				}

				return fmt.Errorf("prefix op:%s not support variable:%s", ctx.prefix.GetText(), leftVariable.String())
			}
		} else if ctx.postfix != nil {
			switch ctx.postfix.GetText() {
			case "++", "--":
				leftVariable := v.node2Variables[ctx.Expression(0)]
				if leftVariable == nil || leftVariable.Name == "" {
					return fmt.Errorf("left variable not found")
				}
				paramValue, exist := v.paramsMap[leftVariable.Name]
				if !exist {
					return fmt.Errorf("variable %v not found", leftVariable.Name)
				} else {
					if leftVariable.VariableType == UNKNOWN {
						leftValue, err := utils.ToFloat(v.paramsMap[leftVariable.Name])
						if err != nil {
							return err
						}
						if ctx.postfix.GetText() == "++" {
							v.paramsMap[leftVariable.Name] = leftValue + 1
						} else if ctx.postfix.GetText() == "--" {
							v.paramsMap[leftVariable.Name] = leftValue - 1
						}
						return leftValue
					} else if leftVariable.VariableType == ARRAYORMAP {
						vType := reflect.TypeOf(paramValue)
						if vType.Kind() == reflect.Slice {
							index := -1
							if leftVariable.Index >= 0 {
								index = leftVariable.Index
							} else if leftVariable.IndexnVariableName != "" {
								index = utils.ToIntWithDefaultValue(v.paramsMap[leftVariable.IndexnVariableName], -1)
							}
							if index == -1 {
								return fmt.Errorf("invalid index, variable:%v", leftVariable.String())
							}
							switch sliceValue := paramValue.(type) {
							case []any:
								leftValue, err := utils.ToFloat(sliceValue[index])
								if err != nil {
									return err
								}
								if ctx.postfix.GetText() == "++" {
									sliceValue[index] = leftValue + 1
								} else if ctx.postfix.GetText() == "--" {
									sliceValue[index] = leftValue - 1
								}
								return leftValue
							case []float64:
								leftValue, err := utils.ToFloat(sliceValue[index])
								if err != nil {
									return err
								}
								if ctx.postfix.GetText() == "++" {
									sliceValue[index] = leftValue + 1
								} else if ctx.postfix.GetText() == "--" {
									sliceValue[index] = leftValue - 1
								}
								return leftValue
							case []int:
								leftValue, err := utils.ToFloat(sliceValue[index])
								if err != nil {
									return err
								}
								if ctx.postfix.GetText() == "++" {
									sliceValue[index] = utils.ToInt(leftValue + 1)
								} else if ctx.postfix.GetText() == "--" {
									sliceValue[index] = utils.ToInt(leftValue - 1)
								}
								return leftValue
							default:
								return fmt.Errorf("unspport variable type, variable:%v", sliceValue)
							}
						} else if vType.Kind() == reflect.Map {
							index := ""
							if leftVariable.IndexnName != "" {
								index = leftVariable.IndexnName
							} else if leftVariable.IndexnVariableName != "" {
								index = utils.ToStringWithDefaultValue(v.paramsMap[leftVariable.IndexnVariableName], "")
							}
							if index == "" {
								return fmt.Errorf("invalid index, variable:%v", leftVariable.String())
							}
							if mapValue, ok := paramValue.(map[string]any); ok {
								leftValue, err := utils.ToFloat(mapValue[index])
								if err != nil {
									return err
								}
								if ctx.postfix.GetText() == "++" {
									mapValue[index] = leftValue + 1
								} else if ctx.postfix.GetText() == "--" {
									mapValue[index] = leftValue - 1
								}
								return leftValue
							}
						}
					}
				}

				return fmt.Errorf("postfix op:%s not support variable:%s", ctx.prefix.GetText(), leftVariable.String())
			}

		} else if ctx.bop != nil {
			left := v.Visit(ctx.Expression(0))
			if ctx.bop.GetText() == "." {
				// get field value from struct
				if ctx.IDENTIFIER() != nil {
					value := reflect.ValueOf(left)
					if value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
						value = value.Elem()
					}

					// check value is a struct
					if value.Kind() != reflect.Struct {
						return fmt.Errorf("the value:%v is not a struct", value)
					}

					field, ok := value.Type().FieldByName(ctx.IDENTIFIER().GetText())
					if !ok {
						return fmt.Errorf("the struct:%v does not contain a field named '%s'", value, ctx.IDENTIFIER().GetText())
					}

					fieldValue := value.FieldByName(ctx.IDENTIFIER().GetText())
					if !fieldValue.IsValid() || !fieldValue.CanInterface() {
						return fmt.Errorf("the field:%s is not valid or cannot be accessed", ctx.IDENTIFIER().GetText())
					}

					fieldValueInterface := fieldValue.Interface()
					switch field.Type.Kind() {
					case reflect.String:
						return fieldValueInterface.(string)
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint64:
						return float64(utils.ToInt(fieldValueInterface))
					case reflect.Float32:
						return float64(fieldValueInterface.(float32))
					case reflect.Float64:
						return fieldValueInterface.(float64)
					case reflect.Bool:
						return fieldValueInterface.(bool)
					default:
						return fieldValueInterface
					}
				} else if ctx.FunctionCall() != nil {
					value := reflect.ValueOf(left)
					if value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
						value = value.Elem()
					}

					// check value is a struct
					if value.Kind() != reflect.Struct {
						return fmt.Errorf("the value:%v is not a struct", value)
					}

					funcName := ctx.FunctionCall().IDENTIFIER().GetText()
					_, ok := value.Type().FieldByName(funcName)
					if ok {
						// funcName is the filed of the value, its type is a function
						fieldFuncValue := value.FieldByName(funcName)
						if !fieldFuncValue.IsValid() || !fieldFuncValue.CanInterface() {
							return fmt.Errorf("the field:%s is not valid or cannot be accessed", funcName)
						}
						if fieldFuncValue.Type().Kind() != reflect.Func {
							return fmt.Errorf("the field:%s is not func ,but get kind:%v", funcName, fieldFuncValue.Type().Kind())
						}

						var in []reflect.Value
						for i := 0; i < fieldFuncValue.Type().NumIn(); i++ {
							val := v.Visit(ctx.FunctionCall().ExpressionList().Expression(i))
							in = append(in, utils.ReflectValueof(fieldFuncValue.Type().In(i), val))
						}
						out := fieldFuncValue.Call(in)
						errorType := reflect.TypeOf((*error)(nil)).Elem() // 获取 error 接口的类型
						for _, val := range out {
							if val.Kind() == reflect.Interface && val.Type().Implements(errorType) {
								return val.Interface()
							}
							if val.Type().Kind() == reflect.Ptr && val.Elem().Type().Implements(errorType) {
								return val.Interface()
							}
						}

						if len(out) == 0 {
							return nil
						}

						return out[0].Interface()

					} else {
						var (
							m  reflect.Method
							ok bool
						)
						// funcName is the method name of the value
						m, ok = value.Type().MethodByName(funcName)
						if !ok {
							m, ok = reflect.PointerTo(value.Type()).MethodByName(funcName)
							if !ok {
								return fmt.Errorf("value:%v not have the method name:'%s'", left, funcName)
							}
						}
						if !m.Func.IsValid() || !m.Func.CanInterface() {
							return fmt.Errorf("the method:%s is not valid or cannot be accessed", funcName)
						}
						var in []reflect.Value

						if m.Func.Type().In(0).Kind() == reflect.Struct { // method receiver is struct
							in = append(in, value)
						} else if m.Func.Type().In(0).Kind() == reflect.Pointer { // method receiver is pointer
							if value.CanAddr() {
								in = append(in, value.Addr())
							} else {
								pointerValue := reflect.New(value.Type())
								pointerValue.Elem().Set(value)
								in = append(in, pointerValue)
							}
						}
						for i := 1; i < m.Func.Type().NumIn(); i++ {
							val := v.Visit(ctx.FunctionCall().ExpressionList().Expression(i - 1))
							in = append(in, utils.ReflectValueof(m.Func.Type().In(i), val))
						}
						out := m.Func.Call(in)
						errorType := reflect.TypeOf((*error)(nil)).Elem() // 获取 error 接口的类型
						for _, val := range out {
							if val.Kind() == reflect.Interface && val.Type().Implements(errorType) {
								return val.Interface()
							}
							if val.Type().Kind() == reflect.Ptr && val.Elem().Type().Implements(errorType) {
								return val.Interface()
							}
						}

						if len(out) == 0 {
							return nil
						}

						return out[0].Interface()

					}

				}
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
			case int:
				return utils.LE[float64](float64(val), right.(float64))
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
			case int:
				return utils.GE[float64](float64(val), right.(float64))
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
			case int:
				return utils.GT[float64](float64(val), right.(float64))
			default:
				return fmt.Errorf("type:%T not support op:%s, value:%v", val, ctx.bop.GetText(), left)
			}
		case "<":
			switch val := left.(type) {
			case float64:
				return utils.LT[float64](val, right.(float64))
			case string:
				return utils.LT[string](val, right.(string))
			case time.Time:
				return utils.LT[int64](val.UnixMilli(), right.(time.Time).UnixMilli())
			case int:
				return utils.LT[float64](float64(val), right.(float64))
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
			val := reflect.ValueOf(left)
			if val.Kind() == reflect.Slice {
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
					return fmt.Errorf("op:[] not support type:%T, value:%v", left, left)

				}
			} else if val.Kind() == reflect.Map {
				index, ok := right.(string)
				if !ok {
					return fmt.Errorf("right value:%v should be string", right)
				}
				switch m := left.(type) {
				case map[string]string:
					return m[index]
				case map[string]float64:
					return m[index]

				case map[string]any:
					return m[index]
				case map[string]int:
					return float64(m[index])
				case map[string]int64:
					return float64(m[index])
				case map[string]float32:
					return float64(m[index])
				case map[string]bool:
					return m[index]
				default:
					return fmt.Errorf("op:[] not support type:%T, value:%v", left, left)

				}

			} else {
				return fmt.Errorf("op:[] not support type:%T, value:%v", left, left)
			}

		case "=":
			leftVariable := v.node2Variables[ctx.Expression(0)]
			if leftVariable == nil || leftVariable.Name == "" {
				return fmt.Errorf("left variable not found, error:%v, at %d:%d", ctx.Expression(0).GetText(), ctx.Expression(0).GetStart().GetLine(),
					ctx.Expression(0).GetStart().GetColumn())
			}
			if paramValue, exist := v.paramsMap[leftVariable.Name]; !exist {
				v.paramsMap[leftVariable.Name] = right
			} else {
				if leftVariable.VariableType == UNKNOWN {
					v.paramsMap[leftVariable.Name] = right
				} else if leftVariable.VariableType == ARRAYORMAP {
					vType := reflect.TypeOf(paramValue)
					if vType.Kind() == reflect.Slice {
						index := -1
						if leftVariable.Index >= 0 {
							index = leftVariable.Index
						} else if leftVariable.IndexnVariableName != "" {
							index = utils.ToIntWithDefaultValue(v.paramsMap[leftVariable.IndexnVariableName], -1)
						}
						if index == -1 {
							return fmt.Errorf("invalid index, variable:%v", leftVariable.String())
						}
						switch sliceValue := paramValue.(type) {
						case []any:
							sliceValue[index] = right
						case []float64:
							sliceValue[index] = right.(float64)
						case []string:
							sliceValue[index] = right.(string)
						case []bool:
							sliceValue[index] = right.(bool)
						case []int:
							sliceValue[index] = utils.ToInt(right)
						default:
							return fmt.Errorf("unspport variable type, variable:%v", sliceValue)
						}
					} else if vType.Kind() == reflect.Map {
						index := ""
						if leftVariable.IndexnName != "" {
							index = leftVariable.IndexnName
						} else if leftVariable.IndexnVariableName != "" {
							index = utils.ToStringWithDefaultValue(v.paramsMap[leftVariable.IndexnVariableName], "")
						}
						if index == "" {
							return fmt.Errorf("invalid index, variable:%v", leftVariable.String())
						}
						if mapValue, ok := paramValue.(map[string]any); ok {
							mapValue[index] = right
						}
					}
				}
			}
		case "+=", "-=", "*=", "/=":
			leftVariable := v.node2Variables[ctx.Expression(0)]
			if leftVariable == nil || leftVariable.Name == "" {
				return fmt.Errorf("left variable not found")
			}
			paramValue, exist := v.paramsMap[leftVariable.Name]
			if !exist {
				return fmt.Errorf("variable %v not found", leftVariable.Name)
			} else {
				rightValue, err := utils.ToFloat(right)
				if err != nil {
					return err
				}
				if leftVariable.VariableType == UNKNOWN {
					leftValue, err := utils.ToFloat(v.paramsMap[leftVariable.Name])
					if err != nil {
						return err
					}
					switch ctx.bop.GetText() {
					case "+=":
						v.paramsMap[leftVariable.Name] = leftValue + rightValue
					case "-=":
						v.paramsMap[leftVariable.Name] = leftValue - rightValue
					case "*=":
						v.paramsMap[leftVariable.Name] = leftValue * rightValue
					case "/=":
						if rightValue == 0 {
							return fmt.Errorf("divide by zero")
						}
						v.paramsMap[leftVariable.Name] = leftValue / rightValue
					}
				} else if leftVariable.VariableType == ARRAYORMAP {
					vType := reflect.TypeOf(paramValue)
					if vType.Kind() == reflect.Slice {
						index := -1
						if leftVariable.Index >= 0 {
							index = leftVariable.Index
						} else if leftVariable.IndexnVariableName != "" {
							index = utils.ToIntWithDefaultValue(v.paramsMap[leftVariable.IndexnVariableName], -1)
						}
						if index == -1 {
							return fmt.Errorf("invalid index, variable:%v", leftVariable.String())
						}
						switch sliceValue := paramValue.(type) {
						case []any:
							leftValue, err := utils.ToFloat(sliceValue[index])
							if err != nil {
								return err
							}
							switch ctx.bop.GetText() {
							case "+=":
								sliceValue[index] = leftValue + rightValue
							case "-=":
								sliceValue[index] = leftValue - rightValue
							case "*=":
								sliceValue[index] = leftValue * rightValue
							case "/=":
								if rightValue == 0 {
									return fmt.Errorf("divide by zero")
								}
								sliceValue[index] = leftValue / rightValue
							}
						case []float64:
							leftValue, err := utils.ToFloat(sliceValue[index])
							if err != nil {
								return err
							}
							switch ctx.bop.GetText() {
							case "+=":
								sliceValue[index] = leftValue + rightValue
							case "-=":
								sliceValue[index] = leftValue - rightValue
							case "*=":
								sliceValue[index] = leftValue * rightValue
							case "/=":
								if rightValue == 0 {
									return fmt.Errorf("divide by zero")
								}
								sliceValue[index] = leftValue / rightValue
							}
						case []int:
							leftValue, err := utils.ToFloat(sliceValue[index])
							if err != nil {
								return err
							}
							sliceValue[index] = utils.ToInt(leftValue + rightValue)
							switch ctx.bop.GetText() {
							case "+=":
								sliceValue[index] = utils.ToInt(leftValue + rightValue)
							case "-=":
								sliceValue[index] = utils.ToInt(leftValue - rightValue)
							case "*=":
								sliceValue[index] = utils.ToInt(leftValue * rightValue)
							case "/=":
								if rightValue == 0 {
									return fmt.Errorf("divide by zero")
								}
								sliceValue[index] = utils.ToInt(leftValue / rightValue)
							}
						default:
							return fmt.Errorf("unspport variable type, variable:%v", sliceValue)
						}
					} else if vType.Kind() == reflect.Map {
						index := ""
						if leftVariable.IndexnName != "" {
							index = leftVariable.IndexnName
						} else if leftVariable.IndexnVariableName != "" {
							index = utils.ToStringWithDefaultValue(v.paramsMap[leftVariable.IndexnVariableName], "")
						}
						if index == "" {
							return fmt.Errorf("invalid index, variable:%v", leftVariable.String())
						}
						if mapValue, ok := paramValue.(map[string]any); ok {
							leftValue, err := utils.ToFloat(mapValue[index])
							if err != nil {
								return err
							}
							switch ctx.bop.GetText() {
							case "+=":
								mapValue[index] = leftValue + rightValue
							case "-=":
								mapValue[index] = leftValue - rightValue
							case "*=":
								mapValue[index] = leftValue * rightValue
							case "/=":
								if rightValue == 0 {
									return fmt.Errorf("divide by zero")
								}
								mapValue[index] = leftValue / rightValue
							}
						}
					}
				}
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
func (v *StatementASTEvaluator) VisitFloat(ctx *FloatContext) interface{} {

	str := ctx.GetText()
	if f, err := strconv.ParseFloat(str, 64); err == nil {
		return f
	}

	return nil
}

func (v *StatementASTEvaluator) VisitString(ctx *StringContext) interface{} {
	str := ctx.GetText()
	str = strings.Trim(str, "'")
	if t, ok := utils.TryParseTime(str); ok {
		return t
	}
	return str
}

func (v *StatementASTEvaluator) VisitBool(ctx *BoolContext) interface{} {
	str := ctx.GetText()
	return str == "true"
}

func (v *StatementASTEvaluator) VisitArrays(ctx *ArraysContext) interface{} {
	return v.Visit(ctx.Array())
}

func (v *StatementASTEvaluator) VisitArray(ctx *ArrayContext) interface{} {
	arr := make([]any, 0, len(ctx.AllArray_value()))

	for i := 0; i < len(ctx.AllArray_value()); i++ {
		val := v.Visit(ctx.Array_value(i))
		arr = append(arr, val)
	}

	return arr
}

func (v *StatementASTEvaluator) VisitArray_value(ctx *Array_valueContext) interface{} {
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
func (v *StatementASTEvaluator) VisitIdentifier(ctx *IdentifierContext) interface{} {
	str := ctx.GetText()
	str = strings.Trim(str, "[]${}")
	if str != "" {
		if val, ok := v.paramsMap[str]; ok {
			return val
		}
	}
	return fmt.Errorf("param:%s not found", str)
}
func (v *StatementASTEvaluator) VisitFunctionCall(ctx *FunctionCallContext) interface{} {
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

func (v *StatementASTEvaluator) VisitExpressionList(ctx *ExpressionListContext) interface{} {
	var params []any
	for _, exprContext := range ctx.AllExpression() {
		params = append(params, v.Visit(exprContext))
	}

	return params
}
