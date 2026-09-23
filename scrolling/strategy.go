package scrolling

import (
	"errors"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
)

type Alignment uint8

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// PageConfig lays out newline-delimited horizontal lines moving vertically.
// Width controls alignment; zero keeps each line at X. LineHeight is a minimum
// in pixels, expanded for larger mixed fonts. Commands trigger at line positions.
type PageConfig struct {
	Width, LineHeight float64
	Align             Alignment
}

// RecycledConfig selects exact recycled-slot transport through scrolling.New.
// Ring.Speed is pixels per Update, and its controls fire when slots are recycled.
// Vertical swaps movement axes without rotating the glyph images.
type RecycledConfig struct {
	Ring     RingConfig
	Vertical bool
}

// ProjectedConfig selects exact visible-slot form changes through scrolling.New.
// PixelsPerUpdate and Planes.PhaseStep preserve authored fixed-step recurrences.
// Face supports arbitrary metrics; an optional raster colors the projected text.
type ProjectedConfig struct {
	Planes          PlanesConfig
	Face            Face
	Raster          *ebiten.Image
	Draw            PlaneDraw
	PixelsPerUpdate float64
}

// SlicedConfig exposes the strip-history DNA transport through the same New
// constructor. Film is borrowed. Pixels are inserted in Stream.SliceWidth-wide
// slots; rotation uses film frames per second and continues when transport stops.
type SlicedConfig struct {
	Stream                       SliceStreamConfig
	Film                         *DNAFrames
	Draw                         DNADrawConfig
	SlicesPerUpdate              int
	RotationSpeed, RotationPhase float64
	Offsets                      []float64
	OnControl                    func(SliceControl) bool
	AdvanceAt                    func(kit.Frame) int
	RotationAt                   func(kit.Frame) float64
}

// FeedbackLayer feeds the text image into a persistent DNA face once per Update.
// PhaseSpeed uses rows per second; PhaseAt may retain an authored integer clock.
type FeedbackLayer struct {
	Config     FeedbackDNAConfig
	Gradient   *ebiten.Image
	X, Y       float64
	Phase      int
	PhaseSpeed float64
	PhaseAt    func(kit.Frame) int
}

// OutputConfig composes image deformation over any scrolling transport. Glyph
// modes run first, optional feedback faces next, then ordered image passes. The
// same ImagePass functions can process a logo or a complete scene with Pipeline.
type OutputConfig struct {
	Width, Height int
	Feedback      []FeedbackLayer
	Passes        []kit.ImagePass
}

func (s *Scrolling) finish() (*Scrolling, error) {
	if err := s.prepareModes(); err != nil {
		return nil, err
	}
	if s.config.Output == nil {
		return s, nil
	}
	c := s.config.Output
	if c.Width <= 0 || c.Height <= 0 {
		return nil, fmt.Errorf("scrolling: invalid output dimensions")
	}
	var source kit.Effect = kit.Func{OnDraw: s.drawCore}
	if len(c.Feedback) > 0 {
		f := &scrollFeedback{source: source, layers: append([]FeedbackLayer(nil), c.Feedback...)}
		for _, layer := range f.layers {
			if layer.Config.Width > c.Width || !finite(layer.X) || !finite(layer.Y) || !finite(layer.PhaseSpeed) {
				f.Close()
				return nil, fmt.Errorf("scrolling: invalid feedback layer")
			}
			face, err := NewFeedbackDNA(layer.Config)
			if err != nil {
				f.Close()
				return nil, err
			}
			f.faces = append(f.faces, face)
		}
		f.canvas = ebiten.NewImageWithOptions(image.Rect(0, 0, c.Width, c.Height), &ebiten.NewImageOptions{Unmanaged: true})
		source = f
	}
	if len(c.Passes) > 0 {
		p, err := kit.NewPipeline(source, c.Width, c.Height, c.Passes...)
		if err != nil {
			kit.Close(source)
			return nil, err
		}
		source = p
	}
	s.output = source
	return s, nil
}

