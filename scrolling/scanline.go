package scrolling

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// ScanlineWrap selects how a horizontally sampled strip crosses the text image.
type ScanlineWrap uint8

const (
	// ScanlineSplit draws the end and beginning as separate quads.
	ScanlineSplit ScanlineWrap = iota
	// ScanlineAddressRepeat asks the graphics sampler to repeat source pixels.
	ScanlineAddressRepeat
	// ScanlineReject draws only full-width source windows strictly inside the image.
	ScanlineReject
)

// ScanlineConfig composes proportional bitmap text, an authored cumulative
// displacement table, a cyclic source-row bounce and a visible strip renderer.
// The font, text, dimensions, clock, wave and wrapping are independently editable.
// WaveStep and BounceRate use simulation ticks, not display refresh callbacks.
type ScanlineConfig struct {
	Font              *Atlas
	Text              string
	Wave              []int
	Program           *composite.DisplacementProgram // Optional intro plus repeated wave.
	ViewportWidth     int
	ViewportHeight    int
	SurfaceWidth      int
	SourceRows        int
	RowHeight         int
	StripHeight       int
	DestinationY      int
	Scale             float64
	WaveStep          int
	CursorRows        int
	BounceAmplitude   float64
	BounceRate        float64
	MissingAdvance    float64
	MissingRune       rune // Draw this literal glyph for unsupported text runes.
	ClampCursor       bool
	Wrap              ScanlineWrap
	AlternateDiagonal bool
	Background        color.Color
	UseTime           bool // Sample Frame.Time instead of Frame.Tick for variable-speed playback.
	SurfaceUnmanaged  bool // Keep source texture outside the graphics atlas.
}

// ScanlineScroll owns a bounded text-window image and reuses its triangle arrays.
// Update selects the visible glyph range; Draw only refreshes the image when that
// range changes. The supplied font image and displacement table stay borrowed.
type ScanlineScroll struct {
	config          ScanlineConfig
	text            *Scrolling
	surface         *ebiten.Image
	positions       []int
	waves           []int
	vertices        []ebiten.Vertex
	indices         []uint16
	triangleOptions ebiten.DrawTrianglesOptions
	letter, decal   int
	displayedLetter int
	bounce          int
	waveStart       int
}

// ScanlineState exposes the visible cursor and row bounce for timed cues or
// editor inspection without advancing or redrawing the scrolling surface.
type ScanlineState struct {
	Letter, Decal, Bounce, WaveStart int
}

// State returns the last validated update state without advancing the scroll.
func (s *ScanlineScroll) State() ScanlineState {
	if s == nil {
		return ScanlineState{}
	}
	return ScanlineState{Letter: s.letter, Decal: s.decal, Bounce: s.bounce, WaveStart: s.waveStart}
}

