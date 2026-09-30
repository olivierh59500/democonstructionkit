package motion

import "fmt"

// WordMiddleAssociation chooses where a nested signed high-word truncation
// occurs. Reordering these products changes a quantized Euler matrix.
type WordMiddleAssociation uint8

const (
	WordMiddleYFirst WordMiddleAssociation = iota
	WordMiddleXFirst
)

// WordEulerMatrixConfig uses the same copied phase bank as WordEulerVelocity.
// ProductShift scales high-word products; ThirdAxisShift scales the direct
// sine entry in the final row. All arithmetic uses wrapping words/longwords.
type WordEulerMatrixConfig struct {
	Sines                        []int16
	Period, Quantum, Quarter     uint16
	PhasePolicy                  WordVelocityPhase
	ProductShift, ThirdAxisShift uint8
	MiddleAssociation            WordMiddleAssociation
}

type WordEulerMatrix struct {
	phase                    *WordEulerVelocity
	productShift, thirdShift uint8
	middle                   WordMiddleAssociation
}

// SampleWordEulerMatrix reads an immutable caller bank without copying it. It
// suits offline oracles/importers; animated effects should own a compiled matrix.
func SampleWordEulerMatrix(c WordEulerMatrixConfig, angles [3]int16) ([9]int16, bool) {
	if c.Period == 0 || c.Period > 32768 || c.Quantum == 0 || c.Quantum > c.Period || c.Period%c.Quantum != 0 || len(c.Sines) != int(c.Period/c.Quantum) || c.Quarter >= c.Period || c.PhasePolicy > WordPhaseGuarded || c.ProductShift > 31 || c.ThirdAxisShift > 15 || c.MiddleAssociation > WordMiddleXFirst {
		return [9]int16{}, false
	}
	phase := WordEulerVelocity{config: WordEulerVelocityConfig{Sines: c.Sines, Period: c.Period, Quantum: c.Quantum, Quarter: c.Quarter, PhasePolicy: c.PhasePolicy}}
	m := WordEulerMatrix{phase: &phase, productShift: c.ProductShift, thirdShift: c.ThirdAxisShift, middle: c.MiddleAssociation}
	return m.Sample(angles)
}

func NewWordEulerMatrix(c WordEulerMatrixConfig) (*WordEulerMatrix, error) {
	if c.ProductShift > 31 || c.ThirdAxisShift > 15 || c.MiddleAssociation > WordMiddleXFirst {
		return nil, fmt.Errorf("motion: invalid word matrix scaling or association")
	}
	p, err := NewWordEulerVelocity(WordEulerVelocityConfig{Sines: c.Sines, Period: c.Period, Quantum: c.Quantum, Quarter: c.Quarter, PhasePolicy: c.PhasePolicy})
	if err != nil {
		return nil, err
	}
	return &WordEulerMatrix{phase: p, productShift: c.ProductShift, thirdShift: c.ThirdAxisShift, middle: c.MiddleAssociation}, nil
}

// Sample returns a column-major signed matrix, without advancing any clock.
func (m *WordEulerMatrix) Sample(angles [3]int16) ([9]int16, bool) {
	if m == nil {
		return [9]int16{}, false
	}
	var wave [6]int16
	for i, a := range angles {
		s, c, ok := m.phase.phase(a)
		if !ok {
			return [9]int16{}, false
		}
		wave[i*2], wave[i*2+1] = s, c
	}
	sx, cx, sy, cy, sz, cz := wave[0], wave[1], wave[2], wave[3], wave[4], wave[5]
	mul := func(a, b int16) int32 { return int32(a) * int32(b) }
	high := func(v int32) int16 { return int16(v >> 16) }
	truncate := func(v int32) int16 { return int16(int32(high(v)) >> m.productShift) }
	middle := mul(high(mul(sz, sy)<<1), sx)
	if m.middle == WordMiddleXFirst {
		middle = mul(high(mul(sz, sx)<<1), sy)
	}
	return [9]int16{
		truncate(mul(cx, cy)), truncate(mul(sx, cz) - mul(high(mul(sz, sy)<<1), cx)),
		truncate(mul(sx, sz) + mul(high(mul(cz, sy)<<1), cx)),
		-truncate(mul(sx, cy)), truncate(mul(cx, cz) + middle),
		truncate(mul(cx, sz) - mul(high(mul(cz, sy)<<1), sx)),
		-(sy >> m.thirdShift), -truncate(mul(cy, sz)), truncate(mul(cy, cz)),
	}, true
}
