package effects

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/render"
)

// CheckerboardComposition selects the XOR operation used to make floor cells.
// MaskXOR batches each set of stripes on a separate surface. DirectXOR draws
// individual horizontal stripes over the vertical stripes with triangle XOR.
type CheckerboardComposition uint8

const (
	CheckerboardMaskXOR CheckerboardComposition = iota
	CheckerboardDirectXOR
)

// CheckerboardClockOrder specifies when the oscillator is sampled. BeforeMove
// uses the new oscillator values for the current frame; AfterMove uses the old
// values, then advances them for the next frame.
type CheckerboardClockOrder uint8

const (
	CheckerboardBeforeMove CheckerboardClockOrder = iota
	CheckerboardAfterMove
)

// CheckerboardWrap selects either mathematical modulo or a single strict wrap
// at the boundary. The latter retains an exact-period coordinate for one frame.
type CheckerboardWrap uint8

const (
	CheckerboardModuloWrap CheckerboardWrap = iota
	CheckerboardSingleWrap
)

// PerspectiveCheckerboardConfig describes a complete moving floor. Vertical
// coordinates describe one projected stripe at index zero; TopStep and
// BottomStep place later stripes. Horizontal bands use a perspective divide:
// Horizon + FocalLength/(FocalLength + depth - offsetY)*ProjectionHeight.
// Position optionally supplies independent offsets for each tick, replacing
// the oscillator and wrap motion. It can be any deterministic Go trajectory.
// Zero Position means the built-in two-oscillator motion is used.
type PerspectiveCheckerboardConfig struct {
	Width, Height                                      int
	Color                                              color.RGBA
	VerticalCount                                      int
	TopLeftX, TopRightX, BottomRightX, BottomLeftX     float64
	TopY, BottomY, TopStep, BottomStep                 float64
	TopOffsetScale, BottomOffsetScale                  float64
	HorizontalFirst, HorizontalCount                   int
	Horizon, FocalLength, ProjectionHeight             float64
	DepthStep, BandDepth                               float64
	SpeedAmplitude, SpeedDivisor, SpeedPhaseStep       float64
	LateralAmplitude, LateralDivisor, LateralPhaseStep float64
	InitialSpeed, InitialLateral                       float64
	LateralRate, ForwardRate                           float64
	XPeriod, YPeriod                                   float64
	ClockOrder                                         CheckerboardClockOrder
	Wrap                                               CheckerboardWrap
	Composition                                        CheckerboardComposition
	AntiAlias, Unmanaged                               bool
	Filter                                             ebiten.Filter
	X, Y, Opacity                                      float64
	Position                                           func(tick uint64) (x, y float64)
}

// DefaultPerspectiveCheckerboardConfig returns an editable 320 by 80 floor.
// Its opaque purple geometry matches the shared 3D DOC scene; callers choose
// the composition, placement, clock and opacity required by their production.
func DefaultPerspectiveCheckerboardConfig() PerspectiveCheckerboardConfig {
	return PerspectiveCheckerboardConfig{
		Width: 320, Height: 80, Color: color.RGBA{R: 136, B: 136, A: 255},
		VerticalCount: 11,
		TopLeftX:      -8, TopRightX: 8, BottomRightX: -752, BottomLeftX: -848,
		TopY: 0, BottomY: 80, TopStep: 32, BottomStep: 192,
		TopOffsetScale: 1, BottomOffsetScale: 6,
		HorizontalFirst: -2, HorizontalCount: 10,
		Horizon: -20, FocalLength: 250, ProjectionHeight: 50,
		DepthStep: 64, BandDepth: 32,
		SpeedAmplitude: -1, SpeedDivisor: 40, SpeedPhaseStep: .16,
		LateralAmplitude: 128, LateralDivisor: 40, LateralPhaseStep: .8,
		InitialSpeed: 1, LateralRate: .01, ForwardRate: 315 * .032,
		XPeriod: 32, YPeriod: 64, X: 32, Y: 149, Opacity: 1,
	}
}

