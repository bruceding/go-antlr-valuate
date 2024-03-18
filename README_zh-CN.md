# go-antlr-valuate
参考 [govaluate](https://github.com/Knetic/govaluate) 实现， 使用 [antlr4](https://www.antlr.org/) 重新实现, 可以更快速的开发及增加功能。

## 如何安装

```
go get -u github.com/bruceding/go-antlr-valuate
```

## 如何使用

直接执行表达式

```go
expr, err := valuate.NewEvaluableExpression("10 > 0");
result, err := expr.Evaluate(nil);
// result is now set to "true", the bool value.
```
表达式里带有参数，如何执行？

```go
expr, err := valuate.NewEvaluableExpression("foo > 0");

parameters := make(map[string]interface{}, 8)
parameters["foo"] = -1;

result, err := expr.Evaluate(parameters);
// result is now set to "false", the bool value.

```
如何使用 string 类型进行判断？

```go
expr, err := valuate.NewEvaluableExpression("http_response_body == 'service is ok'");

parameters := make(map[string]interface{}, 8)
parameters["http_response_body"] = "service is ok";

result, err := expr.Evaluate(parameters);
// result is now set to "true", the bool value.

```
如何使用时间类型进行判断？

```go
expr, err := valuate.NewEvaluableExpression("'2014-01-02' > '2014-01-01 23:59:59'");
result, err := expr.Evaluate(nil);

// result is now set to true
```
上面的例子都是 bool 类型的返回， 也可以直接计算得出结果

```go
expr, err := valuate.NewEvaluableExpression("(mem_used / total_mem) * 100");

parameters := make(map[string]interface{}, 8)
parameters["total_mem"] = 1024;
parameters["mem_used"] = 512;

result, err := expr.Evaluate(parameters);
// result is now set to "50.0", the float64 value.

```

## 变量约束
* 第一个字符必须是字母
* 后面的字符可以是字母、数字，下划线(_)、 连字符(-)
* 可以用 `${}` 包裹变量名称
比如下面的表达式， 目前是解析成两个参数， response 和 time, response 减去 time 之后， 再和 100 进行比较。
```
"response-time < 100"
```
如果想把 response-time 当成一个参数怎么处理？ 可以写成

```
${response-time} < 100
```
## 函数的支持

与 govaluate  不同的是，我们内置了部分的 function 的实现， 具体可以参考 [expression_functions.go](parser/expression_functions.go)

如果实现自定义的 function ， 可以写成
```go
functions := map[string]parser.ExpressionFunction {
		"strlen": func(args ...interface{}) (interface{}, error) {
			length := len(args[0].(string))
			return (float64)(length), nil
		},
	}

expString := "strlen('someReallyLongInputString') <= 16"
expr, _ := valuate.NewEvaluableExpressionWithFunctions(expString, functions)

result, _ := expr.Evaluate(nil)
// result is now "false", the boolean value
```
我们内置了 len 函数来返回 string 或者 array 的长度， 上面的代码可以写成
```go
expString := "len('someReallyLongInputString') <= 16"
expr, _ := valuate.NewEvaluableExpression(expString)
result, _ := expr.Evaluate(nil)
// result is now "false", the boolean value
```

## 访问器
如果变量中有数组的类型参数， 可以通过下标来访问数组元素，arr 是一个数组，访问第一个元素
```
arr[1]
```
系统支持数组的定义，下标也可以通过参数直接传递, a 是一个参数
```
(1,2,3,4)[a]
```

如果变量中有 struct 的类型参数，可以按照通常的方式访问字段或者方法。如果 foo 是个 struct, 有变量 Bar, 可以这样访问
```
foo.Bar > 2
```
也可以嵌套访问， 比如 Bar 也是一个 struct, 拥有变量 F, 

```
foo.Bar.F > 2
```
同样的，访问方法也是类似的。 如果 foo 是一个 struct, 拥有字段名称 Function, 但是 Function 是一个方法， 可以这样访问
```
foo.Function() > 2
```
如果 Function 可以有参数，可以这样访问
```
foo.Function(1, 2) > 2
```
那么，如果 foo 拥有一个方法，不管是通过 struct 还是指针定义的，都可以使用上面的形式进行访问。具体的测试可以参考 TestOperatorParseWithFunction， 在[这里](parser/ast_evaluator_test.go) 能找到。

如果 foo 是个 map 数组, 有索引 bar, 可以这样访问
```
foo['bar'] > 2
```

## 支持哪些运算符和数据类型
* 数值类型, 全部解析成 float64(123.45)
* string 类型， 使用 `'` ('foo')
* 时间类型，和 string 类型一样的定义，如果符合时间的格式规范，会转成 time.Time
* bool 类型， true 或者 false 表示
* 数组类型，使用 `,` 分隔元素， 使用 `()` 定义。 例如 (1,2,3, 'foo')
* 使用 `()` 改变运算顺序
* 前缀修饰，支持 ++ -- - +（一元加与减） !(逻辑非) ~(逐位非)
* 算术运算符，支持 + - * / %(取模) **(次方，同 ^) 
* 比较运算符，支持 == != < <= > >=
* 逻辑运算符，支持 && ||
* 位运算符 & | 
* 三元运算符 ? :
* 访问数组坐标 []
* 访问字段或者方法 `.`

## 语法的使用
antlr4 的语法文件参考 [Govaluate.g4](parser/Govaluate.g4)

如果调整了语法，可以通过 
```
cd parser 
antlr4 -Dlanguage=Go  -visitor Govaluate.g4
```
生成代码

