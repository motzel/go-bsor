package utils

import "github.com/motzel/go-bsor/buffer"

func SliceMap[T any, S any](data []T, f func(T) S) []S {
	mapped := make([]S, len(data))

	for i, e := range data {
		mapped[i] = f(e)
	}

	return mapped
}

func SliceMin[T buffer.Value](data []T) T {
	min := T(0)

	for i, val := range data {
		if i == 0 || val < min {
			min = val
		}
	}

	return min
}

func SliceMax[T buffer.Value](data []T) T {
	max := T(0)

	for i, val := range data {
		if i == 0 || val > max {
			max = val
		}
	}

	return max
}
