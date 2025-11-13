package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTernary(t *testing.T) {
	testCases := []struct {
		name      string
		condition bool
		trueVal   any
		falseVal  any
		expected  any
	}{
		// --- String Type ---
		{
			name:      "string with true condition",
			condition: true,
			trueVal:   "hello",
			falseVal:  "world",
			expected:  "hello",
		},
		{
			name:      "string with false condition",
			condition: false,
			trueVal:   "hello",
			falseVal:  "world",
			expected:  "world",
		},

		// --- Integer Type ---
		{
			name:      "int with true condition",
			condition: true,
			trueVal:   100,
			falseVal:  200,
			expected:  100,
		},
		{
			name:      "int with false condition",
			condition: false,
			trueVal:   100,
			falseVal:  200,
			expected:  200,
		},

		// --- Float Type ---
		{
			name:      "float64 with true condition",
			condition: true,
			trueVal:   3.14,
			falseVal:  -1.0,
			expected:  3.14,
		},

		// --- Boolean Type ---
		{
			name:      "bool with false condition",
			condition: false,
			trueVal:   true,
			falseVal:  false,
			expected:  false,
		},

		// --- Struct Type ---
		{
			name:      "struct with true condition",
			condition: true,
			trueVal:   struct{ name string }{"Alice"},
			falseVal:  struct{ name string }{"Bob"},
			expected:  struct{ name string }{"Alice"},
		},

		// --- Pointer Type ---
		{
			name:      "pointer with false condition",
			condition: false,
			trueVal:   &struct{ name string }{"Alice"},
			falseVal:  &struct{ name string }{"Bob"},
			expected:  &struct{ name string }{"Bob"},
		},

		// --- Nil Value ---
		// {
		// 	name:      "nil pointer with false condition",
		// 	condition: false,
		// 	trueVal:   &struct{ name string }{"Not Nil"},
		// 	falseVal:  nil,
		// 	expected:  nil,
		// },
		// {
		// 	name:      "nil pointer with true condition",
		// 	condition: true,
		// 	trueVal:   nil,
		// 	falseVal:  &struct{ name string }{"Not Nil"},
		// 	expected:  nil,
		// },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result any
			switch v := tc.trueVal.(type) {
			case string:
				result = Ternary(tc.condition, v, tc.falseVal.(string))
			case int:
				result = Ternary(tc.condition, v, tc.falseVal.(int))
			case float64:
				result = Ternary(tc.condition, v, tc.falseVal.(float64))
			case bool:
				result = Ternary(tc.condition, v, tc.falseVal.(bool))
			case struct{ name string }:
				result = Ternary(tc.condition, v, tc.falseVal.(struct{ name string }))
			case *struct{ name string }:
				result = Ternary(tc.condition, v, tc.falseVal.(*struct{ name string }))
			case nil:
				result = Ternary(tc.condition, nil, tc.falseVal.(*struct{ name string }))
			default:
				t.Fatalf("Unhandled type in test case: %T", v)
			}

			assert.Equal(t, tc.expected, result, fmt.Sprintf("Failed on test case: %s", tc.name))
		})
	}
}
