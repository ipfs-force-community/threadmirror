package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToIntSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []int
	}{
		{
			name:     "normal case with integers",
			input:    []interface{}{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "empty slice",
			input:    []interface{}{},
			expected: []int{},
		},
		{
			name:     "mixed types with integers and strings",
			input:    []interface{}{1, "hello", 3, "world", 5},
			expected: []int{1, 3, 5},
		},
		{
			name:     "mixed types with integers and floats",
			input:    []interface{}{1, 2.5, 3, 4.7, 5},
			expected: []int{1, 3, 5},
		},
		{
			name:     "all non-integer types",
			input:    []interface{}{"hello", 2.5, true, nil},
			expected: []int{},
		},
		{
			name:     "single integer",
			input:    []interface{}{42},
			expected: []int{42},
		},
		{
			name:     "single non-integer",
			input:    []interface{}{"hello"},
			expected: []int{},
		},
		{
			name:     "negative integers",
			input:    []interface{}{-1, -2, 0, 1, 2},
			expected: []int{-1, -2, 0, 1, 2},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "string input",
			input:    "not a slice",
			expected: nil,
		},
		{
			name:     "integer input",
			input:    42,
			expected: nil,
		},
		{
			name:     "float input",
			input:    3.14,
			expected: nil,
		},
		{
			name:     "boolean input",
			input:    true,
			expected: nil,
		},
		{
			name:     "map input",
			input:    map[string]int{"key": 1},
			expected: nil,
		},
		{
			name:     "slice of strings",
			input:    []string{"a", "b", "c"},
			expected: nil,
		},
		{
			name:     "slice of ints (not interface{})",
			input:    []int{1, 2, 3},
			expected: nil,
		},
		{
			name:     "zero value integer",
			input:    []interface{}{0},
			expected: []int{0},
		},
		{
			name:     "large integers",
			input:    []interface{}{1000000, 2000000, 3000000},
			expected: []int{1000000, 2000000, 3000000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toIntSlice(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestToIntSlice_EdgeCases(t *testing.T) {
	t.Run("slice with nil elements", func(t *testing.T) {
		input := []interface{}{1, nil, 3, nil, 5}
		expected := []int{1, 3, 5}
		result := toIntSlice(input)
		assert.Equal(t, expected, result)
	})

	t.Run("slice with nil elements", func(t *testing.T) {
		m := make(map[string]interface{})
		m["indices"] = []interface{}{1, 3, 5}

		expected := []int{1, 3, 5}
		result := toIntSlice(m["indices"])
		assert.Equal(t, expected, result)
	})

	t.Run("slice with pointers to interface{}", func(t *testing.T) {
		// This test case covers the debug scenario where indices contains pointers
		// to interface{} values rather than direct values
		var val1 interface{} = 1
		var val2 interface{} = 3
		input := []interface{}{&val1, &val2}

		// Now the function should properly handle pointers and return [1, 3]
		result := toIntSlice(input)
		expected := []int{1, 3}
		assert.Equal(t, expected, result)
	})

	t.Run("slice with mixed pointers and direct values", func(t *testing.T) {
		var val1 interface{} = 2
		input := []interface{}{1, &val1, 3}

		// Now both direct values and pointers should work
		result := toIntSlice(input)
		assert.Equal(t, []int{1, 2, 3}, result) // All values should be extracted
	})

	t.Run("slice with pointers to non-integer values", func(t *testing.T) {
		var val1 interface{} = "hello"
		var val2 interface{} = 42
		input := []interface{}{&val1, &val2}

		result := toIntSlice(input)
		assert.Equal(t, []int{42}, result) // Only the integer value should be extracted
	})

	t.Run("slice with complex numbers", func(t *testing.T) {
		input := []interface{}{1, 2 + 3i, 3}
		expected := []int{1, 3}
		result := toIntSlice(input)
		assert.Equal(t, expected, result)
	})

	t.Run("very large slice", func(t *testing.T) {
		input := make([]interface{}, 1000)
		expected := make([]int, 1000)
		for i := 0; i < 1000; i++ {
			input[i] = i
			expected[i] = i
		}
		result := toIntSlice(input)
		assert.Equal(t, expected, result)
	})
}
