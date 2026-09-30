package motion

import (
	"fmt"
	"math"
)

// FixedPoint holds exactly represented integer coordinates as float64 values.
// It is independent of renderer types and can be adapted to a contour mapper.
type FixedPoint struct{ X, Y float64 }

// TablePolarConfig maps an angular table sample and an integer radius to each
// output axis. Wave must contain a power-of-two number of signed samples.
// Phases select the sample at (phase-angle) modulo the table length. Shift
// applies an arithmetic right shift after multiplication, retaining negative
// integer rounding. WordBits optionally wraps signed coordinates before and
// after adding Center; zero retains the full integer result.
type TablePolarConfig struct {
	Wave     []int16
	Shift    uint8
	Center   [2]int64
	Phases   [2]int
	WordBits uint8
}

// TablePolar is an immutable contour mapping. It copies the input wave once;
// Point does not allocate and may be called concurrently.
type TablePolar struct {
	wave     []int16
	mask     uint64
	shift    uint8
	center   [2]int64
	phases   [2]uint64
	wordBits uint8
}

func NewTablePolar(c TablePolarConfig) (*TablePolar, error) {
	count := len(c.Wave)
	if count == 0 || count > 65536 || count&(count-1) != 0 || c.Shift > 31 ||
		!validContourCenter(c.Center, c.WordBits) {
		return nil, fmt.Errorf("motion: invalid table polar mapping")
	}
	return &TablePolar{
		wave: append([]int16(nil), c.Wave...), mask: uint64(count - 1),
		shift: c.Shift, center: c.Center,
		phases:   [2]uint64{uint64(c.Phases[0]), uint64(c.Phases[1])},
		wordBits: c.WordBits,
	}, nil
}

// Point maps angle and radius without changing either input. Invalid or
// overflowing arithmetic returns false rather than a partially mapped point.
// The resulting integers must be exactly representable by FixedPoint.
func (p *TablePolar) Point(angle, radius int64) (FixedPoint, bool) {
	if p == nil || len(p.wave) == 0 {
		return FixedPoint{}, false
	}
	var axes [2]int64
	for axis := range axes {
		index := (p.phases[axis] - uint64(angle)) & p.mask
		product, ok := contourMultiply(int64(p.wave[index]), radius)
		if !ok {
			return FixedPoint{}, false
		}
		axes[axis] = product >> p.shift
	}
	return contourPosition(axes, p.center, p.wordBits)
}

// RationalGridConfig describes a perspective mapping with independent affine
// numerators and one affine denominator. Each triplet is {base, row, column}:
// value = base + row*inputRow + column*inputColumn. Division truncates toward
// zero, as integer division does. WordBits has the same signed-wrap behavior
// as TablePolarConfig. Center components must be within +/-2^53.
type RationalGridConfig struct {
	X, Y, Denominator [3]int64
	Center            [2]int64
	WordBits          uint8
}

// RationalGrid is an immutable integer perspective mapping. Point checks each
// multiplication, addition and division; no table or per-point allocation is
// required, and negative denominators are supported.
type RationalGrid struct {
	x, y, denominator [3]int64
	center            [2]int64
	wordBits          uint8
}

func NewRationalGrid(c RationalGridConfig) (*RationalGrid, error) {
	if !validContourCenter(c.Center, c.WordBits) || c.Denominator == [3]int64{} {
		return nil, fmt.Errorf("motion: invalid rational grid mapping")
	}
	return &RationalGrid{
		x: c.X, y: c.Y, denominator: c.Denominator,
		center: c.Center, wordBits: c.WordBits,
	}, nil
}

// Point returns false for a zero denominator or overflowing intermediate
// arithmetic. A false result is suitable for omitting an entire contour.
func (p *RationalGrid) Point(column, row int64) (FixedPoint, bool) {
	if p == nil {
		return FixedPoint{}, false
	}
	denominator, ok := contourAffine(p.denominator, column, row)
	if !ok || denominator == 0 {
		return FixedPoint{}, false
	}
	var axes [2]int64
	for axis, coefficients := range [2][3]int64{p.x, p.y} {
		numerator, valid := contourAffine(coefficients, column, row)
		if !valid || numerator == math.MinInt64 && denominator == -1 {
			return FixedPoint{}, false
		}
		axes[axis] = numerator / denominator
	}
	return contourPosition(axes, p.center, p.wordBits)
}

const contourExactInteger = int64(1 << 53)

func validContourCenter(center [2]int64, bits uint8) bool {
	return bits <= 32 && center[0] >= -contourExactInteger && center[0] <= contourExactInteger &&
		center[1] >= -contourExactInteger && center[1] <= contourExactInteger
}

func contourPosition(axes, center [2]int64, bits uint8) (FixedPoint, bool) {
	for axis := range axes {
		value, ok := contourAdd(contourSignedWord(axes[axis], bits), center[axis])
		if !ok {
			return FixedPoint{}, false
		}
		value = contourSignedWord(value, bits)
		if value < -contourExactInteger || value > contourExactInteger {
			return FixedPoint{}, false
		}
		axes[axis] = value
	}
	return FixedPoint{X: float64(axes[0]), Y: float64(axes[1])}, true
}

func contourSignedWord(value int64, bits uint8) int64 {
	if bits == 0 {
		return value
	}
	shift := 64 - bits
	return value << shift >> shift
}

func contourAffine(coefficients [3]int64, column, row int64) (int64, bool) {
	rowValue, ok := contourMultiply(coefficients[1], row)
	if !ok {
		return 0, false
	}
	value, ok := contourAdd(coefficients[0], rowValue)
	if !ok {
		return 0, false
	}
	columnValue, ok := contourMultiply(coefficients[2], column)
	if !ok {
		return 0, false
	}
	return contourAdd(value, columnValue)
}

func contourMultiply(a, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	if a > 0 {
		if b > 0 {
			if a > math.MaxInt64/b {
				return 0, false
			}
		} else if b < math.MinInt64/a {
			return 0, false
		}
	} else {
		if b > 0 {
			if a < math.MinInt64/b {
				return 0, false
			}
		} else if a < math.MaxInt64/b {
			return 0, false
		}
	}
	return a * b, true
}

func contourAdd(a, b int64) (int64, bool) {
	if b > 0 && a > math.MaxInt64-b || b < 0 && a < math.MinInt64-b {
		return 0, false
	}
	return a + b, true
}
