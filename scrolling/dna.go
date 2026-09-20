package scrolling

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// DNAFrameConfig describes the front/back filmstrip used by Phenomena. Height=0
// adapts to each glyph; no alphabet order or atlas layout is assumed. Front and
// Back are optional color masks, Core is the central raster between the faces.
// Supplied images remain caller-owned.
type DNAFrameConfig struct {
	Frames            int
	Height            int
	Step              float64 // Vertical pixels/frame; zero scales the original 2.25/33 ratio.
	Front, Back, Core *ebiten.Image
	CoreY             float64
}

type DNAEntry struct{ Y, Width, Height int }

type DNAFrames struct {
	Image   *ebiten.Image
	Entries []DNAEntry
	Count   int
	batch   *composite.QuadBatch
}

// NewDNAFrames builds both rotating faces from arbitrary glyph images, including
// proportional or odd-width cells. Transparent glyphs still retain their width.
func NewDNAFrames(glyphs []*ebiten.Image, c DNAFrameConfig) (*DNAFrames, error) {
	if c.Frames == 0 {
		c.Frames = 30
	}
	if c.Frames < 2 || c.Frames > 512 || c.Frames%2 != 0 || c.Height < 0 || !finite(c.Step) || c.Step < 0 || !finite(c.CoreY) || len(glyphs) == 0 {
		return nil, fmt.Errorf("scrolling: invalid DNA filmstrip configuration")
	}
	d := &DNAFrames{Count: c.Frames, Entries: make([]DNAEntry, len(glyphs)), batch: composite.NewQuadBatch(256)}
	d.batch.AlternateDiagonal = true
	width, height := 0, 0
	for i, g := range glyphs {
		if g == nil {
			return nil, fmt.Errorf("scrolling: nil DNA glyph %d", i)
		}
		h := c.Height
		if h == 0 {
			h = int(math.Ceil(float64(g.Bounds().Dy()) * 33 / 26))
		}
		if h < g.Bounds().Dy() {
			return nil, fmt.Errorf("scrolling: DNA height crops glyph %d", i)
		}
		d.Entries[i] = DNAEntry{Y: height, Width: g.Bounds().Dx(), Height: h}
		width = max(width, g.Bounds().Dx()*c.Frames)
		height += h
	}
	if width < 1 || height < 1 || width > 16384 || height > 16384 {
		return nil, fmt.Errorf("scrolling: DNA filmstrip exceeds texture dimensions")
	}
	d.Image = ebiten.NewImage(width, height)
	for i, g := range glyphs {
		d.buildGlyph(g, d.Entries[i], c)
	}
	return d, nil
}

func (d *DNAFrames) buildGlyph(g *ebiten.Image, e DNAEntry, c DNAFrameConfig) {
	w, h := e.Width, e.Height
	front, back := ebiten.NewImage(w, h), ebiten.NewImage(w, h)
	defer front.Deallocate()
	defer back.Deallocate()
	op := ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, float64(h-g.Bounds().Dy()))
	front.DrawImage(g, &op)
	op.GeoM.Reset()
	op.GeoM.Scale(-1, -1)
	op.GeoM.Translate(float64(w), float64(g.Bounds().Dy()))
	back.DrawImage(front, &op)
	a, b := ebiten.NewImage(w*d.Count, h), ebiten.NewImage(w*d.Count, h)
	defer a.Deallocate()
	defer b.Deallocate()
	step := c.Step
	if step == 0 {
		step = float64(h) * 2.25 / 33 * 30 / float64(d.Count)
	}
	frontY, backY := float64(h), 0.0
	for frame := 0; frame < d.Count; frame++ {
		op.GeoM.Reset()
		op.GeoM.Translate(float64(frame*w), frontY)
		a.DrawImage(front, &op)
		if frame == d.Count/2 {
			backY = -float64(h)
		}
		op.GeoM.Reset()
		op.GeoM.Translate(float64(frame*w), backY)
		b.DrawImage(back, &op)
		frontY -= step
		backY += step
	}
	tint := func(mask, gradient *ebiten.Image) *ebiten.Image {
		if gradient == nil {
			return mask
		}
		out := ebiten.NewImage(w*d.Count, h)
		opt := ebiten.DrawImageOptions{}
		opt.GeoM.Scale(float64(w*d.Count)/float64(gradient.Bounds().Dx()), float64(h)/float64(gradient.Bounds().Dy()))
		out.DrawImage(gradient, &opt)
		opt = ebiten.DrawImageOptions{Blend: ebiten.BlendDestinationIn}
		out.DrawImage(mask, &opt)
		return out
	}
	aColored, bColored := tint(a, c.Front), tint(b, c.Back)
	if aColored != a {
		defer aColored.Deallocate()
	}
	if bColored != b {
		defer bColored.Deallocate()
	}
	op = ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, float64(e.Y))
	d.Image.DrawImage(bColored, &op)
	if c.Core != nil {
		op.GeoM.Reset()
		op.GeoM.Scale(float64(w*d.Count)/float64(c.Core.Bounds().Dx()), 1)
		op.GeoM.Translate(0, float64(e.Y)+c.CoreY)
		d.Image.DrawImage(c.Core, &op)
	}
	op.GeoM.Reset()
	op.GeoM.Translate(0, float64(e.Y))
	d.Image.DrawImage(aColored, &op)
}

func (d *DNAFrames) Close() error {
	if d.Image != nil {
		d.Image.Deallocate()
		d.Image = nil
	}
	return nil
}

func (d *DNAFrames) FrameAt(rotation, offset float64) int {
	return int(motion.Wrap(rotation+offset, float64(d.Count)))
}