func newTransport(c Config) (*Scrolling, error) {
	if !finite(c.X) || !finite(c.Y) || c.Speed != 0 || c.Gap != 0 || c.Advance != 0 || c.Repeat || c.Font != "" || c.Shape != "" {
		return nil, fmt.Errorf("scrolling: configure speed, repetition and font in the selected transport")
	}
	count := 0
	if c.Recycled != nil {
		count++
	}
	if c.Projected != nil {
		count++
	}
	if c.Sliced != nil {
		count++
	}
	if c.Crawl != nil {
		count++
	}
	if c.Bands != nil {
		count++
	}
	if c.Slots != nil {
		count++
	}
	if c.Feed != nil {
		count++
	}
	if c.Scanline != nil {
		count++
	}
	if c.Profiled != nil {
		count++
	}
	if count != 1 || c.Page != nil || c.Text != "" || c.Tokens != nil || c.Glyphs != nil || c.Controls != nil || len(c.Fonts) > 0 || len(c.Modes) > 0 || c.Map != nil || len(c.Shapes) > 0 || len(c.Effects) > 0 || c.Sequence != nil {
		return nil, fmt.Errorf("scrolling: choose one transport; configure glyph modes on the regular transport and image passes on any transport")
	}
	s := &Scrolling{config: c}
	if c.Recycled != nil {
		r, err := NewRing(c.Recycled.Ring)
		if err != nil {
			return nil, err
		}
		config := *c.Recycled
		config.Vertical = config.Vertical || c.Vertical
		s.backend = &recycledTransport{ring: r, config: config, x: c.X, y: c.Y}
	} else if c.Projected != nil {
		if c.X != 0 || c.Y != 0 || c.Vertical {
			return nil, fmt.Errorf("scrolling: projected placement belongs in Projected.Draw")
		}
		cfg := *c.Projected
		if !finite(cfg.PixelsPerUpdate) || cfg.PixelsPerUpdate < 0 {
			return nil, fmt.Errorf("scrolling: invalid projected transport speed")
		}
		p, err := NewPlanes(cfg.Planes)
		if err != nil {
			return nil, err
		}
		r, err := NewPlaneRenderer(cfg.Face, cfg.Raster)
		if err != nil {
			return nil, err
		}
		s.backend = &projectedTransport{planes: p, renderer: r, config: cfg}
	} else if c.Feed != nil {
		if c.X != 0 || c.Y != 0 || c.Vertical {
			return nil, fmt.Errorf("scrolling: feed placement belongs in the composition")
		}
		feed, err := NewFeed(*c.Feed)
		if err != nil {
			return nil, err
		}
		s.backend = feed
	} else if c.Scanline != nil {
		if c.X != 0 || c.Y != 0 || c.Vertical {
			return nil, fmt.Errorf("scrolling: scanline placement belongs in ScanlineConfig")
		}
		scanline, err := NewScanlineScroll(*c.Scanline)
		if err != nil {
			return nil, err
		}
		s.backend = scanline
	} else if c.Profiled != nil {
		if c.X != 0 || c.Y != 0 || c.Vertical {
			return nil, fmt.Errorf("scrolling: profiled placement belongs in ProfiledConfig")
		}
		profiled, err := NewProfiledScroll(*c.Profiled)
		if err != nil {
			return nil, err
		}
		s.backend = profiled
	} else if c.Bands != nil {
		if c.X != 0 || c.Y != 0 || c.Vertical {
			return nil, fmt.Errorf("scrolling: text-band placement belongs in BandsConfig")
		}
		bands, err := NewBitmapBands(*c.Bands)
		if err != nil {
			return nil, err
		}
		s.backend = bands
	} else if c.Slots != nil {
		if c.X != 0 || c.Y != 0 || c.Vertical {
			return nil, fmt.Errorf("scrolling: slot placement belongs in SlotsConfig")
		}
		slots, err := NewBitmapSlots(*c.Slots)
		if err != nil {
			return nil, err
		}
		s.backend = slots
	} else if c.Crawl != nil {
		if c.X != 0 || c.Y != 0 || c.Vertical {
			return nil, fmt.Errorf("scrolling: crawl placement belongs in CrawlConfig")
		}
		crawl, err := NewCrawl(*c.Crawl)
		if err != nil {
			return nil, err
		}
		s.backend = crawl
	} else {
		cfg := *c.Sliced
		if c.X != 0 || c.Y != 0 || c.Vertical || cfg.Film == nil || cfg.Film.Image == nil || cfg.SlicesPerUpdate < 0 || !finite(cfg.RotationSpeed) || !finite(cfg.RotationPhase) || cfg.Draw.SliceWidth != cfg.Stream.SliceWidth {
			return nil, fmt.Errorf("scrolling: invalid sliced transport or placement")
		}
		stream, err := NewSliceStream(cfg.Stream)
		if err != nil {
			return nil, err
		}
		cfg.Offsets = append([]float64(nil), cfg.Offsets...)
		if err := stream.SetFrames(cfg.RotationPhase, cfg.Offsets, cfg.Film.Count); err != nil {
			return nil, err
		}
		stream.Reset()
		s.backend = &slicedTransport{stream: stream, config: cfg}
	}
	result, err := s.finish()
	if err != nil {
		kit.Close(s.backend)
	}
	return result, err
}