// Validate rejects geometry that cannot be projected or batched safely.
func (c PerspectiveCheckerboardConfig) Validate() error {
	if c.Width <= 0 || c.Height <= 0 || c.VerticalCount <= 0 || c.HorizontalCount <= 0 ||
		c.VerticalCount > 16383 || c.HorizontalCount > 16383 || c.FocalLength <= 0 ||
		c.SpeedDivisor == 0 || c.LateralDivisor == 0 || c.XPeriod <= 0 || c.YPeriod <= 0 ||
		c.Opacity < 0 || c.Opacity > 1 || c.ClockOrder > CheckerboardAfterMove ||
		c.Wrap > CheckerboardSingleWrap || c.Composition > CheckerboardDirectXOR {
		return fmt.Errorf("effects: invalid perspective checkerboard geometry or motion")
	}
	for _, v := range []float64{
		c.TopLeftX, c.TopRightX, c.BottomRightX, c.BottomLeftX, c.TopY, c.BottomY,
		c.TopStep, c.BottomStep, c.TopOffsetScale, c.BottomOffsetScale, c.Horizon,
		c.FocalLength, c.ProjectionHeight, c.DepthStep, c.BandDepth,
		c.SpeedAmplitude, c.SpeedDivisor, c.SpeedPhaseStep, c.LateralAmplitude,
		c.LateralDivisor, c.LateralPhaseStep, c.InitialSpeed, c.InitialLateral,
		c.LateralRate, c.ForwardRate, c.XPeriod, c.YPeriod, c.X, c.Y, c.Opacity,
	} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("effects: non-finite perspective checkerboard parameter")
		}
	}
	return nil
}

// PerspectiveCheckerboard owns its motion, geometry and bounded GPU surfaces.
// Update advances once per logical frame. Draw never advances the animation,
// so the same floor can be layered, transformed or drawn more than once.
type PerspectiveCheckerboard struct {
	config             PerspectiveCheckerboardConfig
	board, mask, white *ebiten.Image
	batch              *render.Batch
	vertices           []ebiten.Vertex
	indices            []uint16
	xOffset, yOffset   float64
	speed, lateral     float64
	speedPhase         float64
	lateralPhase       float64
	tick               uint64
	closed             bool
}

func NewPerspectiveCheckerboard(c PerspectiveCheckerboardConfig) (*PerspectiveCheckerboard, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	newImage := func(w, h int) *ebiten.Image {
		if c.Unmanaged {
			return ebiten.NewImageWithOptions(image.Rect(0, 0, w, h), &ebiten.NewImageOptions{Unmanaged: true})
		}
		return ebiten.NewImage(w, h)
	}
	p := &PerspectiveCheckerboard{
		config: c, board: newImage(c.Width, c.Height), white: newImage(1, 1),
		speed: c.InitialSpeed, lateral: c.InitialLateral,
	}
	p.white.Fill(color.White)
	if c.Composition == CheckerboardMaskXOR {
		p.mask = newImage(c.Width, c.Height)
		capacity := max(c.VerticalCount, c.HorizontalCount) * 4
		p.vertices = make([]ebiten.Vertex, 0, capacity)
		p.indices = make([]uint16, 0, capacity*6/4)
	} else {
		p.batch = render.NewBatch(2)
		p.batch.Options.AntiAlias = c.AntiAlias
	}
	return p, nil
}

func (p *PerspectiveCheckerboard) Update(kit.Frame) error {
	if p == nil || p.closed {
		return nil
	}
	p.tick++
	if p.config.Position != nil {
		x, y := p.config.Position(p.tick)
		if math.IsNaN(x) || math.IsInf(x, 0) || math.IsNaN(y) || math.IsInf(y, 0) {
			return fmt.Errorf("effects: non-finite perspective checkerboard position")
		}
		p.xOffset, p.yOffset = x, y
		return nil
	}
	if p.config.ClockOrder == CheckerboardBeforeMove {
		p.advanceOscillator()
	}
	p.xOffset = p.wrap(p.xOffset+p.lateral*p.speed*p.config.LateralRate, p.config.XPeriod)
	p.yOffset = p.wrap(p.yOffset+p.speed*p.config.ForwardRate, p.config.YPeriod)
	if p.config.ClockOrder == CheckerboardAfterMove {
		p.advanceOscillator()
	}
	return nil
}

func (p *PerspectiveCheckerboard) advanceOscillator() {
	c := p.config
	p.speed = c.SpeedAmplitude * math.Cos(p.speedPhase/c.SpeedDivisor)
	p.speedPhase += c.SpeedPhaseStep
	p.lateral = c.LateralAmplitude * math.Cos(p.lateralPhase/c.LateralDivisor)
	p.lateralPhase += c.LateralPhaseStep
}

