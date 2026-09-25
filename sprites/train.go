package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// TrainAxis places one coordinate at Offset and optionally adds either a
// stateful bounce or a phase-spaced sine/cosine wave. A wave uses Frame.Time in
// the caller's chosen units; a bounce advances once per Train.Update.
type TrainAxis struct {
	Offset float64
	Bounce *motion.BounceBankConfig
	Wave   *motion.Wave
}

// TrainConfig describes a reusable train of raster strips, sprites or logos.
// Images are borrowed; Count defaults to their number. Spacing is additional
// screen-space separation independent of each axis's motion.
type TrainConfig struct {
	Images                           []*ebiten.Image
	Count                            int
	X, Y                             TrainAxis
	Spacing                          motion.Point
	ScaleX, ScaleY, AnchorX, AnchorY float64
	Opacity                          float64
	Filter                           ebiten.Filter
	Blend                            ebiten.Blend
	Reverse                          bool
}

// Train owns its motion and prepared image poses but borrows the images.
type Train struct {
	group   *Group
	xBounce *motion.BounceBank
	yBounce *motion.BounceBank
}

func NewTrain(config TrainConfig) (*Train, error) {
	count := config.Count
	if count == 0 {
		count = len(config.Images)
	}
	xBounce, xWave, err := compileTrainAxis(config.X, count)
	if err != nil {
		return nil, fmt.Errorf("sprites: x axis: %w", err)
	}
	yBounce, yWave, err := compileTrainAxis(config.Y, count)
	if err != nil {
		return nil, fmt.Errorf("sprites: y axis: %w", err)
	}
	group, err := NewGroup(GroupConfig{
		Frames: config.Images, Count: count, FrameStride: 1, Speed: 1,
		Spacing: config.Spacing, ScaleX: config.ScaleX, ScaleY: config.ScaleY,
		AnchorX: config.AnchorX, AnchorY: config.AnchorY, Opacity: config.Opacity,
		Filter: config.Filter, Blend: config.Blend, Reverse: config.Reverse,
		Formation: func(time float64, index int) motion.Point {
			point := motion.Point{X: config.X.Offset, Y: config.Y.Offset}
			if xBounce != nil {
				point.X += xBounce.At(index)
			} else if xWave != nil {
				point.X += xWave.At(float64(index), time)
			}
			if yBounce != nil {
				point.Y += yBounce.At(index)
			} else if yWave != nil {
				point.Y += yWave.At(float64(index), time)
			}
			return point
		},
	})
	if err != nil {
		return nil, err
	}
	return &Train{group: group, xBounce: xBounce, yBounce: yBounce}, nil
}

func compileTrainAxis(axis TrainAxis, count int) (*motion.BounceBank, *motion.Wave, error) {
	if math.IsNaN(axis.Offset) || math.IsInf(axis.Offset, 0) || axis.Bounce != nil && axis.Wave != nil {
		return nil, nil, fmt.Errorf("invalid axis configuration")
	}
	if axis.Bounce != nil {
		bank, err := motion.NewBounceBank(*axis.Bounce)
		if err != nil {
			return nil, nil, err
		}
		if bank.Len() != count {
			return nil, nil, fmt.Errorf("bounce count %d differs from image count %d", bank.Len(), count)
		}
		return bank, nil, nil
	}
	if axis.Wave != nil {
		wave := *axis.Wave
		for _, value := range []float64{wave.Amplitude, wave.Spatial, wave.Speed, wave.Phase} {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return nil, nil, fmt.Errorf("nonfinite wave parameter")
			}
		}
		return nil, &wave, nil
	}
	return nil, nil, nil
}

// Update advances bouncing axes and samples all poses once. Draw never advances
// motion, so a train can be drawn into more than one destination per update.
func (train *Train) Update(frame kit.Frame) error {
	if math.IsNaN(frame.Time) || math.IsInf(frame.Time, 0) {
		return fmt.Errorf("sprites: nonfinite train time")
	}
	if train.xBounce != nil {
		train.xBounce.Step()
	}
	if train.yBounce != nil {
		train.yBounce.Step()
	}
	return train.group.Update(frame)
}

func (train *Train) Draw(dst *ebiten.Image) { train.group.Draw(dst) }

// Poses returns borrowed poses for inspection or an alternate renderer.
func (train *Train) Poses() []GroupPose { return train.group.Poses() }

// Reset restores bouncing axes and the initial time-zero pose.
func (train *Train) Reset() error {
	if train.xBounce != nil {
		train.xBounce.Reset()
	}
	if train.yBounce != nil {
		train.yBounce.Reset()
	}
	return train.group.Update(kit.Frame{})
}
