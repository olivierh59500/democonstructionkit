package composite

import (
	"errors"
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

// BackgroundConfig describes a cropped image repeated on either axis. A zero
// period disables repetition on that axis. Periods are in unscaled source pixels
// and need not match the crop: a larger period leaves gaps, a smaller one overlaps
// copies in increasing row/column order. CopiesX and CopiesY optionally limit
// repetition to indices [0, count), matching finite backdrop strips. Empty
// Source uses the complete image.
// Zero scale defaults to one. Negative scales are rejected; mirror the source in
// a preceding image pass when required. Source and destination remain borrowed.
type BackgroundConfig struct {
	Source                                 image.Rectangle
	PeriodX, PeriodY                       float64
	SingleCopyOnEntryX, SingleCopyOnEntryY bool // Suppress repeats while the origin enters from that viewport edge.
	CopiesX, CopiesY                       int  // Zero repeats without an index limit.
	ScaleX, ScaleY                         float64
	ParallaxX, ParallaxY                   float64
	Filter                                 ebiten.Filter
	Blend                                  ebiten.Blend
	ColorScale                             ebiten.ColorScale
	MaxCopies                              int // Fallback draw budget; zero defaults to 16384. Repeated quads cost one.
}

// DefaultBackgroundConfig follows the camera at full speed with unit scaling.
// A zero parallax factor intentionally pins that axis to the viewport.
func DefaultBackgroundConfig() BackgroundConfig {
	return BackgroundConfig{ScaleX: 1, ScaleY: 1, ParallaxX: 1, ParallaxY: 1}
}

// BackgroundPose separates the screen origin from camera motion in source pixels.
// X and Y are the image's destination origin before camera motion. Repeated axes
// are rebased near the viewport, keeping long-running animation coordinates small.
type BackgroundPose struct {
	X, Y, CameraX, CameraY float64
}

// Background reuses its source crop and submits only visible image copies. The
// same object can render logos, textures or live surfaces. It does not advance
// animation, so drawing several masks/layers from one pose remains synchronized.
type Background struct {
	config       BackgroundConfig
	source, crop *ebiten.Image
	vertices     [4]ebiten.Vertex
	err          error
}

var ErrBackgroundBudget = errors.New("composite: repeated background exceeds its copy budget")

// Err reports the last Draw's resource-budget error. Invalid excessive-density
// frames are skipped in full rather than freezing or rendering a partial layer.
func (b *Background) Err() error { return b.err }

func NewBackground(c BackgroundConfig) (*Background, error) {
	if c.MaxCopies == 0 {
		c.MaxCopies = 16384
	}
	if c.MaxCopies < 1 || c.MaxCopies > 1<<20 {
		return nil, fmt.Errorf("composite: invalid background copy budget")
	}
	if c.ScaleX == 0 {
		c.ScaleX = 1
	}
	if c.ScaleY == 0 {
		c.ScaleY = 1
	}
	for _, value := range []float64{c.PeriodX, c.PeriodY, c.ScaleX, c.ScaleY, c.ParallaxX, c.ParallaxY} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("composite: background parameters must be finite")
		}
	}
	if c.PeriodX < 0 || c.PeriodY < 0 || c.ScaleX <= 0 || c.ScaleY <= 0 ||
		c.CopiesX < 0 || c.CopiesY < 0 || c.CopiesX > 1<<20 || c.CopiesY > 1<<20 {
		return nil, fmt.Errorf("composite: invalid background periods, copy counts or scales")
	}
	return &Background{config: c}, nil
}

