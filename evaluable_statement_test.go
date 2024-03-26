package valuate

import (
	"reflect"
	"strings"
	"testing"
)

func TestEvaluableStatement(t *testing.T) {

	testCases := []struct {
		input       string
		expectValue any
		name        string
	}{
		{
			input:       " values[0] = b;",
			expectValue: []int{4, 2, 3},
			name:        "values",
		},
		{
			input:       "output_values = (1,2,3);\n values[0] = b;\n output_values[0] = values[0];",
			expectValue: []any{float64(4), float64(2), float64(3)},
			name:        "output_values",
		},
		{
			input:       "sum = 0;\n sum += b; sum +=b;",
			expectValue: float64(8),
			name:        "sum",
		},
		{
			input:       "values[0] +=b;",
			expectValue: []int{5, 2, 3},
			name:        "values",
		},
		{
			input:       "c = values[2] - values[0];",
			expectValue: float64(2),
			name:        "c",
		},
		{
			input:       "c = b;\n c -= values[1];",
			expectValue: float64(2),
			name:        "c",
		},
		{
			input:       "a = 8;\n  a /= b;",
			expectValue: float64(2),
			name:        "a",
		},
		{
			input:       "sum = 0;\n  a = ++sum;",
			expectValue: float64(1),
			name:        "a",
		},
		{
			input:       "sum = 0;\n  a = ++sum;",
			expectValue: float64(1),
			name:        "sum",
		},
		{
			input:       "sum = 4;\n  a = --sum;",
			expectValue: float64(3),
			name:        "sum",
		},
		{
			input:       "sum = 4;\n  a = sum--;",
			expectValue: float64(4),
			name:        "a",
		},
		{
			input:       "sum = 4;\n  a = sum--;",
			expectValue: float64(3),
			name:        "sum",
		},
		{
			input: `sum = 0;
			for (i = 0;  i < 3; i++) {
				sum += values[i];
			}`,
			expectValue: float64(6),
			name:        "sum",
		},
		{
			input: `a = 0 ;
			if (b > 2) {
				a = 1;
			} else {
				a = 2;
			}`,
			expectValue: float64(1),
			name:        "a",
		},
		{
			input: `a = 0 ;
			if (b > 8) {
				a = 1;
			} else {
				a = 2;
			}`,
			expectValue: float64(2),
			name:        "a",
		},
		{
			input: `a = 0 ;
			if (b > 8) {
				a = 1;
			}`,
			expectValue: float64(0),
			name:        "a",
		},
		{
			input: `a = 0 ;
			if (b > 8) {
				a = 1;
			} else if (b == 4){
					a = 3;
			} else {
				a = 2;
			}`,
			expectValue: float64(3),
			name:        "a",
		},
		{
			input: `out_values = (0,0,0);
			for (i = 0;  i < 3; i++) {
				out_values[i] = values[i];
			}`,
			expectValue: []any{float64(1), float64(2), float64(3)},
			name:        "out_values",
		},
		{
			input: `out_values = (0,0,0);
			for (i = 0;  i < len(values); i++) {
				out_values[i] = values[i];
			}`,
			expectValue: []any{float64(1), float64(2), float64(3)},
			name:        "out_values",
		},
		{
			input: `
			foreach (input_values_map as k => v) {
				out_values_map[k] = v;
			}
			`,
			expectValue: map[string]any{"1": 1, "2": 2, "3": 3},
			name:        "out_values_map",
		},
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

		resultMap, errs := statement.Evaluate(paramMap)
		if errs != nil {
			t.Fatal(errs)
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

func TestEvaluableStatementError(t *testing.T) {

	testCases := []struct {
		input       string
		expectValue any
		name        string
		errMsg      string
	}{
		{
			input: `
			sum = 0;
			foreach (values as k ) {
				sum += v; 
			}
			`,
			errMsg: "key or value variable not found",
		},
		{
			input: `out_values = (0,0,0);
			for (i = 0;  i ; i++) {
				out_values[i] = values[i];
			}`,
			errMsg: "invalid for statement expression",
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

		_, errs := statement.Evaluate(paramMap)
		if len(errs) == 0 {
			t.Fatal("not found error")
		}
		found := false
		for _, err := range errs {
			if strings.Contains(err.Error(), tcase.errMsg) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("not found error:%s", tcase.errMsg)
		}

	}
}