type recycledTransport struct {
	ring   *Ring
	config RecycledConfig
	x, y   float64
}

func (r *recycledTransport) Update(kit.Frame) error { r.ring.Step(); return nil }
func (r *recycledTransport) Draw(dst *ebiten.Image) {
	if !r.config.Vertical {
		r.ring.DrawAt(dst, r.x, r.y)
		return
	}
	for _, i := range r.ring.order {
		p := r.ring.letters[i]
		region, ok := r.config.Ring.Font.Region(p.Rune)
		if !ok {
			continue
		}
		// A vertical column uses the same slot recurrence and atlas crop.
		op := ebiten.DrawImageOptions{Filter: r.config.Ring.Font.Filter}
		op.GeoM.Translate(r.x+p.Y, r.y+p.X)
		composite.DrawRegion(dst, r.config.Ring.Font.Image, region, &op)
	}
}

type projectedTransport struct {
	planes   *Planes
	renderer *PlaneRenderer
	config   ProjectedConfig
}

func (p *projectedTransport) Update(kit.Frame) error { return p.planes.Step(p.config.PixelsPerUpdate) }
func (p *projectedTransport) Draw(dst *ebiten.Image) {
	p.renderer.Draw(dst, p.planes.Points(), p.config.Draw)
}
func (p *projectedTransport) Close() error { return p.renderer.Close() }

type slicedTransport struct {
	stream *SliceStream
	config SlicedConfig
}

func (s *slicedTransport) Update(f kit.Frame) error {
	count := s.config.SlicesPerUpdate
	if s.config.AdvanceAt != nil {
		count = s.config.AdvanceAt(f)
	}
	if count < 0 {
		return fmt.Errorf("scrolling: negative strip advance")
	}
	s.stream.Step(count, s.config.OnControl)
	rotation := s.config.RotationPhase + f.Time*s.config.RotationSpeed
	if s.config.RotationAt != nil {
		rotation = s.config.RotationAt(f)
	}
	return s.stream.SetFrames(rotation, s.config.Offsets, s.config.Film.Count)
}
func (s *slicedTransport) Draw(dst *ebiten.Image) {
	s.config.Film.DrawSlices(dst, s.stream.Slices(), s.stream.Head(), s.config.Draw)
}

type scrollFeedback struct {
	source kit.Effect
	canvas *ebiten.Image
	layers []FeedbackLayer
	faces  []*FeedbackDNA
	frame  kit.Frame
}

func (f *scrollFeedback) Update(frame kit.Frame) error {
	f.frame = frame
	if err := f.source.Update(frame); err != nil {
		return err
	}
	f.canvas.Clear()
	f.source.Draw(f.canvas)
	for _, face := range f.faces {
		face.Step(f.canvas)
	}
	return nil
}
func (f *scrollFeedback) Draw(dst *ebiten.Image) {
	for i, c := range f.layers {
		phase := c.Phase + int(f.frame.Time*c.PhaseSpeed)
		if c.PhaseAt != nil {
			phase = c.PhaseAt(f.frame)
		}
		f.faces[i].DrawAt(dst, c.X, c.Y, phase, c.Gradient)
	}
}
func (f *scrollFeedback) Close() error {
	for _, face := range f.faces {
		face.Close()
	}
	f.faces = nil
	if f.canvas != nil {
		f.canvas.Deallocate()
		f.canvas = nil
	}
	return kit.Close(f.source)
}

// Close releases optional output history/passes and compatibility renderer
// resources. Font atlases, rasters and gradients remain caller-owned.
func (s *Scrolling) Close() error {
	err := errors.Join(kit.Close(s.output), kit.Close(s.backend))
	s.output, s.backend = nil, nil
	return err
}
