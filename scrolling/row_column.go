package scrolling

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// RowColumnDraw selects the original image-copy or batched-triangle sampling
// topology. Both preserve ordered row and column passes on separate surfaces.
type RowColumnDraw uint8

const (
	RowColumnQuads RowColumnDraw = iota
	RowColumnImages
)

// RowColumnCull selects the text-window policy before its two image passes.
type RowColumnCull uint8

const (
	RowColumnBounds RowColumnCull = iota
	RowColumnMapper
)

// RowColumnConfig composes a fixed-advance bitmap message, a source-row X
// lookup and a destination-column Y cosine. Wave, font, viewport, materials,
// placement and independent clocks are data; no screen-specific formula is
// required in the production. TextSpeed and ColumnSpeed are per Update.
type RowColumnConfig struct {
	Font              *Atlas
	Text              string
	Advance           float64
	ViewportWidth     int
	Height            int
	WorkWidth         int
	RowHeight         int
	ColumnWidth       int
	Wave              []float64
	RowBias           float64
	RowPhaseStep      int
	TextSpeed         float64
	TextRestartX      float64
	InitialMultiplier float64
	ColumnY           float64
	ColumnBias        float64
	ColumnAmplitude   float64
	ColumnSpatial     float64
	ColumnSpeed       float64
	ColumnWrap        float64 // Optional historical phase reset after each cycle.
	ColumnBiasInside  bool
	Cull              RowColumnCull
	DrawMode          RowColumnDraw
}

// RowColumn owns its cached glyph layout, two bounded intermediate images,
// independent text/wave clocks and optional reusable triangle batches.
type RowColumn struct {
	config         RowColumnConfig
	text           *Scrolling
	glyphCount     int
	textWidth      float64
	work           *ebiten.Image
	deformed       *ebiten.Image
	rowBatch       *composite.QuadBatch
	colBatch       *composite.QuadBatch
	mapper         Mapper
	x, columnPhase float64
	rowPhase       int
	multiplier     float64
}

func NewRowColumn(c RowColumnConfig) (*RowColumn, error) {
	if c.Font == nil || c.Font.face.Atlas == nil || c.Text == "" || !finite(c.Advance) || c.Advance <= 0 || c.ViewportWidth < 1 || c.ViewportWidth > 8192 || c.Height < 1 || c.Height > 8192 || c.WorkWidth < c.ViewportWidth || c.WorkWidth > 8192 || c.RowHeight < 1 || c.ColumnWidth < 1 || c.Height%c.RowHeight != 0 || c.ViewportWidth%c.ColumnWidth != 0 || len(c.Wave) == 0 || len(c.Wave) > 1<<20 || c.Cull > RowColumnMapper || c.DrawMode > RowColumnImages {
		return nil, fmt.Errorf("scrolling: invalid row-column dimensions or font")
	}
	for _, value := range []float64{c.RowBias, c.TextSpeed, c.TextRestartX, c.InitialMultiplier, c.ColumnY, c.ColumnBias, c.ColumnAmplitude, c.ColumnSpatial, c.ColumnSpeed, c.ColumnWrap} {
		if !finite(value) {
			return nil, fmt.Errorf("scrolling: nonfinite row-column parameter")
		}
	}
	if c.TextSpeed < 0 || c.InitialMultiplier < 0 || c.ColumnSpeed < 0 || c.ColumnWrap < 0 {
		return nil, fmt.Errorf("scrolling: negative row-column speed")
	}
	if c.InitialMultiplier == 0 {
		c.InitialMultiplier = 1
	}
	for _, offset := range c.Wave {
		if !finite(offset) || !finite(offset+c.RowBias) || offset+c.RowBias < 0 || offset+c.RowBias+float64(c.ViewportWidth) > float64(c.WorkWidth) {
			return nil, fmt.Errorf("scrolling: row lookup exceeds work image")
		}
	}
	glyphs := make([]Glyph, 0, len(c.Text))
	for _, r := range c.Text {
		img, _, ok := c.Font.Glyph(r)
		if !ok {
			img = nil
		}
		glyphs = append(glyphs, Glyph{Image: img, Rune: r, Advance: c.Advance})
	}
	text, err := New(Config{Glyphs: glyphs})
	if err != nil {
		return nil, err
	}
	c.Wave = append([]float64(nil), c.Wave...)
	r := &RowColumn{
		config: c, text: text, glyphCount: len(glyphs), textWidth: text.Length(),
		work:       ebiten.NewImage(c.WorkWidth, c.Height),
		deformed:   ebiten.NewImage(c.ViewportWidth, c.Height),
		multiplier: c.InitialMultiplier,
	}
	if c.DrawMode == RowColumnQuads {
		r.rowBatch = composite.NewQuadBatch(c.Height / c.RowHeight)
		r.colBatch = composite.NewQuadBatch(c.ViewportWidth / c.ColumnWidth)
		r.rowBatch.AlternateDiagonal = true
		r.colBatch.AlternateDiagonal = true
	}
	r.mapper = func(sample Sample, _ *ebiten.DrawImageOptions) bool {
		return sample.X > -r.config.Advance && sample.X < float64(r.config.WorkWidth)
	}
	return r, nil
}

