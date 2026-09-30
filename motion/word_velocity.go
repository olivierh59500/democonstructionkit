package motion

import "fmt"

// WordVelocityPhase defines how signed phase words address an authored sine bank.
type WordVelocityPhase uint8

const (
	// WordPhaseNormalized wraps every signed phase into the positive period,
	// then quantizes it. The quarter phase is applied to that quantized angle.
	WordPhaseNormalized WordVelocityPhase = iota
	// WordPhaseGuarded reads the sine using the unsigned input word; an index
	// outside the bank contributes zero. Its quarter phase wraps the word once
	// using signed comparisons before quantization. If that single correction
	// still leaves the quarter lookup outside the bank, sampling returns false.
	WordPhaseGuarded
)

// WordEulerVelocityConfig supplies the sine bank, angular units and per-axis
// output scaling for three selected Euler matrix products. Period must be an
// exact multiple of Quantum, and Sines must contain Period/Quantum signed words.
// Quarter is the sine-to-cosine phase offset in the same angular units. Shifts
// apply arithmetic right shifts after the final high-word truncation.
type WordEulerVelocityConfig struct {
	Sines       []int16
	Period      uint16
	Quantum     uint16
	Quarter     uint16
	OutputShift [3]uint8
	PhasePolicy WordVelocityPhase
}

// WordEulerVelocity owns an immutable copied sine bank. It evaluates the matrix
// terms cx*cy, sx*cz-sz*sy*cx and sx*sz+cz*sy*cx with the same signed word/long
// truncations throughout. The nested Y/Z products are doubled before retaining
// their high word. Sampling allocates nothing and does not advance angle clocks.
type WordEulerVelocity struct{ config WordEulerVelocityConfig }

func NewWordEulerVelocity(c WordEulerVelocityConfig) (*WordEulerVelocity, error) {
	if c.Period == 0 || c.Period > 32768 || c.Quantum == 0 || c.Quantum > c.Period || c.Period%c.Quantum != 0 ||
		c.Quarter >= c.Period || len(c.Sines) != int(c.Period/c.Quantum) || c.PhasePolicy > WordPhaseGuarded {
		return nil, fmt.Errorf("motion: invalid word velocity phase or sine table")
	}
	for _, shift := range c.OutputShift {
		if shift > 31 {
			return nil, fmt.Errorf("motion: word velocity shift exceeds signed long width")
		}
	}
	c.Sines = append([]int16(nil), c.Sines...)
	return &WordEulerVelocity{config: c}, nil
}

func (v *WordEulerVelocity) phase(angle int16) (sine, cosine int16, ok bool) {
	c := v.config
	if c.PhasePolicy == WordPhaseNormalized {
		phase := int(angle) % int(c.Period)
		if phase < 0 {
			phase += int(c.Period)
		}
		phase -= phase % int(c.Quantum)
		return c.Sines[phase/int(c.Quantum)], c.Sines[((phase+int(c.Quarter))%int(c.Period))/int(c.Quantum)], true
	}
	index := int(uint16(angle) / c.Quantum)
	if index < len(c.Sines) {
		sine = c.Sines[index]
	}
	quarter := uint16(angle) + c.Quarter
	if int16(quarter) > int16(c.Period-c.Quantum) {
		quarter -= c.Period
	} else if int16(quarter) < 0 {
		quarter += c.Period
	}
	index = int(quarter / c.Quantum)
	if index >= len(c.Sines) {
		return 0, 0, false
	}
	return sine, c.Sines[index], true
}

// Sample returns the three velocity words for one caller-owned angular pose.
// False indicates an absent controller or an invalid guarded quarter lookup;
// no unbounded wrapping, array access or floating-point approximation is used.
func (v *WordEulerVelocity) Sample(angles [3]int16) ([3]int16, bool) {
	if v == nil {
		return [3]int16{}, false
	}
	var wave [6]int16
	for axis, angle := range angles {
		sine, cosine, ok := v.phase(angle)
		if !ok {
			return [3]int16{}, false
		}
		wave[axis*2], wave[axis*2+1] = sine, cosine
	}
	sx, cx, sy, cy, sz, cz := wave[0], wave[1], wave[2], wave[3], wave[4], wave[5]
	mul := func(a, b int16) int32 { return int32(a) * int32(b) }
	high := func(value int32) int16 { return int16(value >> 16) }
	return [3]int16{
		int16(int32(high(mul(cx, cy))) >> v.config.OutputShift[0]),
		int16(int32(high(mul(sx, cz)-mul(high(mul(sz, sy)<<1), cx))) >> v.config.OutputShift[1]),
		int16(int32(high(mul(sx, sz)+mul(high(mul(cz, sy)<<1), cx))) >> v.config.OutputShift[2]),
	}, true
}
