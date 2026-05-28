package valuate

import (
	"testing"

	govaluate "github.com/Knetic/govaluate"
	"github.com/bruceding/go-antlr-valuate/parser"
)

// Note: this project uses ${name} as the escaped-variable syntax, while govaluate uses [name].
// All comparison benchmarks below keep the semantic shape identical between the two engines
// so allocation and timing can be compared apples-to-apples.

const (
	fullExpr = "2 > 1 &&" +
		"'something' != 'nothing' || " +
		"'2014-01-20' < 'Wed Jul  8 23:07:35 MDT 2015' && " +
		"${escapedVariable-name} <= unescapedvariableName &&" +
		"modifierTest + 1000 / 2 > (80 * 100 % 2)"
	fullExprGovaluate = "2 > 1 &&" +
		"'something' != 'nothing' || " +
		"'2014-01-20' < 'Wed Jul  8 23:07:35 MDT 2015' && " +
		"[escapedVariable-name] <= unescapedvariableName &&" +
		"modifierTest + 1000 / 2 > (80 * 100 % 2)"
)

func BenchmarkSingleParse(bench *testing.B) {

	for i := 0; i < bench.N; i++ {
		NewEvaluableExpression("1")
	}
}

func BenchmarkSimpleParse(bench *testing.B) {

	for i := 0; i < bench.N; i++ {
		NewEvaluableExpression("(requests_made * requests_succeeded / 100) >= 90")
	}
}

/*
Benchmarks all syntax possibilities in one expression.
*/
func BenchmarkFullParse(bench *testing.B) {

	// represents all the major syntax possibilities.
	for i := 0; i < bench.N; i++ {
		NewEvaluableExpression(fullExpr)
	}
}

/*
Benchmarks the bare-minimum evaluation time
*/
func BenchmarkEvaluationSingle(bench *testing.B) {

	expression, _ := NewEvaluableExpression("1")

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(nil)
	}
}

/*
Benchmarks evaluation times of literals (no variables, no modifiers)

Note: govaluate accepts `(2) > (1)` as paren-wrapped numeric literals, but in
this project's grammar `(literal[, ...])` is an array literal. Both sides use
the paren-free form so each engine evaluates the same logical expression.
*/
func BenchmarkEvaluationNumericLiteral(bench *testing.B) {

	expression, _ := NewEvaluableExpression("2 > 1")

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(nil)
	}
}

/*
Benchmarks evaluation times of literals with modifiers
*/
func BenchmarkEvaluationLiteralModifiers(bench *testing.B) {

	expression, _ := NewEvaluableExpression("2 + 2 == 4")

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(nil)
	}
}

func BenchmarkEvaluationParameter(bench *testing.B) {

	expression, _ := NewEvaluableExpression("requests_made")
	parameters := map[string]interface{}{
		"requests_made": 99.0,
	}

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

/*
Benchmarks evaluation times of parameters
*/
func BenchmarkEvaluationParameters(bench *testing.B) {

	expression, _ := NewEvaluableExpression("requests_made > requests_succeeded")
	parameters := map[string]interface{}{
		"requests_made":      99.0,
		"requests_succeeded": 90.0,
	}

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

/*
Benchmarks evaluation times of parameters + literals with modifiers
*/
func BenchmarkEvaluationParametersModifiers(bench *testing.B) {

	expression, _ := NewEvaluableExpression("(requests_made * requests_succeeded / 100) >= 90")
	parameters := map[string]interface{}{
		"requests_made":      99.0,
		"requests_succeeded": 90.0,
	}

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

// =============================================================================
// govaluate (github.com/Knetic/govaluate) comparison benchmarks
// =============================================================================

func BenchmarkGovaluate_SingleParse(bench *testing.B) {

	for i := 0; i < bench.N; i++ {
		govaluate.NewEvaluableExpression("1")
	}
}

func BenchmarkGovaluate_SimpleParse(bench *testing.B) {

	for i := 0; i < bench.N; i++ {
		govaluate.NewEvaluableExpression("(requests_made * requests_succeeded / 100) >= 90")
	}
}

func BenchmarkGovaluate_FullParse(bench *testing.B) {

	for i := 0; i < bench.N; i++ {
		govaluate.NewEvaluableExpression(fullExprGovaluate)
	}
}

func BenchmarkGovaluate_EvaluationSingle(bench *testing.B) {

	expression, _ := govaluate.NewEvaluableExpression("1")

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(nil)
	}
}

func BenchmarkGovaluate_EvaluationNumericLiteral(bench *testing.B) {

	expression, _ := govaluate.NewEvaluableExpression("2 > 1")

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(nil)
	}
}

func BenchmarkGovaluate_EvaluationLiteralModifiers(bench *testing.B) {

	expression, _ := govaluate.NewEvaluableExpression("2 + 2 == 4")

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(nil)
	}
}