func NewScanlineScroll(c ScanlineConfig) (*ScanlineScroll, error) {
	if c.Font == nil || c.Font.face.Atlas == nil || c.Text == "" || (len(c.Wave) == 0 && c.Program == nil) || c.ViewportWidth < 1 || c.ViewportHeight < 1 || c.SurfaceWidth < c.ViewportWidth || c.SourceRows < 1 || c.RowHeight < 1 || c.StripHeight < 1 || c.WaveStep < 0 || c.CursorRows < 1 || c.SurfaceWidth > 8192 || c.ViewportHeight > 8192 || c.SourceRows > 8192/c.RowHeight || c.RowHeight%c.StripHeight != 0 || c.Wrap > ScanlineReject || !finite(c.Scale) || c.Scale <= 0 || !finite(c.BounceAmplitude) || c.BounceAmplitude < 0 || !finite(c.BounceRate) || !finite(c.MissingAdvance) || c.MissingAdvance < 0 {
		return nil, fmt.Errorf("scrolling: invalid scanline configuration")
	}
	if c.CursorRows > (c.ViewportHeight+c.RowHeight-1)/c.RowHeight {
		return nil, fmt.Errorf("scrolling: scanline cursor exceeds visible rows")
	}
	strips := (c.ViewportHeight + c.StripHeight - 1) / c.StripHeight
	if strips > 16383 {
		return nil, fmt.Errorf("scrolling: scanline strip budget exceeded")
	}
	glyphs := make([]Glyph, 0, len(c.Text))
	positions := make([]int, 0, len(c.Text))
	position := 0
	for _, r := range c.Text {
		glyph := Glyph{Rune: r, Advance: c.MissingAdvance}
		if img, metrics, ok := c.Font.ExactGlyph(r); ok {
			if !finite(metrics.Advance) || metrics.Advance > 1<<30 {
				return nil, fmt.Errorf("scrolling: scanline glyph advance exceeds budget")
			}
			glyph.Image = img
			glyph.Advance = float64(int(metrics.Advance))
			advance := float64(int(metrics.Advance)) * c.Scale
			if !finite(advance) || advance > float64(int(^uint(0)>>1)-position) {
				return nil, fmt.Errorf("scrolling: scanline text advance overflows")
			}
			position += int(advance)
			positions = append(positions, position)
		} else if c.MissingRune != 0 {
			if img, metrics, found := c.Font.ExactGlyph(c.MissingRune); found {
				glyph.Image = img
				glyph.Advance = float64(int(metrics.Advance))
			}
		}
		glyphs = append(glyphs, glyph)
	}
	if len(positions) == 0 {
		return nil, fmt.Errorf("scrolling: scanline text has no visible glyph")
	}
	text, err := New(Config{Glyphs: glyphs})
	if err != nil {
		return nil, err
	}
	var surface *ebiten.Image
	if c.SurfaceUnmanaged {
		surface = ebiten.NewImageWithOptions(image.Rect(0, 0, c.SurfaceWidth, c.SourceRows*c.RowHeight), &ebiten.NewImageOptions{Unmanaged: true})
	} else {
		surface = ebiten.NewImage(c.SurfaceWidth, c.SourceRows*c.RowHeight)
	}
	s := &ScanlineScroll{
		config: c, text: text, positions: positions,
		surface:  surface,
		waves:    make([]int, (c.ViewportHeight+c.RowHeight-1)/c.RowHeight),
		vertices: make([]ebiten.Vertex, 0, strips*8),
		indices:  make([]uint16, 0, strips*12), displayedLetter: -1,
	}
	if err := s.Update(kit.Frame{}); err != nil {
		s.Close()
		return nil, err
	}
	if c.Wrap == ScanlineAddressRepeat {
		s.triangleOptions.Address = ebiten.AddressRepeat
	}
	return s, nil
}

func (s *ScanlineScroll) position(index int) int {
	if index > 0 && (index <= len(s.positions) || !s.config.ClampCursor) {
		return composite.CumulativeAt(s.positions, index-1, 0)
	}
	return 0
}

func (s *ScanlineScroll) Update(frame kit.Frame) error {
	if s.surface == nil {
		return nil
	}
	clock := float64(frame.Tick)
	if s.config.UseTime {
		clock = frame.Time
	}
	if !finite(clock) || clock < 0 || clock > 1<<52 || clock*float64(s.config.WaveStep) > float64(int(^uint(0)>>1)) {
		return fmt.Errorf("scrolling: scanline clock exceeds exact range")
	}
	start := int(clock * float64(s.config.WaveStep))
	s.waveStart = start
	if s.config.Program != nil {
		s.config.Program.Fill(s.waves, start)
	} else {
		composite.FillCumulative(s.waves, s.config.Wave, start, 0)
	}
	decalX := s.waves[0]
	for row := 1; row < s.config.CursorRows; row++ {
		decalX = min(decalX, s.waves[row])
	}
	decalX = max(0, decalX)
	direction := 0
	if decalX > s.decal {
		direction = 1
	} else if decalX < s.decal {
		direction = -1
	}
	i := 0
	for decalX < s.position(s.letter+i) || s.position(s.letter+i+1) <= decalX {
		if direction == 0 {
			break
		}
		i += direction
		if s.letter+i < 0 || (s.config.ClampCursor && s.letter+i >= len(s.positions)) {
			break
		}
	}
	s.letter += i
	if s.config.ClampCursor {
		s.letter = max(0, min(s.letter, len(s.positions)-1))
	}
	s.decal = s.position(s.letter)
	s.bounce = int(s.config.BounceAmplitude * math.Abs(math.Sin(clock*s.config.BounceRate)))
	return nil
}

