package util

import (
	"reflect"
	"testing"
)

func TestMergeSlices(t *testing.T) {
	t.Run("should merge empty slices", func(t *testing.T) {
		var slice1 []int
		var slice2 []int
		result := MergeSlices(slice1, slice2)

		if len(result) != 0 {
			t.Errorf("Expected empty slice, got length %d", len(result))
		}
	})

	t.Run("should merge single slice", func(t *testing.T) {
		slice1 := []int{1, 2, 3}
		result := MergeSlices(slice1)
		expected := []int{1, 2, 3}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("should merge two integer slices", func(t *testing.T) {
		slice1 := []int{1, 2, 3}
		slice2 := []int{4, 5, 6}
		result := MergeSlices(slice1, slice2)
		expected := []int{1, 2, 3, 4, 5, 6}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("should merge multiple integer slices", func(t *testing.T) {
		slice1 := []int{1, 2}
		slice2 := []int{3, 4}
		slice3 := []int{5, 6}
		slice4 := []int{7, 8}
		result := MergeSlices(slice1, slice2, slice3, slice4)
		expected := []int{1, 2, 3, 4, 5, 6, 7, 8}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("should merge string slices", func(t *testing.T) {
		slice1 := []string{"hello", "world"}
		slice2 := []string{"foo", "bar"}
		result := MergeSlices(slice1, slice2)
		expected := []string{"hello", "world", "foo", "bar"}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("should merge slices with different sizes", func(t *testing.T) {
		slice1 := []int{1}
		slice2 := []int{2, 3, 4, 5}
		slice3 := []int{6, 7}
		result := MergeSlices(slice1, slice2, slice3)
		expected := []int{1, 2, 3, 4, 5, 6, 7}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("should handle nil slices", func(t *testing.T) {
		var slice1 []int
		slice2 := []int{1, 2, 3}
		var slice3 []int
		result := MergeSlices(slice1, slice2, slice3)
		expected := []int{1, 2, 3}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("should merge slices with custom struct type", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		slice1 := []Person{
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 25},
		}
		slice2 := []Person{
			{Name: "Charlie", Age: 35},
		}

		result := MergeSlices(slice1, slice2)
		expected := []Person{
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 25},
			{Name: "Charlie", Age: 35},
		}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("should maintain order of elements", func(t *testing.T) {
		slice1 := []int{3, 1, 4}
		slice2 := []int{1, 5, 9}
		slice3 := []int{2, 6}
		result := MergeSlices(slice1, slice2, slice3)
		expected := []int{3, 1, 4, 1, 5, 9, 2, 6}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("should handle no input slices", func(t *testing.T) {
		result := MergeSlices[int]()

		if len(result) != 0 {
			t.Errorf("Expected empty slice, got length %d", len(result))
		}

		if result == nil {
			t.Error("Expected non-nil slice, got nil")
		}
	})

	t.Run("should handle large slices efficiently", func(t *testing.T) {
		// Create large slices to test performance characteristics
		slice1 := make([]int, 1000)
		slice2 := make([]int, 1000)

		for i := 0; i < 1000; i++ {
			slice1[i] = i
			slice2[i] = i + 1000
		}

		result := MergeSlices(slice1, slice2)

		if len(result) != 2000 {
			t.Errorf("Expected length 2000, got %d", len(result))
		}

		// Verify first and last elements
		if result[0] != 0 {
			t.Errorf("Expected first element to be 0, got %d", result[0])
		}

		if result[1999] != 1999 {
			t.Errorf("Expected last element to be 1999, got %d", result[1999])
		}
	})
}
