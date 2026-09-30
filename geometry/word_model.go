package geometry

import (
	"fmt"
	"math"

	"github.com/olivierh59500/democonstructionkit/motion"
)

type WordDivisionOverflow uint8

const (
	WordDivisionKeepLow WordDivisionOverflow = iota
	WordDivisionClamp
	WordDivisionReject
)

type WordZeroDepth uint8

const (
	WordZeroDepthError WordZeroDepth = iota
	WordZeroDepthHide
)

// WordProjectionConfig keeps truncation and division policies separate from
// artwork. Translation is applied to longword X/Y numerators before division.
// Depth combines a shifted transformed Z word, DepthBias and pose.Depth. Center
// additions wrap to signed words. Division overflow can preserve the low word,
// clamp it, or reject the projection.
type WordProjectionConfig struct {
	DepthShift, TranslationShift uint8
	DepthBias                    int16
	Center                       [2]int16
	Overflow                     WordDivisionOverflow
	ZeroDepth                    WordZeroDepth
}

type WordPose struct {
	Matrix      [9]int16
	Translation [2]int16
	Depth       int16
}
type WordScreenPoint struct {
	X, Y    int16
	Visible bool
}
type WordModelConfig struct {
	Points     []motion.WrappedPoint
	Projection WordProjectionConfig
}

// WordModel owns copied geometry and bounded cached projection banks. Identical
// poses do not reproject. A failed pose leaves the preceding frame untouched.
type WordModel struct {
	points             []motion.WrappedPoint
	projected, scratch []WordScreenPoint
	projection         WordProjectionConfig
	pose               WordPose
	valid              bool
}

func ValidateWordProjection(c WordProjectionConfig) error {
	if c.DepthShift > 31 || c.TranslationShift > 31 || c.Overflow > WordDivisionReject || c.ZeroDepth > WordZeroDepthHide {
		return fmt.Errorf("geometry: invalid word projection policy")
	}
	return nil
}
func NewWordModel(c WordModelConfig) (*WordModel, error) {
	if len(c.Points) < 1 || len(c.Points) > 65536 {
		return nil, fmt.Errorf("geometry: invalid word model population")
	}
	if err := ValidateWordProjection(c.Projection); err != nil {
		return nil, err
	}
	return &WordModel{points: append([]motion.WrappedPoint(nil), c.Points...), projected: make([]WordScreenPoint, len(c.Points)), scratch: make([]WordScreenPoint, len(c.Points)), projection: c.Projection}, nil
}

// ProjectWordPoint retains signed 32-bit product/sum wrapping and signed word
// depth. Its scalar form also supports importers and independent CPU oracles.
func ProjectWordPoint(p motion.WrappedPoint, pose WordPose, c WordProjectionConfig) (WordScreenPoint, error) {
	if err := ValidateWordProjection(c); err != nil {
		return WordScreenPoint{}, err
	}
	x, y, z := int32(p.X), int32(p.Y), int32(p.Z)
	m := pose.Matrix
	px := x*int32(m[0]) + y*int32(m[3]) + z*int32(m[6])
	py := x*int32(m[1]) + y*int32(m[4]) + z*int32(m[7])
	pz := x*int32(m[2]) + y*int32(m[5]) + z*int32(m[8])
	depth := int16(uint16(int16(pz>>c.DepthShift)) + uint16(c.DepthBias) + uint16(pose.Depth))
	if depth == 0 {
		if c.ZeroDepth == WordZeroDepthHide {
			return WordScreenPoint{}, nil
		}
		return WordScreenPoint{}, fmt.Errorf("geometry: word projection divides by zero")
	}
	px += int32(pose.Translation[0]) << c.TranslationShift
	py += int32(pose.Translation[1]) << c.TranslationShift
	divide := func(n int32) (int16, error) {
		q := int64(n) / int64(depth)
		if q < math.MinInt16 || q > math.MaxInt16 {
			switch c.Overflow {
			case WordDivisionKeepLow:
				return int16(n), nil
			case WordDivisionClamp:
				return int16(max(math.MinInt16, min(math.MaxInt16, q))), nil
			default:
				return 0, fmt.Errorf("geometry: word quotient overflows")
			}
		}
		return int16(q), nil
	}
	a, err := divide(px)
	if err != nil {
		return WordScreenPoint{}, err
	}
	b, err := divide(py)
	if err != nil {
		return WordScreenPoint{}, err
	}
	return WordScreenPoint{X: int16(uint16(a) + uint16(c.Center[0])), Y: int16(uint16(b) + uint16(c.Center[1])), Visible: true}, nil
}

func (m *WordModel) Project(pose WordPose) error {
	if m == nil {
		return fmt.Errorf("geometry: absent word model")
	}
	if m.valid && pose == m.pose {
		return nil
	}
	for i, p := range m.points {
		point, err := ProjectWordPoint(p, pose, m.projection)
		if err != nil {
			return err
		}
		m.scratch[i] = point
	}
	m.projected, m.scratch = m.scratch, m.projected
	m.pose, m.valid = pose, true
	return nil
}

// Points returns the borrowed immutable projected bank until the next Project.
func (m *WordModel) Points() []WordScreenPoint {
	if m == nil {
		return nil
	}
	return m.projected
}

// WordWinding preserves subtraction as signed words and the cross product as a
// wrapping signed longword. Nonnegative means front-facing for this convention.
func WordWinding(a, b, c WordScreenPoint) int32 {
	return int32(b.X-a.X)*int32(c.Y-a.Y) - int32(b.Y-a.Y)*int32(c.X-a.X)
}
