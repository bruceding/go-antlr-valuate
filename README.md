# go-antlr-valuate
English | [简体中文](./README_zh-CN.md)

Inspired by [govaluate](https://github.com/Knetic/govaluate), this implementation utilizes antlr4 to provide a faster development experience and enhanced functionality.


## How to Install

```
go get -u github.com/bruceding/go-antlr-valuate
```

## How to Use 

Execute an expression directly:

```go
expr, err := valuate.NewEvaluableExpression("10 > 0");
result, err := expr.Evaluate(nil);
// result is now set to "true", the bool value.
```

For expressions containing parameters, execute as follows:

```go
expr, err := valuate.NewEvaluableExpression("foo > 0");

parameters := make(map[string]interface{}, 8)
parameters["foo"] = -1;

result, err := expr.Evaluate(parameters);
// result is now set to "false", the bool value.

```

How to perform string type comparisons?


```go
expr, err := valuate.NewEvaluableExpression("http_response_body == 'service is ok'");

parameters := make(map[string]interface{}, 8)
parameters["http_response_body"] = "service is ok";

result, err := expr.Evaluate(parameters);
// result is now set to "true", the bool value.

```

How to work with date types in comparisons?


```go
expr, err := valuate.NewEvaluableExpression("'2014-01-02' > '2014-01-01 23:59:59'");
result, err := expr.Evaluate(nil);

// result is now set to true
```
The examples above return boolean types, but you can also calculate and obtain a result directly:


```go
expr, err := valuate.NewEvaluableExpression("(mem_used / total_mem) * 100");

parameters := make(map[string]interface{}, 8)
parameters["total_mem"] = 1024;
parameters["mem_used"] = 512;

result, err := expr.Evaluate(parameters);
// result is now set to "50.0", the float64 value.

```

## Special Characters in Variables

For instance, the expression below is parsed into two parameters, response and time, then subtracts time from response, and finally compares it to 100.

```
"response-time < 100"
```

If you want to treat response-time as a single parameter, you can write it like this:

```
${response-time} < 100
```

## Function Support

Unlike govaluate, we have built-in implementation for some functions; you can refer to [expression_functions.go](parser/expression_functions.go).
To implement a custom function, you can define it as follows:

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
We have a built-in len function to return the length of a string or array. The above code can be written as:

```go
expString := "len('someReallyLongInputString') <= 16"
expr, _ := valuate.NewEvaluableExpression(expString)
result, _ := expr.Evaluate(nil)
// result is now "false", the boolean value
```

## Accessors 
If there are array type parameters in the variable, you can access array elements by index, e.g., to access the first element of arr:

```
arr[0]
```

The system supports array definitions, and indexes can also be passed directly as parameters, e.g., a is a parameter:

```
(1,2,3,4)[a]
```

If there are struct type parameters in the variable, you can access fields or methods in the usual way. If foo is a struct with field Bar, you can access it like so:

```
foo.Bar > 2
```

Nested access is also supported, for example if Bar is also a struct with field F:

```
foo.Bar.F > 2
```
Similarly, accessing methods is done in a similar manner. If foo is a struct with a method named Function, it can be accessed like this:

```
foo.Function() > 2
```

If Function can take arguments, it can be accessed like so:

```
foo.Function(1, 2) > 2
```
Therefore, if foo has a method, whether defined through a struct or pointer, it can be accessed using the form above. Specific tests can be found in TestOperatorParseWithFunction, located [here](parser/ast_evaluator_test.go).

If foo is a map with index bar, it can be accessed like this:

```
foo['bar'] > 2
```

## Supported Operators and Data Types

* Numeric types, all parsed as float64 (e.g., 123.45)
* String types, using single quotes (e.g., 'foo')
* Date types, defined like strings but converted to time.Time if they comply with date format specifications
* Boolean types, represented by `true` or `false`
* Array types, elements separated by commas and defined using parentheses (e.g., (1,2,3,'foo'))
* Parentheses `()` to alter order of operations
* Prefix modifiers: ++ -- - + (unary plus and minus), ! (logical not), ~ (bitwise not)
* Arithmetic operators: + - * / % (mod), ** or ^ (power)
* Comparison operators: == != < <= > >=
* Logical operators: && ||
* Bitwise operators: & |
* Ternary operator: ? :
* Access array indices: []
* Access fields or methods:  `.`

## Usage of Syntax

Refer to the antlr4 grammar file [Govaluate.g4](parser/Govaluate.g4)
If the grammar is adjusted, generate the code by running:
 
```
cd parser 
antlr4 -Dlanguage=Go  -visitor Govaluate.g4
```

