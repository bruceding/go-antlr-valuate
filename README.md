# go-antlr-valuate
参考 [govaluate](https://github.com/Knetic/govaluate) 实现， 使用 [antlr4](https://www.antlr.org/) 重新实现, 可以更快速的开发及增加功能。

## 如何安装

```
go get -u github.com/bruceding/go-antlr-valuate
```

## 如何使用

直接执行表达式

```go
expr, err := expression.NewEvaluableExpression("10 > 0");
result, err := expr.Evaluate(nil);
// result is now set to "true", the bool value.
```
表达式里带有参数，如何执行？

```go
expr, err := expression.NewEvaluableExpression("foo > 0");

parameters := make(map[string]interface{}, 8)
parameters["foo"] = -1;

result, err := expr.Evaluate(parameters);
// result is now set to "false", the bool value.

```
如何使用 string 类型进行判断？

```go
expr, err := expression.NewEvaluableExpression("http_response_body == 'service is ok'");

parameters := make(map[string]interface{}, 8)
parameters["http_response_body"] = "service is ok";

result, err := expr.Evaluate(parameters);
// result is now set to "true", the bool value.

```
如何使用时间类型进行判断？

```go
expr, err := expression.NewEvaluableExpression("'2014-01-02' > '2014-01-01 23:59:59'");
result, err := expr.Evaluate(nil);

// result is now set to true
```
上面的例子都是 bool 类型的返回， 也可以直接计算得出结果

```go
expr, err := expression.NewEvaluableExpression("(mem_used / total_mem) * 100");

parameters := make(map[string]interface{}, 8)
parameters["total_mem"] = 1024;
parameters["mem_used"] = 512;

result, err := expre.Evaluate(parameters);
// result is now set to "50.0", the float64 value.

```

## 变量中的特殊字符

比如下面的表达式， 目前是解析成两个参数， response 和 time, response 减去 time 之后， 再和 100 进行比较。
```
"response-time < 100"
```
如果想把 response-time 当成一个参数怎么处理？ 可以写成

```
[response-time] < 100
```
或者 
```
${response-time} < 100
```
## 函数的支持

与 govaluate  不同的是，我们内置了部分的 function 的实现， 具体可以参考 [expression_functions.go](parser/expression_functions.go)

如果实现自定义的 function ， 可以写成
```go
functions := map[string]govaluate.ExpressionFunction {
		"strlen": func(args ...interface{}) (interface{}, error) {
			length := len(args[0].(string))
			return (float64)(length), nil
		},
	}

expString := "strlen('someReallyLongInputString') <= 16"
expr, _ := expression.NewEvaluableExpressionWithFunctions(expString, functions)

result, _ := expr.Evaluate(nil)
// result is now "false", the boolean value
```

## 支持哪些运算符和数据类型
* 数值类型, 全部解析成 float64(123.45)
* string 类型， 使用 `'` ('foo')
* 时间类型，和 string 类型一样的定义，如果符合时间的格式规范，会转成 time.Time
* bool 类型， true 或者 false 表示
* 数组类型，使用 `,` 分隔元素， 使用 `()` 定义。 例如 (1,2,3, 'foo')
* 使用 `()` 改变运算顺序
* 三元运算符 ? :