func (r *RowColumn) SetSpeedMultiplier(value float64) error {
	if !finite(value) || value < 0 {
		return fmt.Errorf("scrolling: invalid row-column multiplier")
	}
	r.multiplier = value
	return nil
}

func (r *RowColumn) Update(kit.Frame) error {
	if r.work == nil {
		return nil
	}
	nextX := r.x - r.config.TextSpeed*r.multiplier
	nextColumn := r.columnPhase + r.config.ColumnSpeed*r.multiplier
	if !finite(nextX) || !finite(nextColumn) {
		return fmt.Errorf("scrolling: row-column clock overflows")
	}
	r.x = nextX
	if r.x < -r.textWidth {
		r.x = r.config.TextRestartX
	}
	r.rowPhase += r.config.RowPhaseStep
	r.columnPhase = nextColumn
	if r.config.ColumnWrap > 0 && r.columnPhase >= r.config.ColumnWrap {
		r.columnPhase -= r.config.ColumnWrap
	}
	return nil
}

func (r *RowColumn) Draw(dst *ebiten.Image) {
	if dst == nil || r.work == nil {
		return
	}
	r.work.Clear()
	r.deformed.Clear()
	state := IdentityState()
	state.X = r.x
	if r.config.Cull == RowColumnBounds {
		state.First = max(0, int(math.Floor((-r.config.Advance-r.x)/r.config.Advance))+1)
		state.End = min(r.glyphCount, int(math.Ceil((float64(r.config.WorkWidth)-r.x)/r.config.Advance)))
	} else {
		state.Map = r.mapper
	}
	r.text.DrawAt(r.work, state)
	if r.rowBatch != nil {
		r.rowBatch.Begin(r.deformed, r.work)
	}
	for row := 0; row < r.config.Height/r.config.RowHeight; row++ {
		index := (r.rowPhase + row) % len(r.config.Wave)
		if index < 0 {
			index += len(r.config.Wave)
		}
		x := int(r.config.Wave[index] + r.config.RowBias)
		y := row * r.config.RowHeight
		source := image.Rect(x, y, x+r.config.ViewportWidth, y+r.config.RowHeight)
		if r.rowBatch != nil {
			r.rowBatch.Rect(source, 0, float32(y), float32(r.config.ViewportWidth), float32(r.config.RowHeight))
		} else {
			op := ebiten.DrawImageOptions{}
			op.GeoM.Translate(0, float64(y))
			composite.Instance{Image: r.work, Source: &source, Options: op}.Draw(r.deformed)
		}
	}
	if r.rowBatch != nil {
		r.rowBatch.Flush()
		r.colBatch.Begin(dst, r.deformed)
	}
	for column := 0; column < r.config.ViewportWidth/r.config.ColumnWidth; column++ {
		x := column * r.config.ColumnWidth
		term := math.Cos(r.columnPhase+float64(column)*r.config.ColumnSpatial) * r.config.ColumnAmplitude
		y := r.config.ColumnY + r.config.ColumnBias + term
		if r.config.ColumnBiasInside {
			y = r.config.ColumnY + (r.config.ColumnBias + term)
		}
		source := image.Rect(x, 0, x+r.config.ColumnWidth, r.config.Height)
		if r.colBatch != nil {
			r.colBatch.Rect(source, float32(x), float32(y), float32(r.config.ColumnWidth), float32(r.config.Height))
		} else {
			op := ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), y)
			composite.Instance{Image: r.deformed, Source: &source, Options: op}.Draw(dst)
		}
	}
	if r.colBatch != nil {
		r.colBatch.Flush()
	}
}

func (r *RowColumn) Close() error {
	if r == nil {
		return nil
	}
	if r.work != nil {
		r.work.Deallocate()
		r.work = nil
	}
	if r.deformed != nil {
		r.deformed.Deallocate()
		r.deformed = nil
	}
	return nil
}
