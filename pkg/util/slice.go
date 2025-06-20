package util

func MergeSlices[T any](slices ...[]T) []T {
	// Calculate total length for pre-allocation
	totalLen := 0
	for _, s := range slices {
		totalLen += len(s)
	}

	// Pre-allocate with exact capacity for better performance
	result := make([]T, 0, totalLen)

	// Append all slices efficiently
	for _, s := range slices {
		result = append(result, s...)
	}

	return result
}