// DNASlice represents a single source column in a circular scroll buffer.
type DNASlice struct{ Glyph, Frame, Slice int }

type DNADrawConfig struct {
	SliceWidth       int
	ScaleX, ScaleY   float32
	OriginX, OriginY float32
	Y                func(index int) float64 // Baseline in unscaled pixels, sampled in draw order.
}

// DrawSlices renders an existing circular buffer, preserving float32 quad
// arithmetic and the original diagonal. The final narrow slice is never lost.
func (d *DNAFrames) DrawSlices(dst *ebiten.Image, slices []DNASlice, head int, c DNADrawConfig) {
	if d.Image == nil || len(slices) == 0 || c.SliceWidth < 1 {
		return
	}
	d.batch.Begin(dst, d.Image)
	head = ((head % len(slices)) + len(slices)) % len(slices)
	for i := range slices {
		s := slices[(head+i)%len(slices)]
		if s.Glyph < 0 || s.Glyph >= len(d.Entries) || s.Frame < 0 || s.Frame >= d.Count || s.Slice < 0 {
			continue
		}
		e := d.Entries[s.Glyph]
		column := s.Slice * c.SliceWidth
		if column >= e.Width {
			continue
		}
		w := min(c.SliceWidth, e.Width-column)
		sy := e.Y
		sx := s.Frame*e.Width + column
		y := 0.0
		if c.Y != nil {
			y = c.Y(i)
		}
		d.batch.Rect(image.Rect(sx, sy, sx+w, sy+e.Height), c.OriginX+float32(i*c.SliceWidth)*c.ScaleX, c.OriginY+float32(y)*c.ScaleY, float32(w)*c.ScaleX, float32(e.Height)*c.ScaleY)
	}
	d.batch.Flush()
}

type DNAConfig struct {
	Film                 DNAFrameConfig
	SliceWidth           int
	RotationSpeed, Phase float64     // Filmstrip frames per second / frame offset.
	Twist                motion.Wave // Phase offset in frames, sampled in text pixels.
	Baseline             motion.Wave
}

// DNA is a reusable scrolling mode. It caches a filmstrip per source glyph, so
// a single scroll can mix faces of different sizes, mappings and advances.
type DNA struct {
	config DNAConfig
	cache  map[*ebiten.Image]*DNAFrames
}

func NewDNA(c DNAConfig) (*DNA, error) {
	if c.SliceWidth == 0 {
		c.SliceWidth = 2
	}
	if c.SliceWidth < 1 || !finite(c.RotationSpeed) || !finite(c.Phase) {
		return nil, fmt.Errorf("scrolling: invalid DNA parameters")
	}
	if c.Film.Frames == 0 {
		c.Film.Frames = 30
	}
	if c.Film.Frames < 2 || c.Film.Frames > 512 || c.Film.Frames%2 != 0 || c.Film.Height < 0 || !finite(c.Film.Step) || c.Film.Step < 0 || !finite(c.Film.CoreY) {
		return nil, fmt.Errorf("scrolling: invalid DNA filmstrip configuration")
	}
	for _, w := range []motion.Wave{c.Twist, c.Baseline} {
		for _, v := range []float64{w.Amplitude, w.Spatial, w.Speed, w.Phase} {
			if !finite(v) {
				return nil, fmt.Errorf("scrolling: invalid DNA wave")
			}
		}
	}
	return &DNA{config: c, cache: map[*ebiten.Image]*DNAFrames{}}, nil
}

// Prepare validates and caches glyphs before rendering. It may be called with
// every face in a scene. Calling it repeatedly does not allocate new filmstrips.
func (d *DNA) Prepare(glyphs ...*ebiten.Image) error {
	for _, g := range glyphs {
		if g == nil || d.cache[g] != nil {
			continue
		}
		f, err := NewDNAFrames([]*ebiten.Image{g}, d.config.Film)
		if err != nil {
			return err
		}
		d.cache[g] = f
	}
	return nil
}

func (d *DNA) Mode() Mode {
	return Mode{Paint: d.paint, Prepare: func(glyphs []Glyph) error {
		for _, g := range glyphs {
			if err := d.Prepare(g.Image); err != nil {
				return err
			}
		}
		return nil
	}}
}

func (d *DNA) paint(dst *ebiten.Image, s Sample, op ebiten.DrawImageOptions) {
	if s.Glyph.Image == nil {
		return
	}
	f := d.cache[s.Glyph.Image]
	if f == nil {
		if err := d.Prepare(s.Glyph.Image); err != nil {
			panic(err)
		}
		f = d.cache[s.Glyph.Image]
	}
	e := f.Entries[0]
	for x := 0; x < e.Width; x += d.config.SliceWidth {
		position := s.Glyph.Offset + float64(x)*s.Glyph.ScaleX
		frame := f.FrameAt(s.Time*d.config.RotationSpeed+d.config.Phase, d.config.Twist.At(position, s.Time))
		r := image.Rect(frame*e.Width+x, 0, frame*e.Width+min(x+d.config.SliceWidth, e.Width), e.Height)
		img := f.Image.SubImage(r).(*ebiten.Image)
		options := op
		options.GeoM.Reset()
		options.GeoM.Translate(float64(x), d.config.Baseline.At(position, s.Time))
		options.GeoM.Concat(op.GeoM)
		dst.DrawImage(img, &options)
	}
}

func (d *DNA) Close() error {
	for _, f := range d.cache {
		f.Close()
	}
	clear(d.cache)
	return nil
}
