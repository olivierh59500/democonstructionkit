package composite

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

// WaterWave shifts reflected rows horizontally. Amplitude and Wavelength use
// destination pixels; Speed is radians per second and Phase is radians. A zero
// amplitude disables waves. Negative speed reverses the animation.
type WaterWave struct {
	Amplitude, Wavelength, Speed, Phase float64
}

// WaterReflectionConfig maps a source crop upside down below Horizon. Source
// uses absolute image coordinates; an empty rectangle selects the whole image.
// X is the destination left edge, and ScaleY controls reflected height. Tint's
// zero value is the identity color scale. Alpha and Fade range from zero to one;
// Fade is the fraction of opacity lost at the bottom, not an absolute opacity.
// RowHeight is the source-pixel height of a wave strip (zero means one). Larger
// values reduce geometry work at the cost of a less detailed wave curve.
type WaterReflectionConfig struct {
	Source             image.Rectangle
	X, Horizon, ScaleY float64
	Alpha              float32
	Tint               ebiten.ColorScale
	Fade               float64
	Wave               WaterWave
	RowHeight          int
	Filter             ebiten.Filter
	Blend              ebiten.Blend
}

// DefaultWaterReflectionConfig returns a half-opacity, full-height reflection.
// Set Horizon to the waterline and optionally crop Source before constructing it.
func DefaultWaterReflectionConfig() WaterReflectionConfig {
	return WaterReflectionConfig{ScaleY: 1, Alpha: .5, Wave: WaterWave{Wavelength: 32}}
}

const waterBatchRows = 256

// WaterReflection draws a live image, such as a scrolling surface, logo, or scene.
// It never advances or redraws that source and never reads pixels back from the
// GPU. Draw only the reflection after drawing the source normally. Input images
// remain caller-owned and must not share backing storage with the destination.
//
// Flat reflections use one DrawImage call. Waves/fading use bounded reusable
// triangles, including the final partial row. Initialization and source/crop
// changes may allocate; preparing successive frames of the same source does not.
type WaterReflection struct {
	config   WaterReflectionConfig
	source   *ebiten.Image
	crop     image.Rectangle
	view     *ebiten.Image
	vertices [waterBatchRows * 4]ebiten.Vertex
	indices  [waterBatchRows * 6]uint16
	closed   bool
}

func NewWaterReflection(config WaterReflectionConfig) (*WaterReflection, error) {
	w := &WaterReflection{}
	if err := w.SetConfig(config); err != nil {
		return nil, err
	}
	for row := range waterBatchRows {
		v := uint16(row * 4)
		i := row * 6
		copy(w.indices[i:i+6], []uint16{v, v + 1, v + 2, v, v + 2, v + 3})
	}
	return w, nil
}

// Config returns a value copy that can be changed and applied with SetConfig.
func (w *WaterReflection) Config() WaterReflectionConfig { return w.config }

// SetConfig changes placement and animation without resetting time. Invalid
// values leave the previous configuration unchanged. It allocates no images.
func (w *WaterReflection) SetConfig(c WaterReflectionConfig) error {
	for _, value := range [...]float64{c.X, c.Horizon, c.ScaleY, float64(c.Alpha), c.Fade,
		c.Wave.Amplitude, c.Wave.Wavelength, c.Wave.Speed, c.Wave.Phase,
		float64(c.Tint.R()), float64(c.Tint.G()), float64(c.Tint.B()), float64(c.Tint.A())} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("composite: nonfinite water reflection parameter")
		}
	}
	if c.ScaleY <= 0 || c.Alpha < 0 || c.Alpha > 1 || c.Fade < 0 || c.Fade > 1 ||
		c.RowHeight < 0 || (c.Wave.Amplitude != 0 && c.Wave.Wavelength <= 0) {
		return fmt.Errorf("composite: invalid water reflection scale, opacity, strip height or wavelength")
	}
	w.config = c
	return nil
}

// Draw samples frame.Time, so drawing repeatedly or skipping frames does not
// change wave speed. Tint/alpha use premultiplied blending for transparent logos.
// A crop extending outside the input is clipped before it is reflected.
func (w *WaterReflection) Draw(dst, source *ebiten.Image, frame kit.Frame) {
	if w.closed || dst == nil || source == nil || w.config.Alpha == 0 {
		return
	}
	c := w.config
	crop := c.Source
	if crop.Empty() {
		crop = source.Bounds()
	} else {
		crop = crop.Intersect(source.Bounds())
	}
	if crop.Empty() {
		return
	}
	if w.source != source || w.crop != crop {
		w.source, w.crop = source, crop
		w.view = source
		if crop != source.Bounds() {
			w.view = source.SubImage(crop).(*ebiten.Image)
		}
	}
	if c.Wave.Amplitude == 0 && c.Fade == 0 {
		op := ebiten.DrawImageOptions{Filter: c.Filter, Blend: c.Blend, ColorScale: c.Tint}
		op.GeoM.Scale(1, -c.ScaleY)
		op.GeoM.Translate(c.X, c.Horizon+float64(crop.Dy())*c.ScaleY)
		op.ColorScale.ScaleAlpha(c.Alpha)
		dst.DrawImage(w.view, &op)
		return
	}
	op := ebiten.DrawTrianglesOptions{Filter: c.Filter, Blend: c.Blend,
		ColorScaleMode: ebiten.ColorScaleModePremultipliedAlpha, Address: ebiten.AddressClampToZero}
	for top := 0; top < crop.Dy(); {
		rows, next := w.prepareRows(crop, top, frame.Time)
		dst.DrawTriangles(w.vertices[:rows*4], w.indices[:rows*6], w.view, &op)
		top = next
	}
}

// prepareRows fills at most one batch and never allocates. Reflection starts at
// the bottom edge of the crop, matching a negative DrawImage Y scale exactly.
func (w *WaterReflection) prepareRows(crop image.Rectangle, top int, seconds float64) (rows, next int) {
	c := w.config
	height := float64(crop.Dy())
	step := max(1, c.RowHeight)
	tint := c.Tint
	tint.ScaleAlpha(c.Alpha)
	for top < crop.Dy() && rows < waterBatchRows {
		bottom := min(top+min(step, crop.Dy()-top), crop.Dy())
		for edge, sourceY := range [2]int{top, bottom} {
			distance := float64(sourceY) * c.ScaleY
			x := c.X + c.Wave.offset(distance, seconds)
			y := c.Horizon + distance
			fade := float32(1 - c.Fade*float64(sourceY)/height)
			left, right := rows*4, rows*4+1
			if edge == 1 {
				left, right = rows*4+3, rows*4+2
			}
			vertex := ebiten.Vertex{DstX: float32(x), DstY: float32(y), SrcX: float32(crop.Min.X),
				SrcY: float32(crop.Max.Y - sourceY), ColorR: tint.R() * fade, ColorG: tint.G() * fade,
				ColorB: tint.B() * fade, ColorA: tint.A() * fade}
			w.vertices[left] = vertex
			vertex.DstX += float32(crop.Dx())
			vertex.SrcX = float32(crop.Max.X)
			w.vertices[right] = vertex
		}
		top = bottom
		rows++
	}
	return rows, top
}

func (wave WaterWave) offset(distance, seconds float64) float64 {
	if wave.Amplitude == 0 {
		return 0
	}
	return wave.Amplitude * math.Sin(2*math.Pi*distance/wave.Wavelength+seconds*wave.Speed+wave.Phase)
}

// Close releases borrowed image references. It does not dispose of input images.
func (w *WaterReflection) Close() error {
	w.source, w.view = nil, nil
	w.closed = true
	return nil
}
