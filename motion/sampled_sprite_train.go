package motion

import (
	"fmt"
	"math"
)

// SampledSpriteTrainConfig follows two authored coordinate tables with indexed
// spacing and an optional extra delay after one sprite. WrapAfterLength keeps
// the final path index visible for one frame before returning to zero.
type SampledSpriteTrainConfig struct {
	PathX, PathY             []float64
	Count                    int
	StartCursor              int
	Spacing                  int
	ExtraAfter, ExtraOffset  int
	WrapAfterLength          bool
	XAmplitude, YAmplitude   float64
	XRate, YRate             float64
	XIndexPhase, YIndexPhase float64
	XCos, YCos               bool
	Position                 func(cursor, index int, pathX, pathY float64) (x, y float64)
}

type SampledSpritePose struct {
	Index, Frame, SampleIndex int
	X, Y                      float64
}

// SampledSpriteTrain uses one copy of each coordinate table. Wrapped indexed
// lookup produces the same samples as duplicating both source arrays, while
// reducing per-instance memory for long authored paths.
type SampledSpriteTrain struct {
	config SampledSpriteTrainConfig
	cursor int
	poses  []SampledSpritePose
}

func NewSampledSpriteTrain(c SampledSpriteTrainConfig) (*SampledSpriteTrain, error) {
	period := len(c.PathX)
	if period < 1 || period > 1<<20 || len(c.PathY) != period ||
		c.Count < 1 || c.Count > 1<<16 || c.StartCursor < 0 || c.StartCursor > period {
		return nil, fmt.Errorf("motion: invalid sampled sprite path or population")
	}
	for _, value := range [...]float64{c.XAmplitude, c.YAmplitude, c.XRate, c.YRate,
		c.XIndexPhase, c.YIndexPhase} {
		if !finiteSampledTrain(value) {
			return nil, fmt.Errorf("motion: nonfinite sampled sprite wave")
		}
	}
	for i := range c.PathX {
		if !finiteSampledTrain(c.PathX[i]) || !finiteSampledTrain(c.PathY[i]) {
			return nil, fmt.Errorf("motion: nonfinite sampled sprite position")
		}
	}
	c.PathX = append([]float64(nil), c.PathX...)
	c.PathY = append([]float64(nil), c.PathY...)
	train := &SampledSpriteTrain{config: c, poses: make([]SampledSpritePose, c.Count)}
	train.Reset()
	return train, nil
}

func finiteSampledTrain(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func (t *SampledSpriteTrain) Reset() {
	t.cursor = t.config.StartCursor
	clear(t.poses)
}

// Step advances the source cursor before preparing every sprite pose.
func (t *SampledSpriteTrain) Step() {
	period := len(t.config.PathX)
	t.cursor++
	if t.config.WrapAfterLength {
		if t.cursor > period {
			t.cursor = 0
		}
	} else if t.cursor >= period {
		t.cursor = 0
	}
	for i := range t.poses {
		raw := t.cursor + i*t.config.Spacing
		if i >= t.config.ExtraAfter {
			raw += t.config.ExtraOffset
		}
		index := raw % period
		if index < 0 {
			index += period
		}
		x, y := t.config.PathX[index], t.config.PathY[index]
		if t.config.Position != nil {
			x, y = t.config.Position(t.cursor, i, x, y)
		} else {
			phaseX := float64(t.cursor)*t.config.XRate + float64(i)*t.config.XIndexPhase
			phaseY := float64(t.cursor)*t.config.YRate + float64(i)*t.config.YIndexPhase
			if t.config.XCos {
				x += t.config.XAmplitude * math.Cos(phaseX)
			} else {
				x += t.config.XAmplitude * math.Sin(phaseX)
			}
			if t.config.YCos {
				y += t.config.YAmplitude * math.Cos(phaseY)
			} else {
				y += t.config.YAmplitude * math.Sin(phaseY)
			}
		}
		t.poses[i] = SampledSpritePose{Index: i, Frame: i, SampleIndex: raw, X: x, Y: y}
	}
}

func (t *SampledSpriteTrain) Cursor() int                  { return t.cursor }
func (t *SampledSpriteTrain) Samples() []SampledSpritePose { return t.poses }
