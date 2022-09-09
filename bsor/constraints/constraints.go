package constraints

type NumericValue interface {
	float32 | float64 | byte | int8 | int16 | uint16 | int32 | uint32 | int64 | uint64
}

type Sum interface {
	float64 | int64
}
