package composite

import (
	"fmt"
	"math"
)

// CurveTerm describes a harmonic in a sampled displacement curve. Attack and
// Release are linear amplitude ramps measured in the same units as Extent.
// Alternate reverses the sign on every other sample for split-row effects.
type CurveTerm struct {
	Amplitude, Frequency, Phase float64
	Attack, Release             float64
	Alternate, Cosine           bool
}

// DeltaCurve compiles a finite waveform into integer displacement differences.
// Step and Extent are degrees when Degrees is true, otherwise radians. Drift
// adds total forward motion over the curve; the cumulative offset is rounded
// using -floor(wave-drift), preserving integer scanline motion.
// OmitLastStep retains a compatibility endpoint that stops before Extent-Step.
type DeltaCurve struct {
	Step, Extent, Drift   float64
	Degrees, OmitLastStep bool
	Terms                 []CurveTerm
}

func (c DeltaCurve) Compile() ([]int, error) {
	finite := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
	if !finite(c.Step) || !finite(c.Extent) || !finite(c.Drift) || c.Step <= 0 || c.Extent <= 0 || c.Extent/c.Step > 1<<20 {
		return nil, fmt.Errorf("composite: invalid displacement curve dimensions")
	}
	for _, t := range c.Terms {
		if !finite(t.Amplitude) || !finite(t.Frequency) || !finite(t.Phase) || !finite(t.Attack) || !finite(t.Release) || t.Attack < 0 || t.Release < 0 {
			return nil, fmt.Errorf("composite: invalid curve term")
		}
	}
	limit := c.Extent
	if c.OmitLastStep {
		limit -= c.Step
	}
	values := make([]float64, 0, int(math.Ceil(c.Extent/c.Step)))
	for position := 0.0; position < limit; position += c.Step {
		angle := position
		if c.Degrees {
			angle = position * math.Pi / 180
		}
		sum := 0.0
		for _, term := range c.Terms {
			amplitude := term.Amplitude
			if term.Attack > 0 && position < term.Attack {
				amplitude *= position / term.Attack
			} else if term.Release > 0 && position > c.Extent-term.Release {
				amplitude *= (c.Extent - position) / term.Release
			}
			if term.Alternate && len(values)%2 == 1 {
				amplitude = -amplitude
			}
			phase := angle*term.Frequency + term.Phase
			wave := math.Sin(phase)
			if term.Cosine {
				wave = math.Cos(phase)
			}
			sum += amplitude * wave
		}
		values = append(values, sum)
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("composite: displacement curve has no samples")
	}
	result := make([]int, len(values))
	drift := 0.0
	previous := 0
	for i, value := range values {
		rounded := -int(math.Floor(value - drift))
		result[i] = rounded - previous
		previous = rounded
		drift += c.Drift / float64(len(values))
	}
	return result, nil
}

// JoinDeltaCurves makes one cumulative sequence, maintaining displacement
// continuity between independently authored curves. Empty order is permitted.
func JoinDeltaCurves(curves [][]int, order []int) ([]int, error) {
	count := 0
	for _, index := range order {
		if index < 0 || index >= len(curves) {
			return nil, fmt.Errorf("composite: curve index out of range")
		}
		if len(curves[index]) > 1<<24-count {
			return nil, fmt.Errorf("composite: displacement sequence too large")
		}
		count += len(curves[index])
	}
	values := make([]int, 0, count)
	position := 0
	for _, index := range order {
		for _, delta := range curves[index] {
			position += delta
			values = append(values, position)
		}
	}
	return values, nil
}

// CumulativeAt repeats a cumulative integer sequence while preserving its last
// offset as forward drift per cycle. Negative positions wrap using floor division.
func CumulativeAt(values []int, index, offset int) int {
	if len(values) == 0 {
		return offset
	}
	cycles, position := index/len(values), index%len(values)
	if position < 0 {
		position += len(values)
		cycles--
	}
	return offset + cycles*values[len(values)-1] + values[position]
}

// FillCumulative samples adjacent positions with one division/modulo for the
// entire row batch. It accepts caller-owned storage and allocates nothing.
func FillCumulative(dst, values []int, start, offset int) {
	if len(values) == 0 {
		for i := range dst {
			dst[i] = offset
		}
		return
	}
	cycles, position := start/len(values), start%len(values)
	if position < 0 {
		position += len(values)
		cycles--
	}
	offset += cycles * values[len(values)-1]
	for i := range dst {
		dst[i] = offset + values[position]
		position++
		if position == len(values) {
			position = 0
			offset += values[len(values)-1]
		}
	}
}

// DisplacementProgram combines a one-time introduction with a repeated main
// displacement sequence. It can drive rows, columns, sprites or a text cursor;
// the consumer decides how sample indices correspond to its deterministic clock.
type DisplacementProgram struct {
	intro, loop []int
	introEnd    int
}

func NewDisplacementProgram(intro, loop []int) (*DisplacementProgram, error) {
	if len(loop) == 0 {
		return nil, fmt.Errorf("composite: empty displacement loop")
	}
	p := &DisplacementProgram{intro: append([]int(nil), intro...), loop: append([]int(nil), loop...)}
	if len(intro) > 0 {
		p.introEnd = intro[len(intro)-1]
	}
	return p, nil
}
func (p *DisplacementProgram) At(index int) int {
	if index >= 0 && index < len(p.intro) {
		return p.intro[index]
	}
	return CumulativeAt(p.loop, index-len(p.intro), p.introEnd)
}
func (p *DisplacementProgram) Fill(dst []int, start int) {
	if start < 0 {
		n := len(dst)
		if start >= -len(dst) {
			n = -start
		}
		FillCumulative(dst[:n], p.loop, start-len(p.intro), p.introEnd)
		dst = dst[n:]
		start += n
		if len(dst) == 0 {
			return
		}
	}
	n := 0
	if start >= 0 && start < len(p.intro) {
		n = copy(dst, p.intro[start:])
	}
	FillCumulative(dst[n:], p.loop, start+n-len(p.intro), p.introEnd)
}