func BenchmarkGovaluate_EvaluationParameter(bench *testing.B) {

	expression, _ := govaluate.NewEvaluableExpression("requests_made")
	parameters := map[string]interface{}{
		"requests_made": 99.0,
	}

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

func BenchmarkGovaluate_EvaluationParameters(bench *testing.B) {

	expression, _ := govaluate.NewEvaluableExpression("requests_made > requests_succeeded")
	parameters := map[string]interface{}{
		"requests_made":      99.0,
		"requests_succeeded": 90.0,
	}

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

func BenchmarkGovaluate_EvaluationParametersModifiers(bench *testing.B) {

	expression, _ := govaluate.NewEvaluableExpression("(requests_made * requests_succeeded / 100) >= 90")
	parameters := map[string]interface{}{
		"requests_made":      99.0,
		"requests_succeeded": 90.0,
	}

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

// =============================================================================
// Additional comparison scenarios: logical, ternary, function call, nested
// arithmetic. Each scenario covers both Parse and Evaluate.
// =============================================================================

const (
	logicalExpr  = "requests_made > 90 && requests_succeeded > 80"
	ternaryExpr  = "requests_made > 90 ? 1.0 : 0.0"
	funcCallExpr = "strlen('someReallyLongInputString') <= 16"
	nestedExpr   = "((a + b) * 2) / (c + 1) > 0.5"
)

// govaluateStrlen mirrors the project's built-in `len` for strings, registered
// for govaluate so the function-call benchmark is comparable.
var govaluateStrlen = map[string]govaluate.ExpressionFunction{
	"strlen": func(args ...interface{}) (interface{}, error) {
		return float64(len(args[0].(string))), nil
	},
}

// ----- Logical (&&) -----

func BenchmarkParse_Logical(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		NewEvaluableExpression(logicalExpr)
	}
}

func BenchmarkGovaluate_Parse_Logical(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		govaluate.NewEvaluableExpression(logicalExpr)
	}
}

func BenchmarkEvaluation_Logical(bench *testing.B) {
	expression, _ := NewEvaluableExpression(logicalExpr)
	parameters := map[string]interface{}{
		"requests_made":      99.0,
		"requests_succeeded": 90.0,
	}
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

func BenchmarkGovaluate_Evaluation_Logical(bench *testing.B) {
	expression, _ := govaluate.NewEvaluableExpression(logicalExpr)
	parameters := map[string]interface{}{
		"requests_made":      99.0,
		"requests_succeeded": 90.0,
	}
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

// ----- Ternary (? :) -----

func BenchmarkParse_Ternary(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		NewEvaluableExpression(ternaryExpr)
	}
}

func BenchmarkGovaluate_Parse_Ternary(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		govaluate.NewEvaluableExpression(ternaryExpr)
	}
}

func BenchmarkEvaluation_Ternary(bench *testing.B) {
	expression, _ := NewEvaluableExpression(ternaryExpr)
	parameters := map[string]interface{}{
		"requests_made": 99.0,
	}
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

func BenchmarkGovaluate_Evaluation_Ternary(bench *testing.B) {
	expression, _ := govaluate.NewEvaluableExpression(ternaryExpr)
	parameters := map[string]interface{}{
		"requests_made": 99.0,
	}
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

// ----- Function call (custom strlen) -----

func BenchmarkParse_FunctionCall(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		NewEvaluableExpression(funcCallExpr)
	}
}

func BenchmarkGovaluate_Parse_FunctionCall(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		govaluate.NewEvaluableExpressionWithFunctions(funcCallExpr, govaluateStrlen)
	}
}

func BenchmarkEvaluation_FunctionCall(bench *testing.B) {
	expression, _ := NewEvaluableExpression(funcCallExpr)
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(nil)
	}
}

