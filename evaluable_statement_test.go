package valuate

import (
	"reflect"
	"testing"
)

func TestEvaluableStatement(t *testing.T) {

	testCases := []struct {
		input       string
		expectValue any
		name        string
	}{
		{
			input: `
			sum = 0;
			foreach (values as k => v) {
				sum += v; 
			}
			`,
			expectValue: float64(6),
			name:        "sum",
		},
	}
	for _, tcase := range testCases {
		statement, err := NewEvaluableStatement(tcase.input)
		if err != nil {
			t.Fatal(err)
		}

		paramMap := map[string]any{
			"b":                4,
			"values":           []int{1, 2, 3},
			"input_values_map": map[string]any{"1": 1, "2": 2, "3": 3},
			"out_values_map":   map[string]any{},
		}

		resultMap, err := statement.Evaluate(paramMap)
		if err != nil {
			t.Fatal(err)
		}

		switch expectVal := tcase.expectValue.(type) {
		case float64:
			if expectVal != resultMap[tcase.name] {
				t.Fatalf("expect value is not equal, expected %v, got %v", expectVal, resultMap[tcase.name])
			}
		case []int:
			if !reflect.DeepEqual(expectVal, resultMap[tcase.name]) {
				t.Fatalf("expect value is not equal, expected %v, %T, got %v, %T", expectVal, expectVal, resultMap[tcase.name], resultMap[tcase.name])
			}
		case []any:
			if !reflect.DeepEqual(expectVal, resultMap[tcase.name]) {
				t.Fatalf("expect value is not equal, expected %v, %T, got %v, %T", expectVal, expectVal, resultMap[tcase.name], resultMap[tcase.name])
			}
		case map[string]any:
			if !reflect.DeepEqual(expectVal, resultMap[tcase.name]) {
				t.Fatalf("expect value is not equal, expected %v, %T, got %v, %T", expectVal, expectVal, resultMap[tcase.name], resultMap[tcase.name])
			}

		default:
			t.Fatalf("expect value is not equal, expected %v, %T, got %v, %T", expectVal, expectVal, resultMap[tcase.name], resultMap[tcase.name])
		}
	}
}
