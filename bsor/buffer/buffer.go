package buffer

import (
	"github.com/motzel/go-bsor/bsor/constraints"
	"github.com/motzel/go-bsor/bsor/utils"
	"math"
	"sort"
)

type Stats[T constraints.NumericValue] struct {
	Min    T       `json:"min"`
	Avg    float64 `json:"avg"`
	Median T       `json:"med"`
	Max    T       `json:"max"`
}

type StatsSlice[T constraints.NumericValue] struct {
	Min    []T       `json:"min"`
	Avg    []float64 `json:"avg"`
	Median []T       `json:"med"`
	Max    []T       `json:"max"`
}

type Buffer[T constraints.NumericValue, S constraints.Sum] struct {
	values []T
	sum    S
}

func (buffer *Buffer[T, S]) Add(value T) {
	buffer.values = append(buffer.values, value)
	buffer.sum += S(value)
}

func (buffer *Buffer[T, S]) Avg() float64 {
	length := len(buffer.values)

	if length > 0 {
		return float64(buffer.sum) / float64(length)
	} else {
		return 0
	}
}

func (buffer *Buffer[T, S]) Min() T {
	return utils.SliceMin[T](buffer.Values())
}

func (buffer *Buffer[T, S]) Max() T {
	return utils.SliceMax[T](buffer.Values())
}

func (buffer *Buffer[T, S]) Values() []T {
	return buffer.values
}

func (buffer *Buffer[T, S]) Sum() S {
	return buffer.sum
}

func (buffer *Buffer[T, S]) Length() int {
	return len(buffer.values)
}

func sortSlice[T constraints.NumericValue](s []T) {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}

func (buffer *Buffer[T, S]) Median() T {
	length := len(buffer.values)
	if length == 0 {
		return T(0)
	}

	sortSlice(buffer.values)

	if length%2 == 0 {
		return (buffer.values[length/2-1] + buffer.values[length/2]) / 2
	} else {
		return buffer.values[length/2]
	}
}

func (buffer *Buffer[T, S]) Stats() Stats[T] {
	return Stats[T]{
		Min:    buffer.Min(),
		Avg:    buffer.Avg(),
		Median: buffer.Median(),
		Max:    buffer.Max(),
	}
}

func NewBuffer[T constraints.NumericValue, S constraints.Sum](length int) Buffer[T, S] {
	return Buffer[T, S]{values: make([]T, 0, length), sum: 0}
}

type CircularBuffer[T constraints.NumericValue, S constraints.Sum] struct {
	Buffer[T, S]
	position int
	size     int
}

func (buffer *CircularBuffer[T, S]) Add(value T) {
	buffer.values[buffer.position] = value
	buffer.sum += S(value)
	buffer.size++
	buffer.position++

	if buffer.position >= len(buffer.values) {
		buffer.position = 0
	}
}

func (buffer *CircularBuffer[T, S]) Avg() float64 {
	sum, count := buffer.sumAndCount()

	if count > 0 {
		return float64(sum) / float64(count)
	} else {
		return 0
	}
}

func (buffer *CircularBuffer[T, S]) AvgAllTime() float64 {
	if buffer.size > 0 {
		return float64(buffer.sum) / float64(buffer.size)
	} else {
		return 0
	}
}

func (buffer *CircularBuffer[T, S]) Size() int {
	return buffer.size
}

func (buffer *CircularBuffer[T, S]) Sum() S {
	sum, _ := buffer.sumAndCount()

	return sum
}

func (buffer *CircularBuffer[T, S]) SumAllTime() S {
	return buffer.sum
}

func (buffer *CircularBuffer[T, S]) sumAndCount() (S, int) {
	if buffer.size > 0 {
		length := buffer.Length()
		count := int(math.Min(float64(length), float64(buffer.size)))

		sum := S(0)
		for i := 0; i < count; i++ {
			sum += S(buffer.values[i])
		}

		return sum, count
	} else {
		return 0, 0
	}
}

func NewCircularBuffer[T constraints.NumericValue, S constraints.Sum](length int) CircularBuffer[T, S] {
	return CircularBuffer[T, S]{Buffer: Buffer[T, S]{values: make([]T, length), sum: 0}, position: 0, size: 0}
}
