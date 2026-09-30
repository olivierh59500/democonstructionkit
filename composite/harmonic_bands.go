package composite

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/render"
)

// HarmonicBandsConfig describes colored strips between two independently
// moving edges. Motion.X is the left edge's Y and Motion.Y the right edge's Y.
// ClockScale multiplies Frame.Time; its all-zero default is {1, 0}. Each color
// creates one strip. Bounds apply after optional rounding, preserving authored
// integer edge clamps. White is an optional borrowed one-pixel white image.
type HarmonicBandsConfig struct {
	LeftX, RightX, Thickness float64
	Colors                   []color.NRGBA
	Motion                   motion.HarmonicFormationConfig
	ClockScale               [2]float64
	PixelSnap                bool
	Bounds                   *motion.FormationBounds
	White                    *ebiten.Image
	Blend                    ebiten.Blend
}

// HarmonicBands samples each edge once per update, then batches filled or
// outlined strips without intermediate images. Poses and draws never advance
// the clocks. The same cached strip bank can be drawn on multiple layers.
type HarmonicBands struct {
	config    HarmonicBandsConfig
	formation *motion.HarmonicFormation
	poses     []motion.Point
	paints    []color.Color
	white     *ebiten.Image
	owned     bool
	batch     *render.Batch
}

func NewHarmonicBands(c HarmonicBandsConfig) (*HarmonicBands, error) {
	if len(c.Colors) < 1 || len(c.Colors) > 4096 || c.RightX <= c.LeftX || c.Thickness <= 0 {
		return nil, fmt.Errorf("composite: invalid harmonic strip count or dimensions")
	}
	for _, value := range []float64{c.LeftX, c.RightX, c.Thickness, c.ClockScale[0], c.ClockScale[1]} {
		if !bandFinite(value) {
			return nil, fmt.Errorf("composite: nonfinite harmonic strip setting")
		}
	}
	if c.White != nil && (c.White.Bounds().Min.X != 0 || c.White.Bounds().Min.Y != 0 || c.White.Bounds().Dx() != 1 || c.White.Bounds().Dy() != 1) {
		return nil, fmt.Errorf("composite: harmonic strips need a one-pixel white material")
	}
	if c.Bounds != nil {
		b := *c.Bounds
		if !bandFinite(b.Min.X) || !bandFinite(b.Min.Y) || !bandFinite(b.Max.X) || !bandFinite(b.Max.Y) || b.Min.X > b.Max.X || b.Min.Y > b.Max.Y {
			return nil, fmt.Errorf("composite: invalid harmonic strip bounds")
		}
		c.Bounds = &b
	}
	formation, err := motion.NewHarmonicFormation(c.Motion)
	if err != nil {
		return nil, err
	}
	if formation.IndexOffsetCount() != 0 && formation.IndexOffsetCount() < len(c.Colors) {
		return nil, fmt.Errorf("composite: harmonic strip phases are shorter than color bank")
	}
	if c.ClockScale == [2]float64{} {
		c.ClockScale[0] = 1
	}
	c.Colors = append([]color.NRGBA(nil), c.Colors...)
	c.Motion = motion.HarmonicFormationConfig{} // The compiled formation owns its data.
	b := &HarmonicBands{config: c, formation: formation, poses: make([]motion.Point, len(c.Colors)), paints: make([]color.Color, len(c.Colors)),
		white: c.White, batch: render.NewBatch(min(20000, len(c.Colors)*8))}
	for i, paint := range c.Colors {
		b.paints[i] = paint // Box each immutable color once, outside draw loops.
	}
	b.batch.Options.Blend = c.Blend
	if b.white == nil {
		b.white = ebiten.NewImage(1, 1)
		b.white.Fill(color.White)
		b.owned = true
	}
	if err := b.Sample([2]float64{}, 1); err != nil {
		b.Close()
		return nil, err
	}
	return b, nil
}

func (b *HarmonicBands) Update(f kit.Frame) error {
	if b == nil || b.white == nil || !bandFinite(f.Time) {
		return fmt.Errorf("composite: invalid harmonic strip update")
	}
	return b.Sample([2]float64{f.Time * b.config.ClockScale[0], f.Time * b.config.ClockScale[1]}, 1)
}

// Sample accepts explicit clock/envelope values, for music or cue controllers.
// It prepares the same retained poses used by both rendering materials.
func (b *HarmonicBands) Sample(clocks [2]float64, envelope float64) error {
	if b == nil || b.white == nil || !bandFinite(clocks[0]) || !bandFinite(clocks[1]) || !bandFinite(envelope) {
		return fmt.Errorf("composite: invalid harmonic strip clocks")
	}
	for i := range b.poses {
		p := b.formation.At(i, clocks, envelope)
		if !bandFinite(p.X) || !bandFinite(p.Y) {
			return fmt.Errorf("composite: nonfinite harmonic strip pose")
		}
		if b.config.PixelSnap {
			p.X, p.Y = math.Round(p.X), math.Round(p.Y)
		}
		if bounds := b.config.Bounds; bounds != nil {
			p.X, p.Y = max(bounds.Min.X, min(bounds.Max.X, p.X)), max(bounds.Min.Y, min(bounds.Max.Y, p.Y))
		}
		b.poses[i] = p
	}
	return nil
}

// Poses returns borrowed left/right edge heights, valid until the next update.
func (b *HarmonicBands) Poses() []motion.Point { return b.poses }

func (b *HarmonicBands) Draw(dst *ebiten.Image) { b.DrawAt(dst, 0, 0) }
func (b *HarmonicBands) DrawAt(dst *ebiten.Image, x, y float64) {
	if b == nil || b.white == nil || dst == nil {
		return
	}
	b.batch.Begin(dst, b.white)
	for i, p := range b.poses {
		paint := b.paints[i]
		left, right := x+b.config.LeftX, x+b.config.RightX
		b.batch.Quad([4]ebiten.Vertex{
			render.Vertex(left, y+p.X, 0, 0, paint), render.Vertex(right, y+p.Y, 1, 0, paint),
			render.Vertex(right, y+p.Y+b.config.Thickness, 1, 1, paint), render.Vertex(left, y+p.X+b.config.Thickness, 0, 1, paint),
		})
	}
	b.batch.Flush()
}

// DrawOutline draws the cached strips with their own colors and uniform width.
func (b *HarmonicBands) DrawOutline(dst *ebiten.Image, width float64) {
	if b == nil || b.white == nil || dst == nil {
		return
	}
	b.batch.Begin(dst, b.white)
	for i, p := range b.poses {
		quad := [4]geometry.Vec2{{X: b.config.LeftX, Y: p.X}, {X: b.config.RightX, Y: p.Y},
			{X: b.config.RightX, Y: p.Y + b.config.Thickness}, {X: b.config.LeftX, Y: p.X + b.config.Thickness}}
		b.batch.StrokePath(quad[:], render.PathStroke{Width: width}, b.paints[i])
	}
	b.batch.Flush()
}

// Close releases only an internally created white material; borrowed artwork
// remains caller-owned. Closing twice is safe.
func (b *HarmonicBands) Close() error {
	if b != nil {
		if b.owned && b.white != nil {
			b.white.Deallocate()
		}
		b.white = nil
	}
	return nil
}

var _ kit.Effect = (*HarmonicBands)(nil)
