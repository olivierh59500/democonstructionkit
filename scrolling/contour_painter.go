package scrolling

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/render"
)

func (s *Scrolling) prepareContourModes() error {
	var modes map[string]Mode
	for name, mode := range s.config.Modes {
		if mode.Contours == nil {
			continue
		}
		if s.backend != nil || mode.Paint != nil || mode.Cells != nil {
			return fmt.Errorf("scrolling: mode %q has incompatible contour painting", name)
		}
		if modes == nil {
			modes = make(map[string]Mode, len(s.config.Modes))
			for key, value := range s.config.Modes {
				modes[key] = value
			}
			s.contourPainters = make(map[string]*ContourPainter)
		}
		config := *mode.Contours
		if len(config.Fonts) == 0 {
			config.Fonts = make(map[string]*font.ContourBank)
			for key, face := range s.config.Fonts {
				if face.Contours != nil {
					config.Fonts[key] = face.Contours
				}
			}
			if config.Font == "" {
				config.Font = s.config.Font
			}
		}
		painter, err := NewContourPainter(config)
		if err != nil {
			return fmt.Errorf("scrolling: contour mode %q: %w", name, err)
		}
		s.contourPainters[name] = painter
		mode.Paint = painter.Paint
		modes[name] = mode
	}
	if modes != nil {
		s.config.Modes = modes
	}
	return nil
}

// ContourPainterController returns the borrowed renderer owned by this mode.
func (s *Scrolling) ContourPainterController(mode string) *ContourPainter {
	return s.contourPainters[mode]
}

// ContourSample identifies one original font point before projection. Index is
// the visible glyph's scrolling index; Point remains in authored font units.
type ContourSample struct {
	Sample
	Point           geometry.Vec2
	Contour, Vertex int
}

// ContourPainterConfig works directly or through Mode.Contours. Map returns
// final destination coordinates; its GeoM argument includes the ordinary scroll
// transforms. Nil Map applies that GeoM to the original point. UV can select a
// raster/material coordinate independently of geometry. Texture is borrowed;
// nil allocates one owned white pixel. Fonts and callbacks stay borrowed.
type ContourPainterConfig struct {
	Fonts          map[string]*font.ContourBank
	Font           string
	Map            func(ContourSample, ebiten.GeoM) (geometry.Vec2, bool)
	UV             func(ContourSample, geometry.Vec2) geometry.Vec2
	Texture        *ebiten.Image
	Color          *color.NRGBA
	FillRule       ebiten.FillRule
	Blend          ebiten.Blend
	Filter         ebiten.Filter
	Address        ebiten.Address
	BatchTriangles int
}

// ContourPainter owns cached geometry buffers and optionally a white pixel.
// Scrolling batches a complete glyph run together, retaining parity between
// holes and overlapping glyphs. Draw never scans artwork or advances a clock.
type ContourPainter struct {
	config     ContourPainterConfig
	fonts      map[string]*font.ContourBank
	texture    *ebiten.Image
	ownTexture bool
	batch      *render.Batch
	points     []ebiten.Vertex
	base       ebiten.Vertex
	active     bool
	dst        *ebiten.Image
	used       int
	err        error
}

func NewContourPainter(c ContourPainterConfig) (*ContourPainter, error) {
	if len(c.Fonts) == 0 || len(c.Fonts) > 256 || c.FillRule < ebiten.FillRuleFillAll || c.FillRule > ebiten.FillRuleEvenOdd {
		return nil, fmt.Errorf("scrolling: invalid contour fonts or fill rule")
	}
	if c.BatchTriangles == 0 {
		c.BatchTriangles = 20000
	}
	if c.BatchTriangles < 2 || c.BatchTriangles > 20000 {
		return nil, fmt.Errorf("scrolling: invalid contour triangle budget")
	}
	p := &ContourPainter{config: c, fonts: make(map[string]*font.ContourBank, len(c.Fonts))}
	maxPoints := 0
	for name, bank := range c.Fonts {
		if bank == nil {
			return nil, fmt.Errorf("scrolling: missing contour font %q", name)
		}
		p.fonts[name] = bank
		if p.config.Font == "" && len(c.Fonts) == 1 {
			p.config.Font = name
		}
		for _, character := range bank.Characters() {
			glyph, _ := bank.Glyph(character)
			triangles := 0
			for _, contour := range glyph.Contours {
				maxPoints = max(maxPoints, len(contour))
				triangles += max(0, len(contour)-2)
			}
			if triangles > c.BatchTriangles {
				return nil, fmt.Errorf("scrolling: contour glyph exceeds triangle budget")
			}
		}
	}
	if p.config.Font == "" {
		p.config.Font = "default"
	}
	if p.fonts[p.config.Font] == nil {
		return nil, fmt.Errorf("scrolling: missing default contour font %q", p.config.Font)
	}
	paint := color.NRGBA{255, 255, 255, 255}
	if c.Color != nil {
		paint = *c.Color
	}
	p.base = render.Vertex(0, 0, 0, 0, paint)
	p.config.Fonts, p.config.Color = nil, nil
	p.points = make([]ebiten.Vertex, maxPoints)
	p.texture = c.Texture
	if p.texture == nil {
		p.texture = ebiten.NewImage(1, 1)
		p.texture.Fill(color.White)
		p.ownTexture = true
	}
	p.batch = render.NewBatch(c.BatchTriangles)
	p.batch.Options.FillRule = c.FillRule
	p.batch.Options.Blend = c.Blend
	p.batch.Options.Filter = c.Filter
	p.batch.Options.Address = c.Address
	return p, nil
}

