package buffer

import (
	"sort"
)

type Value interface {
	float32 | float64 | byte | int8 | int16 | uint16 | int32 | uint32 | int64 | uint64
}

type sum interface {
	float64 | int64
}

type Buffer[T Value, S sum] struct {
	Values []T
	Sum    S
}

func (buffer *Buffer[T, S]) Add(value T) {
	buffer.Values = append(buffer.Values, value)
	buffer.Sum += S(value)
}

func (buffer *Buffer[T, S]) Avg() float64 {
	length := len(buffer.Values)

	if length > 0 {
		return float64(buffer.Sum) / float64(length)
	} else {
		return 0
	}
}

func sortSlice[T Value](s []T) {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}

func (buffer *Buffer[T, S]) Median() T {
	length := len(buffer.Values)
	if length == 0 {
		return T(0)
	}

	sortSlice(buffer.Values)

	if length%2 == 0 {
		return (buffer.Values[length/2-1] + buffer.Values[length/2]) / 2
	} else {
		return buffer.Values[length/2]
	}
}

func NewBuffer[T Value, S sum](length int) Buffer[T, S] {
	return Buffer[T, S]{Values: make([]T, 0, length), Sum: 0}
}
