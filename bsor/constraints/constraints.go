package constraints

type Float interface {
	float32 | float64
}

type Signed interface {
	int8 | int16 | int32 | int64
}

type Unsigned interface {
	byte | uint16 | int32 | uint32 | uint64
}

type HighestPrecision interface {
	float64 | int64
}

type NumericValue interface {
	Float | Signed | Unsigned
}

type Sum interface {
	HighestPrecision
}