func (p *ContourPainter) begin(dst *ebiten.Image) {
	p.err = nil
	p.used = 0
	p.active = true
	p.dst = dst
	p.batch.Begin(dst, p.texture)
}
func (p *ContourPainter) end() {
	if p.err != nil {
		p.batch.Discard()
	} else {
		p.batch.Flush()
	}
	p.active = false
	p.dst = nil
}

// Err reports an exceeded run budget. Split a fill-all run or enlarge its
// configured budget; parity runs must fit one batch to retain overlapping holes.
func (p *ContourPainter) Err() error {
	if p == nil {
		return nil
	}
	return p.err
}

func (p *ContourPainter) glyph(sample Sample) font.ContourGlyph {
	name := sample.Glyph.Font
	if name == "" {
		name = p.config.Font
	}
	if bank := p.fonts[name]; bank != nil {
		glyph, _ := bank.Glyph(sample.Glyph.Rune)
		return glyph
	}
	return font.ContourGlyph{}
}

// Paint implements Painter. A direct call draws one glyph; scrolling.New groups
// its contour mode into one complete run for correct even-odd overlap semantics.
func (p *ContourPainter) Paint(dst *ebiten.Image, sample Sample, op ebiten.DrawImageOptions) {
	if p == nil || p.texture == nil || dst == nil || !finite(sample.Time) {
		return
	}
	ownRun := !p.active
	if ownRun {
		p.begin(dst)
		defer p.end()
	}
	if p.dst != dst || p.err != nil {
		return
	}
	glyph := p.glyph(sample)
	for contourIndex, contour := range glyph.Contours {
		if len(contour) < 3 {
			continue
		}
		triangles := len(contour) - 2
		if p.used+triangles > p.config.BatchTriangles {
			p.err = fmt.Errorf("scrolling: contour run exceeds triangle budget")
			return
		}
		base := p.base
		base.ColorR *= op.ColorScale.R()
		base.ColorG *= op.ColorScale.G()
		base.ColorB *= op.ColorScale.B()
		base.ColorA *= op.ColorScale.A()
		valid := true
		for i, point := range contour {
			s := ContourSample{Sample: sample, Point: point, Contour: contourIndex, Vertex: i}
			mapped := point
			if p.config.Map != nil {
				mapped, valid = p.config.Map(s, op.GeoM)
			} else {
				mapped.X, mapped.Y = op.GeoM.Apply(point.X, point.Y)
			}
			if !valid || !contourCoordinate(mapped.X) || !contourCoordinate(mapped.Y) {
				valid = false
				break
			}
			uv := geometry.Vec2{}
			if p.config.UV != nil {
				uv = p.config.UV(s, mapped)
			}
			if !contourCoordinate(uv.X) || !contourCoordinate(uv.Y) {
				valid = false
				break
			}
			v := base
			v.DstX, v.DstY, v.SrcX, v.SrcY = float32(mapped.X), float32(mapped.Y), float32(uv.X), float32(uv.Y)
			p.points[i] = v
		}
		if !valid {
			continue
		}
		p.used += triangles
		p.batch.Fan(len(contour), func(i int) ebiten.Vertex { return p.points[i] })
	}
}

func contourCoordinate(value float64) bool {
	return finite(value) && math.Abs(value) <= math.MaxFloat32
}

// Close releases owned resources and preserves the caller's font banks/material.
func (p *ContourPainter) Close() error {
	if p != nil && p.texture != nil {
		if p.active {
			p.end()
		}
		if p.ownTexture {
			p.texture.Deallocate()
		}
		p.texture = nil
		p.points = nil
		p.fonts = nil
		p.batch = nil
	}
	return nil
}
