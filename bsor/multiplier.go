package bsor

const MaxMultiplier = 8

type MultiplierCounter struct {
	value    byte
	progress byte
}

func NewMultiplierCounter() *MultiplierCounter {
	multiplier := MultiplierCounter{}

	multiplier.Reset()

	return &multiplier
}

func (multiplier *MultiplierCounter) Reset() byte {
	multiplier.value = 1
	multiplier.progress = 1

	return multiplier.value
}

func (multiplier *MultiplierCounter) Value() byte {
	return multiplier.value
}

func (multiplier *MultiplierCounter) Inc() byte {
	if multiplier.value >= MaxMultiplier {
		return MaxMultiplier
	}

	if multiplier.progress+1 >= multiplier.value*2 {
		multiplier.value *= 2
		multiplier.progress = 0
	} else {
		multiplier.progress++
	}

	return multiplier.value
}

func (multiplier *MultiplierCounter) Dec() byte {
	if multiplier.value > 1 {
		multiplier.value /= 2
	}

	multiplier.progress = 1

	return multiplier.value
}
