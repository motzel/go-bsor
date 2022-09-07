package buffer

type value interface {
	float32 | float64 | byte | int8 | int16 | int32 | int64
}

type sum interface {
	float64 | int64
}

type Buffer[T value, S sum] struct {
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

func NewBuffer[T value, S sum](length int) Buffer[T, S] {
	return Buffer[T, S]{Values: make([]T, 0, length), Sum: 0}
}
