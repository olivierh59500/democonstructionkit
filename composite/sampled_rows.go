package composite

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// SampledRowsConfig borrows one image and renders its strips in copy-then-row
// order. A program chooses source row, destination pose and scale on each tick.
// UseTime samples Frame.Time, including any preceding scene introduction.
type SampledRowsConfig struct {
	Image                *ebiten.Image
	Program              motion.SampledRowProgram
	Rows, Copies         int
	SourceX, SourceWidth float64
	StartTime, TimeStep  float64
	UseTime              bool
	Filter               ebiten.Filter
	Blend                ebiten.Blend
}

// SampledRows owns only two bounded pose banks and no GPU surface. Draw reuses
// the selected crops without advancing time or changing the borrowed image.
type SampledRows struct {
	config  SampledRowsConfig
	poses   []motion.SampledRowPose
	scratch []motion.SampledRowPose
	time    float64
}

func NewSampledRows(c SampledRowsConfig) (*SampledRows, error) {
	if c.Image == nil || c.Program == nil || c.Rows < 1 || c.Copies < 1 ||
		c.Rows > 4096 || c.Copies > 256 || c.Rows*c.Copies > 16_384 || c.SourceWidth <= 0 {
		return nil, fmt.Errorf("composite: invalid sampled rows dimensions or source")
	}
	for _, value := range [...]float64{c.SourceX, c.SourceWidth, c.StartTime, c.TimeStep} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("composite: nonfinite sampled rows configuration")
		}
	}
	r := &SampledRows{config: c, poses: make([]motion.SampledRowPose, c.Rows*c.Copies),
		scratch: make([]motion.SampledRowPose, c.Rows*c.Copies), time: c.StartTime}
	if err := r.sample(c.StartTime); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *SampledRows) Update(frame kit.Frame) error {
	if r == nil {
		return fmt.Errorf("composite: nil sampled rows")
	}
	next := r.time + r.config.TimeStep
	if r.config.UseTime {
		next = frame.Time
	}
	if math.IsNaN(next) || math.IsInf(next, 0) {
		return fmt.Errorf("composite: nonfinite sampled rows clock")
	}
	return r.sample(next)
}

func (r *SampledRows) sample(time float64) error {
	for copyIndex := 0; copyIndex < r.config.Copies; copyIndex++ {
		for row := 0; row < r.config.Rows; row++ {
			pose := r.config.Program.Sample(time, copyIndex, row)
			for _, value := range [...]float64{pose.SourceY, pose.SourceHeight, pose.X, pose.Y, pose.ScaleX, pose.ScaleY} {
				if math.IsNaN(value) || math.IsInf(value, 0) {
					return fmt.Errorf("composite: nonfinite sampled row pose")
				}
			}
			r.scratch[copyIndex*r.config.Rows+row] = pose
		}
	}
	r.poses, r.scratch = r.scratch, r.poses
	r.time = time
	return nil
}

func (r *SampledRows) Draw(dst *ebiten.Image) { r.DrawAt(dst, 0, 0) }

func (r *SampledRows) DrawAt(dst *ebiten.Image, x, y float64) {
	if r == nil || dst == nil {
		return
	}
	for _, pose := range r.poses {
		if pose.SourceHeight <= 0 || pose.ScaleX == 0 || pose.ScaleY == 0 {
			continue
		}
		var options ebiten.DrawImageOptions
		options.Filter, options.Blend = r.config.Filter, r.config.Blend
		options.GeoM.Scale(pose.ScaleX, pose.ScaleY)
		options.GeoM.Translate(x+pose.X, y+pose.Y)
		DrawRegion(dst, r.config.Image, Region{X: r.config.SourceX, Y: pose.SourceY,
			Width: r.config.SourceWidth, Height: pose.SourceHeight}, &options)
	}
}

func (r *SampledRows) Time() float64                  { return r.time }
func (r *SampledRows) Poses() []motion.SampledRowPose { return r.poses }

var _ kit.Effect = (*SampledRows)(nil)
