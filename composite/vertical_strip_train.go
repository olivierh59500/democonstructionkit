package composite

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

// VerticalStripTrainConfig samples a source image at a moving row phase and
// draws fixed-height strips at independently spaced destination rows. Source
// stride controls the cached view grid; a sample between grid rows selects
// the preceding view. Out-of-range strips are omitted, not wrapped silently.
type VerticalStripTrainConfig struct {
	Image                     *ebiten.Image
	SourceX, SourceY          int
	SourceWidth, SourceStride int
	StripHeight, SampleStep   int
	Count                     int
	OutputX, OutputY, RowStep float64
	Phase, Velocity           float64
	Wrap                      *RasterWrap
	Filter                    ebiten.Filter
	Blend                     ebiten.Blend
}

// VerticalStripTrain owns a scalar phase and cached borrowed source views.
// Updating never draws; multiple layers may sample the same clock phase.
type VerticalStripTrain struct {
	config VerticalStripTrainConfig
	views  []*ebiten.Image
	phase  float64
}

func NewVerticalStripTrain(c VerticalStripTrainConfig) (*VerticalStripTrain, error) {
	if c.Image == nil || c.SourceStride < 1 || c.StripHeight < 1 || c.Count < 1 || c.Count > 16383 ||
		c.SourceX < 0 || c.SourceY < 0 || c.SourceWidth < 0 ||
		c.SourceY+c.StripHeight > c.Image.Bounds().Dy() ||
		c.SourceX+c.SourceWidth > c.Image.Bounds().Dx() {
		return nil, fmt.Errorf("composite: invalid vertical strip source or count")
	}
	if c.SourceWidth == 0 {
		c.SourceWidth = c.Image.Bounds().Dx() - c.SourceX
	}
	if c.SourceWidth < 1 {
		return nil, fmt.Errorf("composite: empty vertical strip width")
	}
	for _, value := range [...]float64{c.OutputX, c.OutputY, c.RowStep, c.Phase, c.Velocity} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("composite: nonfinite vertical strip setting")
		}
	}
	if c.Wrap != nil {
		if math.IsNaN(c.Wrap.Boundary) || math.IsInf(c.Wrap.Boundary, 0) ||
			math.IsNaN(c.Wrap.Restart) || math.IsInf(c.Wrap.Restart, 0) {
			return nil, fmt.Errorf("composite: invalid vertical strip wrap")
		}
		copy := *c.Wrap
		c.Wrap = &copy
	}
	maxSourceY := c.Image.Bounds().Dy() - c.StripHeight - c.SourceY
	viewCount := maxSourceY/c.SourceStride + 1
	if viewCount < 1 || viewCount > 1<<16 {
		return nil, fmt.Errorf("composite: too many vertical strip source views")
	}
	train := &VerticalStripTrain{config: c, views: make([]*ebiten.Image, viewCount), phase: c.Phase}
	for i := range train.views {
		y := c.Image.Bounds().Min.Y + c.SourceY + i*c.SourceStride
		x := c.Image.Bounds().Min.X + c.SourceX
		train.views[i] = c.Image.SubImage(image.Rect(x, y, x+c.SourceWidth, y+c.StripHeight)).(*ebiten.Image)
	}
	return train, nil
}

func (t *VerticalStripTrain) Step() {
	if t != nil {
		t.phase = rasterNext(t.phase, t.config.Velocity, t.config.Wrap)
	}
}

func (t *VerticalStripTrain) Update(kit.Frame) error {
	if t == nil {
		return fmt.Errorf("composite: nil vertical strip train")
	}
	t.Step()
	return nil
}

func (t *VerticalStripTrain) Draw(dst *ebiten.Image) { t.DrawAt(dst, 0, 0) }

// DrawAt adds a parent layer offset without changing the source phase.
func (t *VerticalStripTrain) DrawAt(dst *ebiten.Image, x, y float64) {
	if t == nil || dst == nil {
		return
	}
	c := t.config
	for i := 0; i < c.Count; i++ {
		sample := int(t.phase) + i*c.SampleStep
		index := sample / c.SourceStride
		if sample < 0 || index < 0 || index >= len(t.views) {
			continue
		}
		var op ebiten.DrawImageOptions
		op.Filter, op.Blend = c.Filter, c.Blend
		op.GeoM.Translate(c.OutputX+x, c.OutputY+y+float64(i)*c.RowStep)
		dst.DrawImage(t.views[index], &op)
	}
}

func (t *VerticalStripTrain) Phase() float64 { return t.phase }

func (t *VerticalStripTrain) SetPhase(value float64) error {
	if t == nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("composite: invalid vertical strip phase")
	}
	t.phase = value
	return nil
}

// SetVelocity changes source travel on the next update without rebuilding
// cached views or disturbing the current phase.
func (t *VerticalStripTrain) SetVelocity(value float64) error {
	if t == nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("composite: invalid vertical strip velocity")
	}
	t.config.Velocity = value
	return nil
}

func (t *VerticalStripTrain) Close() error {
	if t != nil {
		t.views = nil
	}
	return nil
}