func (s *ScanlineScroll) Draw(dst *ebiten.Image) {
	if dst == nil || s.surface == nil {
		return
	}
	if s.letter != s.displayedLetter {
		s.displayedLetter = s.letter
		s.surface.Clear()
		state := s.text.Window(s.letter, float64(s.config.SurfaceWidth)/s.config.Scale)
		state.ScaleX = s.config.Scale
		state.ScaleY = s.config.Scale
		state.X *= s.config.Scale
		s.text.DrawAt(s.surface, state)
	}
	if s.config.Background != nil {
		dst.Fill(s.config.Background)
	}
	s.vertices, s.indices = s.vertices[:0], s.indices[:0]
	for top := 0; top < s.config.ViewportHeight; top += s.config.StripHeight {
		height := min(s.config.StripHeight, s.config.ViewportHeight-top)
		row := top / s.config.RowHeight
		raw := s.waves[row] - s.decal
		srcY := ((row+s.bounce)%s.config.SourceRows)*s.config.RowHeight + top%s.config.RowHeight
		if s.config.Wrap == ScanlineReject {
			if raw >= 0 && raw < s.config.SurfaceWidth-s.config.ViewportWidth {
				s.quad(0, s.config.DestinationY+top, raw, srcY, s.config.ViewportWidth, height)
			}
			continue
		}
		if raw < 0 {
			width := s.config.ViewportWidth + raw
			if width > 0 {
				s.quad(-raw, s.config.DestinationY+top, 0, srcY, width, height)
			}
			continue
		}
		sx := raw % s.config.SurfaceWidth
		if s.config.Wrap == ScanlineSplit && sx >= s.config.SurfaceWidth-s.config.ViewportWidth {
			first := s.config.SurfaceWidth - sx
			s.quad(0, s.config.DestinationY+top, sx, srcY, first, height)
			if second := s.config.ViewportWidth - first; second > 0 {
				s.quad(first, s.config.DestinationY+top, 0, srcY, second, height)
			}
		} else {
			s.quad(0, s.config.DestinationY+top, sx, srcY, s.config.ViewportWidth, height)
		}
	}
	if len(s.indices) > 0 {
		dst.DrawTriangles(s.vertices, s.indices, s.surface, &s.triangleOptions)
	}
}

func (s *ScanlineScroll) quad(dx, dy, sx, sy, width, height int) {
	base := uint16(len(s.vertices))
	x0, y0 := float32(dx), float32(dy)
	x1, y1 := float32(dx+width), float32(dy+height)
	u0, v0 := float32(sx), float32(sy)
	u1, v1 := float32(sx+width), float32(sy+height)
	s.vertices = append(s.vertices,
		ebiten.Vertex{DstX: x0, DstY: y0, SrcX: u0, SrcY: v0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		ebiten.Vertex{DstX: x1, DstY: y0, SrcX: u1, SrcY: v0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		ebiten.Vertex{DstX: x1, DstY: y1, SrcX: u1, SrcY: v1, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		ebiten.Vertex{DstX: x0, DstY: y1, SrcX: u0, SrcY: v1, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
	)
	if s.config.AlternateDiagonal {
		s.indices = append(s.indices, base, base+1, base+3, base+1, base+2, base+3)
	} else {
		s.indices = append(s.indices, base, base+1, base+2, base, base+2, base+3)
	}
}

func (s *ScanlineScroll) Close() error {
	if s.surface != nil {
		s.surface.Deallocate()
		s.surface = nil
	}
	return nil
}
