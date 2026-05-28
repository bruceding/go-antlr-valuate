//go:build go1.25

package valuate

import (
	"math"
	"sync"
	"testing"
	"testing/synctest"

	"github.com/bruceding/go-antlr-valuate/parser"
)

// TestConcurrentEvaluate_Synctest verifies that a single EvaluableExpression
// instance can be shared safely across many goroutines. It exercises:
//   - the sync.Pool reuse path of ASTEvaluator (Get / Reset / Put);
//   - the read-only PreCompiled caches (Literals / VarNames / ExprChildren /
//     ResolvedFuncs) populated at parse time;
//   - the constant-fold fast path in VisitExpression;
//   - user-defined functions invoked concurrently with shared parameter maps.
//
// Each sub-case runs inside a synctest bubble so the goroutine schedule is
// deterministic; combined with `go test -race` this surfaces any data race
// or stale-state bug in the shared caches.
func TestConcurrentEvaluate_Synctest(t *testing.T) {
	const (
		goroutines = 64
		iterations = 200
	)

	// Pre-compute the expected RankScore so we can verify result correctness
	// without relying on map iteration order:
	//   0.8 * 1.0 + 0.3 * 0.0 + 12.0 / 35.0 * 0.15
	rankScoreWant := 0.8*1.0 + 0.3*0.0 + 12.0/35.0*0.15

	cases := []struct {
		name      string
		expr      string
		funcs     map[string]parser.ExpressionFunction
		params    map[string]interface{}
		want      interface{}
		wantFloat *float64 // when non-nil, compare with epsilon
	}{
		{
			name: "constant_folded_bool",
			expr: "2 + 2 == 4",
			want: true,
		},
		{
			name: "constant_folded_arith",
			expr: "1 + 2 * 3 - 4",
			want: float64(3),
		},
		{
			name:   "single_param",
			expr:   "requests_made",
			params: map[string]interface{}{"requests_made": 99.0},
			want:   float64(99),
		},
		{
			name:   "ternary_short_circuit",
			expr:   "requests_made > 90 ? 1.0 : 0.0",
			params: map[string]interface{}{"requests_made": 99.0},
			want:   float64(1),
		},
		{
			name:      "rank_score_business",
			expr:      rankScoreExpr,
			params:    rankScoreParams(),
			wantFloat: &rankScoreWant,
		},
		{
			name:   "max_index_business",
			expr:   maxIndexExpr,
			funcs:  scoreFunctions,
			params: maxIndexParamsAny(),
			want:   float64(34),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				expr, err := NewEvaluableExpressionWithFunctions(tc.expr, tc.funcs)
				if err != nil {
					t.Fatalf("parse %q: %v", tc.expr, err)
				}

				var wg sync.WaitGroup
				wg.Add(goroutines)
				for g := 0; g < goroutines; g++ {
					go func() {
						defer wg.Done()
						for i := 0; i < iterations; i++ {
							got, evalErr := expr.Evaluate(tc.params)
							if evalErr != nil {
								t.Errorf("evaluate failed: %v", evalErr)
								return
							}
							if tc.wantFloat != nil {
								gotF, ok := got.(float64)
								if !ok {
									t.Errorf("got %T(%v), want float64", got, got)
									return
								}
								if math.Abs(gotF-*tc.wantFloat) > 1e-9 {
									t.Errorf("got %v, want %v", gotF, *tc.wantFloat)
									return
								}
								continue
							}
							if got != tc.want {
								t.Errorf("got %v(%T), want %v(%T)", got, got, tc.want, tc.want)
								return
							}
						}
					}()
				}
				wg.Wait()
			})
		})
	}
}

// TestConcurrentParseAndEvaluate_Synctest verifies that parsing and evaluating
// independent EvaluableExpression instances concurrently does not corrupt the
// shared built-in function map (BuildInFunctionsMap) or any package-level
// state.
func TestConcurrentParseAndEvaluate_Synctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const goroutines = 32
		const iterations = 50

		var wg sync.WaitGroup
		wg.Add(goroutines)
		for g := 0; g < goroutines; g++ {
			go func() {
				defer wg.Done()
				for i := 0; i < iterations; i++ {
					expr, err := NewEvaluableExpressionWithFunctions(maxIndexExpr, scoreFunctions)
					if err != nil {
						t.Errorf("parse failed: %v", err)
						return
					}
					got, err := expr.Evaluate(maxIndexParamsAny())
					if err != nil {
						t.Errorf("evaluate failed: %v", err)
						return
					}
					if got != float64(34) {
						t.Errorf("got %v, want 34", got)
						return
					}
				}
			}()
		}
		wg.Wait()
	})
}