// Draw places the image using an independently controlled camera and parallax.
// Use dst.SubImage to select a viewport without allocating another render target.
func (b *Background) Draw(dst, source *ebiten.Image, pose BackgroundPose) {
	if dst == nil || source == nil || b == nil {
		return
	}
	b.err = nil
	if b.source != source {
		b.source, b.crop = source, source
		if !b.config.Source.Empty() {
			region := b.config.Source.Intersect(source.Bounds())
			if region.Empty() {
				b.crop = nil
			} else {
				b.crop = source.SubImage(region).(*ebiten.Image)
			}
		}
	}
	if b.crop == nil {
		return
	}
	c := b.config
	x := pose.X - float64(pose.CameraX*c.ParallaxX*c.ScaleX)
	y := pose.Y - float64(pose.CameraY*c.ParallaxY*c.ScaleY)
	if math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) {
		return
	}
	view, tile := dst.Bounds(), b.crop.Bounds()
	w, h := float64(tile.Dx())*c.ScaleX, float64(tile.Dy())*c.ScaleY
	px, py := c.PeriodX*c.ScaleX, c.PeriodY*c.ScaleY
	px = backgroundEntryPeriod(x, float64(view.Min.X), px, c.SingleCopyOnEntryX)
	py = backgroundEntryPeriod(y, float64(view.Min.Y), py, c.SingleCopyOnEntryY)
	if px == c.PeriodX*c.ScaleX && py == c.PeriodY*c.ScaleY && b.repeatedQuad(dst, x, y, w, h) {
		return
	}
	x, firstX, lastX, okX := backgroundCopyRangeLimit(x, w, px, float64(view.Min.X), float64(view.Max.X), c.MaxCopies, c.CopiesX)
	y, firstY, lastY, okY := backgroundCopyRangeLimit(y, h, py, float64(view.Min.Y), float64(view.Max.Y), c.MaxCopies, c.CopiesY)
	if !okX || !okY {
		b.err = ErrBackgroundBudget
		return
	}
	if firstX > lastX || firstY > lastY {
		return
	}
	if (lastX - firstX + 1) > c.MaxCopies/(lastY-firstY+1) {
		b.err = ErrBackgroundBudget
		return
	}
	op := ebiten.DrawImageOptions{Filter: c.Filter, Blend: c.Blend, ColorScale: c.ColorScale}
	for iy := firstY; iy <= lastY; iy++ {
		for ix := firstX; ix <= lastX; ix++ {
			op.GeoM.Reset()
			op.GeoM.Scale(c.ScaleX, c.ScaleY)
			op.GeoM.Translate(x+float64(float64(ix)*c.PeriodX*c.ScaleX), y+float64(float64(iy)*c.PeriodY*c.ScaleY))
			dst.DrawImage(b.crop, &op)
		}
	}
}

// repeatedQuad avoids per-tile submissions for nearest-filtered, tightly tiled
// images. Linear filtering deliberately keeps independent cropped image edges.
func (b *Background) repeatedQuad(dst *ebiten.Image, x, y, width, height float64) bool {
	c, source := b.config, b.crop.Bounds()
	if c.Filter != ebiten.FilterNearest || c.CopiesX != 0 || c.CopiesY != 0 || c.PeriodX == 0 && c.PeriodY == 0 ||
		c.PeriodX != 0 && c.PeriodX != float64(source.Dx()) ||
		c.PeriodY != 0 && c.PeriodY != float64(source.Dy()) {
		return false
	}
	// Match pixel-aligned sprite placement exactly; fractional transforms retain
	// DrawImage's triangle interpolation and clipping rules.
	if x != math.Trunc(x) || y != math.Trunc(y) || c.ScaleX != math.Trunc(c.ScaleX) || c.ScaleY != math.Trunc(c.ScaleY) {
		return false
	}
	view := dst.Bounds()
	left, top, right, bottom := float64(view.Min.X), float64(view.Min.Y), float64(view.Max.X), float64(view.Max.Y)
	if c.PeriodX == 0 {
		left, right = math.Max(left, x), math.Min(right, x+width)
	}
	if c.PeriodY == 0 {
		top, bottom = math.Max(top, y), math.Min(bottom, y+height)
	}
	if right <= left || bottom <= top {
		return true
	}
	// Texture coordinates stay bounded even after long camera travel.
	if c.PeriodX != 0 {
		x = math.Mod(x, width)
	}
	if c.PeriodY != 0 {
		y = math.Mod(y, height)
	}
	for i, p := range [4][2]float64{{left, top}, {right, top}, {left, bottom}, {right, bottom}} {
		b.vertices[i] = ebiten.Vertex{DstX: float32(p[0]), DstY: float32(p[1]), SrcX: float32(float64(source.Min.X) + (p[0]-x)/c.ScaleX), SrcY: float32(float64(source.Min.Y) + (p[1]-y)/c.ScaleY), ColorR: c.ColorScale.R(), ColorG: c.ColorScale.G(), ColorB: c.ColorScale.B(), ColorA: c.ColorScale.A()}
	}
	indices := [6]uint16{0, 1, 2, 1, 2, 3}
	dst.DrawTriangles(b.vertices[:], indices[:], b.crop, &ebiten.DrawTrianglesOptions{Address: ebiten.AddressRepeat, Filter: c.Filter, Blend: c.Blend, ColorScaleMode: ebiten.ColorScaleModePremultipliedAlpha})
	return true
}

