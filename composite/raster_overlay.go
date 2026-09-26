package composite

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

// RasterWrap resets a moving image after it crosses Boundary. Inclusive
// controls whether equality triggers the reset. Direction follows Velocity.
// One authored reset is applied per logical step, preserving scene timing.
type RasterWrap struct {
	Boundary, Restart float64
	Inclusive         bool
}

// RasterCopy places one additional draw relative to the overlay phase.
type RasterCopy struct{ X, Y float64 }

// RasterOverlayConfig places a borrowed raster image over an existing layer.
// Draw uses the destination's current alpha with BlendSourceAtop/SourceIn, or
// ordinary source-over blending for unmasked rasters. Step advances in pixels
// per logical tick; the optional source crop avoids another texture allocation.
type RasterOverlayConfig struct {
	Image                *ebiten.Image
	Source               image.Rectangle
	X, Y                 float64
	VelocityX, VelocityY float64
	ScaleX, ScaleY       float64
	AngleDegrees         float64
	AnchorX, AnchorY     float64
	Alpha                float64
	WrapX, WrapY         *RasterWrap
	Copies               []RasterCopy // Empty draws one copy at the phase position.
	Filter               ebiten.Filter
	Blend                ebiten.Blend
	ColorScale           ebiten.ColorScale // Zero value keeps the source colors unchanged.
}

// RasterOverlay owns its phase and source view but borrows its image and target.
// It draws without changing phase, so one logical Update can feed several draws.
type RasterOverlay struct {
	config RasterOverlayConfig
	image  *ebiten.Image
	x, y   float64
}

func NewRasterOverlay(c RasterOverlayConfig) (*RasterOverlay, error) {
	if c.Image == nil || c.ScaleX == 0 || c.ScaleY == 0 || c.Alpha < 0 || c.Alpha > 1 {
		return nil, fmt.Errorf("composite: invalid raster image, scale or alpha")
	}
	for _, v := range [...]float64{c.X, c.Y, c.VelocityX, c.VelocityY, c.ScaleX, c.ScaleY, c.AngleDegrees, c.AnchorX, c.AnchorY, c.Alpha} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("composite: nonfinite raster setting")
		}
	}
	for _, value := range [...]float32{c.ColorScale.R(), c.ColorScale.G(), c.ColorScale.B(), c.ColorScale.A()} {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return nil, fmt.Errorf("composite: nonfinite raster color scale")
		}
	}
	for _, wrap := range [...]*RasterWrap{c.WrapX, c.WrapY} {
		if wrap != nil && (math.IsNaN(wrap.Boundary) || math.IsInf(wrap.Boundary, 0) || math.IsNaN(wrap.Restart) || math.IsInf(wrap.Restart, 0)) {
			return nil, fmt.Errorf("composite: nonfinite raster wrap")
		}
	}
	if len(c.Copies) > 1<<16 {
		return nil, fmt.Errorf("composite: too many raster copies")
	}
	for _, copy := range c.Copies {
		if math.IsNaN(copy.X) || math.IsInf(copy.X, 0) || math.IsNaN(copy.Y) || math.IsInf(copy.Y, 0) {
			return nil, fmt.Errorf("composite: nonfinite raster copy")
		}
	}
	view := c.Image
	if !c.Source.Empty() {
		if !c.Source.In(c.Image.Bounds()) {
			return nil, fmt.Errorf("composite: raster crop exceeds image bounds")
		}
		view = c.Image.SubImage(c.Source).(*ebiten.Image)
	}
	if c.WrapX != nil {
		copy := *c.WrapX
		c.WrapX = &copy
	}
	if c.WrapY != nil {
		copy := *c.WrapY
		c.WrapY = &copy
	}
	c.Copies = append([]RasterCopy(nil), c.Copies...)
	return &RasterOverlay{config: c, image: view, x: c.X, y: c.Y}, nil
}

func (r *RasterOverlay) Step() {
	if r == nil {
		return
	}
	r.x = rasterNext(r.x, r.config.VelocityX, r.config.WrapX)
	r.y = rasterNext(r.y, r.config.VelocityY, r.config.WrapY)
}

func (r *RasterOverlay) Update(kit.Frame) error {
	if r == nil {
		return fmt.Errorf("composite: nil raster overlay")
	}
	r.Step()
	return nil
}

func rasterNext(position, velocity float64, wrap *RasterWrap) float64 {
	position += velocity
	if wrap == nil || velocity == 0 {
		return position
	}
	if velocity < 0 && (position < wrap.Boundary || wrap.Inclusive && position == wrap.Boundary) ||
		velocity > 0 && (position > wrap.Boundary || wrap.Inclusive && position == wrap.Boundary) {
		return wrap.Restart
	}
	return position
}

func (r *RasterOverlay) Draw(dst *ebiten.Image) { r.DrawAt(dst, 0, 0) }

// DrawAt applies an additional placement offset without changing the phase.
func (r *RasterOverlay) DrawAt(dst *ebiten.Image, x, y float64) {
	if r == nil || dst == nil || r.image == nil || r.config.Alpha == 0 {
		return
	}
	c := r.config
	count := len(c.Copies)
	if count == 0 {
		count = 1
	}
	for i := 0; i < count; i++ {
		copy := RasterCopy{}
		if len(c.Copies) > 0 {
			copy = c.Copies[i]
		}
		var op ebiten.DrawImageOptions
		op.Filter, op.Blend = c.Filter, c.Blend
		op.ColorScale = c.ColorScale
		op.GeoM.Translate(-c.AnchorX, -c.AnchorY)
		op.GeoM.Scale(c.ScaleX, c.ScaleY)
		op.GeoM.Rotate(c.AngleDegrees * math.Pi / 180)
		op.GeoM.Translate(r.x+x+copy.X, r.y+y+copy.Y)
		op.ColorScale.ScaleAlpha(float32(c.Alpha))
		dst.DrawImage(r.image, &op)
	}
}

func (r *RasterOverlay) Phase() (float64, float64) { return r.x, r.y }

func (r *RasterOverlay) SetPhase(x, y float64) error {
	if r == nil || math.IsNaN(x) || math.IsInf(x, 0) || math.IsNaN(y) || math.IsInf(y, 0) {
		return fmt.Errorf("composite: invalid raster phase")
	}
	r.x, r.y = x, y
	return nil
}

func (r *RasterOverlay) SetVelocity(x, y float64) error {
	if r == nil || math.IsNaN(x) || math.IsInf(x, 0) || math.IsNaN(y) || math.IsInf(y, 0) {
		return fmt.Errorf("composite: invalid raster velocity")
	}
	r.config.VelocityX, r.config.VelocityY = x, y
	return nil
}