func (p *PerspectiveCheckerboard) wrap(v, period float64) float64 {
	if p.config.Wrap == CheckerboardSingleWrap {
		if v > period {
			v -= period
		}
		if v < 0 {
			v += period
		}
		return v
	}
	v = math.Mod(v, period)
	if v < 0 {
		v += period
	}
	return v
}

// Offset returns the current lateral and forward offsets for synchronization.
func (p *PerspectiveCheckerboard) Offset() (x, y float64) {
	return p.xOffset, p.yOffset
}

// Draw renders the floor and places it on dst with the configured opacity.
func (p *PerspectiveCheckerboard) Draw(dst *ebiten.Image) {
	if p == nil || p.closed || dst == nil {
		return
	}
	c := p.config
	p.board.Clear()
	if p.mask != nil {
		p.mask.Clear()
	}
	p.beginQuads(ebiten.BlendSourceOver)
	for i := 0; i < c.VerticalCount; i++ {
		top := float64(i)*c.TopStep + p.xOffset*c.TopOffsetScale
		bottom := float64(i)*c.BottomStep + p.xOffset*c.BottomOffsetScale
		p.quad(p.board, [4][2]float64{
			{c.TopLeftX + top, c.TopY}, {c.TopRightX + top, c.TopY},
			{c.BottomRightX + bottom, c.BottomY}, {c.BottomLeftX + bottom, c.BottomY},
		})
	}
	p.flushQuads(p.board)
	horizontalDst := p.board
	if p.mask != nil {
		horizontalDst = p.mask
		p.beginQuads(ebiten.BlendSourceOver)
	} else {
		p.beginQuads(ebiten.BlendXor)
	}
	for j := 0; j < c.HorizontalCount; j++ {
		i := c.HorizontalFirst + j
		depth := float64(i)*c.DepthStep - p.yOffset
		near, far := c.FocalLength+depth, c.FocalLength+depth+c.BandDepth
		if near == 0 || far == 0 {
			continue
		}
		y1 := c.Horizon + c.FocalLength/near*c.ProjectionHeight
		y2 := c.Horizon + c.FocalLength/far*c.ProjectionHeight
		p.quad(horizontalDst, [4][2]float64{{0, y1}, {float64(c.Width), y1}, {float64(c.Width), y2}, {0, y2}})
	}
	p.flushQuads(horizontalDst)
	if p.mask != nil {
		op := ebiten.DrawImageOptions{CompositeMode: ebiten.CompositeModeXor}
		p.board.DrawImage(p.mask, &op)
	}
	op := ebiten.DrawImageOptions{Filter: c.Filter}
	op.GeoM.Translate(c.X, c.Y)
	op.ColorScale.ScaleAlpha(float32(c.Opacity))
	dst.DrawImage(p.board, &op)
}

func (p *PerspectiveCheckerboard) beginQuads(blend ebiten.Blend) {
	if p.batch != nil {
		p.batch.Options.Blend = blend
		return
	}
	p.vertices = p.vertices[:0]
	p.indices = p.indices[:0]
}

func (p *PerspectiveCheckerboard) quad(dst *ebiten.Image, points [4][2]float64) {
	if p.batch != nil {
		p.batch.Begin(dst, p.white)
		var vertices [4]ebiten.Vertex
		for i, point := range points {
			vertices[i] = render.Vertex(point[0], point[1], .5, .5, p.config.Color)
		}
		p.batch.Quad(vertices)
		p.batch.Flush()
		return
	}
	c := p.config.Color
	base := uint16(len(p.vertices))
	for _, point := range points {
		p.vertices = append(p.vertices, ebiten.Vertex{
			DstX: float32(point[0]), DstY: float32(point[1]),
			ColorR: float32(c.R) / 255, ColorG: float32(c.G) / 255,
			ColorB: float32(c.B) / 255, ColorA: float32(c.A) / 255,
		})
	}
	p.indices = append(p.indices, base, base+1, base+2, base+2, base+3, base)
}

func (p *PerspectiveCheckerboard) flushQuads(dst *ebiten.Image) {
	if p.batch != nil {
		return
	}
	op := ebiten.DrawTrianglesOptions{FillRule: ebiten.FillAll}
	dst.DrawTriangles(p.vertices, p.indices, p.white, &op)
}

func (p *PerspectiveCheckerboard) Close() error {
	if p == nil || p.closed {
		return nil
	}
	p.closed = true
	p.board.Deallocate()
	if p.mask != nil {
		p.mask.Deallocate()
	}
	p.white.Deallocate()
	return nil
}