// DrawAt is the camera-free convenience form for an existing animation's offsets.
func (b *Background) DrawAt(dst, source *ebiten.Image, x, y float64) {
	b.Draw(dst, source, BackgroundPose{X: x, Y: y})
}

// backgroundCopies keeps placement math independent from graphics initialization.
// A nonrepeated axis draws at most one copy. Half-open edges avoid double drawing
// adjacent opaque tiles while preserving all overlapping translucent copies.
func backgroundCopies(origin, extent, period, minimum, maximum float64) (float64, int, int) {
	origin, first, last, _ := backgroundCopyRange(origin, extent, period, minimum, maximum, 16384)
	return origin, first, last
}

func backgroundCopyRange(origin, extent, period, minimum, maximum float64, budget int) (float64, int, int, bool) {
	if period <= 0 {
		if origin+extent <= minimum || origin >= maximum {
			return origin, 1, 0, true
		}
		return origin, 0, 0, true
	}
	// Keep ordinary placement arithmetic identical to direct DrawImage calls.
	// Premature normalization can move nearest-filtered fractional edges across
	// a sampling boundary through floating-point reassociation.
	if math.Abs(origin-minimum) > period*1e6 {
		origin = minimum + math.Mod(origin-minimum, period)
	}
	first := math.Floor((minimum-origin-extent)/period) + 1
	last := math.Ceil((maximum-origin)/period) - 1
	// Bound work and float-to-int conversion before entering either draw loop.
	if math.IsNaN(first) || math.IsNaN(last) || math.IsInf(first, 0) || math.IsInf(last, 0) || math.Abs(first) >= 1<<30 || math.Abs(last) >= 1<<30 || last-first+1 > float64(budget) {
		return origin, 1, 0, false
	}
	return origin, int(first), int(last), true
}

// backgroundCopyRangeLimit keeps finite tile indices anchored to their authored
// origin. Rebasing would change which copies exist when the camera travels far.
func backgroundCopyRangeLimit(origin, extent, period, minimum, maximum float64, budget, copies int) (float64, int, int, bool) {
	if copies == 0 || period <= 0 {
		return backgroundCopyRange(origin, extent, period, minimum, maximum, budget)
	}
	if math.IsNaN(extent) || math.IsInf(extent, 0) || math.IsNaN(period) || math.IsInf(period, 0) {
		return origin, 1, 0, false
	}
	first := math.Floor((minimum-origin-extent)/period) + 1
	last := math.Ceil((maximum-origin)/period) - 1
	if math.IsNaN(first) || math.IsNaN(last) {
		return origin, 1, 0, false
	}
	if first > float64(copies-1) || last < 0 {
		return origin, 1, 0, true
	}
	first = math.Max(first, 0)
	last = math.Min(last, float64(copies-1))
	if first > last {
		return origin, 1, 0, true
	}
	if last-first+1 > float64(budget) {
		return origin, 1, 0, false
	}
	return origin, int(first), int(last), true
}

// BackgroundLayer adapts a background to kit.Layers or kit.Group. Layers borrow
// their image and renderer; no texture is allocated or destroyed by this wrapper.
type BackgroundLayer struct {
	Renderer *Background
	Image    *ebiten.Image
	Pose     BackgroundPose
	Sample   func(kit.Frame) BackgroundPose
	// Velocity is measured in destination pixels per second and is added after
	// Sample, if one is supplied. It makes a tiled backdrop scroll with only an
	// image, a renderer and a velocity; custom motion can still use Sample.
	VelocityX, VelocityY float64
	frame                kit.Frame
}

func (b *BackgroundLayer) Update(frame kit.Frame) error {
	b.frame = frame
	if b.Renderer == nil {
		return fmt.Errorf("composite: nil background renderer")
	}
	if math.IsNaN(b.VelocityX) || math.IsInf(b.VelocityX, 0) || math.IsNaN(b.VelocityY) || math.IsInf(b.VelocityY, 0) ||
		math.IsNaN(frame.Time) || math.IsInf(frame.Time, 0) {
		return fmt.Errorf("composite: nonfinite background motion")
	}
	return b.Renderer.Err()
}
func (b *BackgroundLayer) Draw(dst *ebiten.Image) {
	b.Renderer.Draw(dst, b.Image, b.poseAt())
}

func (b *BackgroundLayer) poseAt() BackgroundPose {
	pose := b.Pose
	if b.Sample != nil {
		pose = b.Sample(b.frame)
	}
	pose.X += b.VelocityX * b.frame.Time
	pose.Y += b.VelocityY * b.frame.Time
	return pose
}
