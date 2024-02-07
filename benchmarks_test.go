package valuate

import "testing"

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
	expression := "2 > 1 &&" +
		"'something' != 'nothing' || " +
		"'2014-01-20' < 'Wed Jul  8 23:07:35 MDT 2015' && " +
		"[escapedVariable-name] <= unescapedvariableName &&" +
		"modifierTest + 1000 / 2 > (80 * 100 % 2)"

	for i := 0; i < bench.N; i++ {
		NewEvaluableExpression(expression)
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
*/
func BenchmarkEvaluationNumericLiteral(bench *testing.B) {

	expression, _ := NewEvaluableExpression("(2) > (1)")

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		expression.Evaluate(nil)
	}
}

/*
Benchmarks evaluation times of literals with modifiers
*/
func BenchmarkEvaluationLiteralModifiers(bench *testing.B) {

	expression, _ := NewEvaluableExpression("(2) + (2) == (4)")

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