func BenchmarkGovaluate_Evaluation_FunctionCall(bench *testing.B) {
	expression, _ := govaluate.NewEvaluableExpressionWithFunctions(funcCallExpr, govaluateStrlen)
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(nil)
	}
}

// ----- Nested arithmetic -----

func BenchmarkParse_NestedArith(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		NewEvaluableExpression(nestedExpr)
	}
}

func BenchmarkGovaluate_Parse_NestedArith(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		govaluate.NewEvaluableExpression(nestedExpr)
	}
}

func BenchmarkEvaluation_NestedArith(bench *testing.B) {
	expression, _ := NewEvaluableExpression(nestedExpr)
	parameters := map[string]interface{}{
		"a": 10.0,
		"b": 20.0,
		"c": 5.0,
	}
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

func BenchmarkGovaluate_Evaluation_NestedArith(bench *testing.B) {
	expression, _ := govaluate.NewEvaluableExpression(nestedExpr)
	parameters := map[string]interface{}{
		"a": 10.0,
		"b": 20.0,
		"c": 5.0,
	}
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

// ----- Struct field access benchmarks removed: govaluate v3 does not support
//       `obj.Field` accessor without a custom Selector / ParameterizedFunctions,
//       so a meaningful 1:1 comparison is not possible here.

// =============================================================================
// Real-world business scenario: PAI-Rec ScoreRewrite/RankScore.
// Mirrors a production ASTType=antlr config snippet:
//
//	max_index = maxIndex(${custom_ple_v2_simple_probs_mmd_play_duration})
//	max_value = maxValue(${custom_ple_v2_simple_probs_mmd_play_duration})
//	RankScore = ${is_valid} * 1.0 + ${is_full} * 0.0 + ${max_index} / 35 * 0.15
//
// This combines: function call on array param, escaped ${var} identifiers, and
// multi-variable arithmetic. It is the primary scenario this optimization
// pass targets.
// =============================================================================

const (
	rankScoreExpr = "${custom_ple_v2_simple_probs_mmd_is_valid} * 1.0 + " +
		"${custom_ple_v2_simple_probs_mmd_is_full} * 0.0 + " +
		"${max_index} / 35 * 0.15"
	maxIndexExpr = "maxIndex(${custom_ple_v2_simple_probs_mmd_play_duration})"
	maxValueExpr = "maxValue(${custom_ple_v2_simple_probs_mmd_play_duration})"
)

// scoreFunctions are the user-defined functions used by ScoreRewrite.
// Both engines receive equivalent implementations so the comparison is fair.
var (
	scoreFunctions = map[string]parser.ExpressionFunction{
		"maxIndex": func(args ...interface{}) (interface{}, error) {
			arr := args[0].([]any)
			var (
				idx int
				max float64
			)
			for i, v := range arr {
				f := v.(float64)
				if i == 0 || f > max {
					max = f
					idx = i
				}
			}
			return float64(idx), nil
		},
		"maxValue": func(args ...interface{}) (interface{}, error) {
			arr := args[0].([]any)
			var max float64
			for i, v := range arr {
				f := v.(float64)
				if i == 0 || f > max {
					max = f
				}
			}
			return max, nil
		},
	}
	govaluateScoreFunctions = map[string]govaluate.ExpressionFunction{
		"maxIndex": func(args ...interface{}) (interface{}, error) {
			arr := args[0].([]float64)
			var (
				idx int
				max float64
			)
			for i, f := range arr {
				if i == 0 || f > max {
					max = f
					idx = i
				}
			}
			return float64(idx), nil
		},
		"maxValue": func(args ...interface{}) (interface{}, error) {
			arr := args[0].([]float64)
			var max float64
			for i, f := range arr {
				if i == 0 || f > max {
					max = f
				}
			}
			return max, nil
		},
	}
)

func rankScoreParams() map[string]interface{} {
	return map[string]interface{}{
		"custom_ple_v2_simple_probs_mmd_is_valid": 0.8,
		"custom_ple_v2_simple_probs_mmd_is_full":  0.3,
		"max_index":                               12.0,
	}
}

func maxIndexParamsAny() map[string]interface{} {
	arr := make([]any, 35)
	for i := range arr {
		arr[i] = float64(i) * 0.1
	}
	return map[string]interface{}{
		"custom_ple_v2_simple_probs_mmd_play_duration": arr,
	}
}

func maxIndexParamsFloat64() map[string]interface{} {
	arr := make([]float64, 35)
	for i := range arr {
		arr[i] = float64(i) * 0.1
	}
	return map[string]interface{}{
		"custom_ple_v2_simple_probs_mmd_play_duration": arr,
	}
}

// ----- RankScore (multi-variable arithmetic with escaped ${...} variables) -----

func BenchmarkParse_RankScore(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		NewEvaluableExpression(rankScoreExpr)
	}
}

func BenchmarkEvaluation_RankScore(bench *testing.B) {
	expression, _ := NewEvaluableExpression(rankScoreExpr)
	parameters := rankScoreParams()
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

func BenchmarkGovaluate_Evaluation_RankScore(bench *testing.B) {
	// govaluate uses [name] for escaped variables instead of ${name}.
	gexpr := "[custom_ple_v2_simple_probs_mmd_is_valid] * 1.0 + " +
		"[custom_ple_v2_simple_probs_mmd_is_full] * 0.0 + " +
		"[max_index] / 35 * 0.15"
	expression, _ := govaluate.NewEvaluableExpression(gexpr)
	parameters := rankScoreParams()
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

// ----- maxIndex(array) (function call on escaped ${...} array variable) -----

func BenchmarkParse_MaxIndex(bench *testing.B) {
	for i := 0; i < bench.N; i++ {
		NewEvaluableExpressionWithFunctions(maxIndexExpr, scoreFunctions)
	}
}

func BenchmarkEvaluation_MaxIndex(bench *testing.B) {
	expression, _ := NewEvaluableExpressionWithFunctions(maxIndexExpr, scoreFunctions)
	parameters := maxIndexParamsAny()
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

func BenchmarkGovaluate_Evaluation_MaxIndex(bench *testing.B) {
	gexpr := "maxIndex([custom_ple_v2_simple_probs_mmd_play_duration])"
	expression, _ := govaluate.NewEvaluableExpressionWithFunctions(gexpr, govaluateScoreFunctions)
	parameters := maxIndexParamsFloat64()
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

// ----- maxValue(array) -----

func BenchmarkEvaluation_MaxValue(bench *testing.B) {
	expression, _ := NewEvaluableExpressionWithFunctions(maxValueExpr, scoreFunctions)
	parameters := maxIndexParamsAny()
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}

func BenchmarkGovaluate_Evaluation_MaxValue(bench *testing.B) {
	gexpr := "maxValue([custom_ple_v2_simple_probs_mmd_play_duration])"
	expression, _ := govaluate.NewEvaluableExpressionWithFunctions(gexpr, govaluateScoreFunctions)
	parameters := maxIndexParamsFloat64()
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(parameters)
	}
}
