package scrolling

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// RowBandPass shifts equal-height source strips into a separately sized output.
// An optional wave is sampled by row and advances by WaveStep per Update.
// VerticalBase + VerticalAmplitude*cos(phase/VerticalDivisor) moves the entire
// pass. VerticalDivisor zero disables the cosine term.
type RowBandPass struct {
	Width, Height, Thickness int
	Wave                     []float64
	WaveStep                 int
	VerticalBase             float64
	VerticalAmplitude        float64
	VerticalDivisor          float64
	VerticalStep             float64
}

// RowBandsConfig combines a circular bitmap text transport with ordered row
// image passes. The final crop and placement belong to the complete effect.
type RowBandsConfig struct {
	Font         *Atlas
	Text         string
	Advance      float64
	FallbackRect image.Rectangle // Literal atlas cell for unsupported/blank runes.
	WorkWidth    int
	Height       int
	TextSpeed    float64 // Pixels per Update; zero holds the first character.
	Passes       []RowBandPass
	Crop         image.Rectangle // Empty selects the complete final image.
	X, Y         float64
}

type rowBandStage struct {
	surface *ebiten.Image
	warp    *composite.RowWarp
}

// RowBands owns a bounded text surface, one image per pass and a cached final
// crop. Font atlas images remain borrowed. Draw never advances any clock.
type RowBands struct {
	config   RowBandsConfig
	text     *Scrolling
	runes    []rune
	work     *ebiten.Image
	stages   []rowBandStage
	visible  *ebiten.Image
	position float64
	mapper   Mapper
}

func NewRowBands(c RowBandsConfig) (*RowBands, error) {
	if c.Font == nil || c.Font.face.Atlas == nil || c.Text == "" || !finite(c.Advance) || c.Advance <= 0 || c.WorkWidth < 1 || c.WorkWidth > 8192 || c.Height < 1 || c.Height > 8192 || !finite(c.TextSpeed) || !finite(c.X) || !finite(c.Y) || len(c.Passes) > 16 {
		return nil, fmt.Errorf("scrolling: invalid row-bands text or dimensions")
	}
	var fallback *ebiten.Image
	if !c.FallbackRect.Empty() {
		if !c.FallbackRect.In(c.Font.face.Atlas.Bounds()) {
			return nil, fmt.Errorf("scrolling: fallback cell is outside the atlas")
		}
		fallback = c.Font.face.Atlas.SubImage(c.FallbackRect).(*ebiten.Image)
	}
	glyphs := make([]Glyph, 0, len(c.Text))
	runes := make([]rune, 0, len(c.Text))
	for _, character := range c.Text {
		img, _, _ := c.Font.Glyph(character)
		if img == nil {
			img = fallback
		}
		glyphs = append(glyphs, Glyph{Image: img, Rune: character, Advance: c.Advance})
		runes = append(runes, character)
	}
	text, err := New(Config{Glyphs: glyphs})
	if err != nil {
		return nil, err
	}
	r := &RowBands{config: c, text: text, runes: runes, work: ebiten.NewImage(c.WorkWidth, c.Height)}
	previousHeight := c.Height
	for _, pass := range c.Passes {
		if pass.Width < 1 || pass.Width > 8192 || pass.Height < 1 || pass.Height > 8192 || pass.Thickness < 1 || pass.Thickness > previousHeight || !finite(pass.VerticalBase) || !finite(pass.VerticalAmplitude) || !finite(pass.VerticalDivisor) || !finite(pass.VerticalStep) || pass.VerticalDivisor < 0 || len(pass.Wave) > 1<<20 {
			r.Close()
			return nil, fmt.Errorf("scrolling: invalid row-band pass")
		}
		warp, err := composite.NewRowWarp(composite.RowWarpConfig{
			Mode: composite.RowWarpDestinationX, Thickness: pass.Thickness,
			Wave: pass.Wave, WaveStep: pass.WaveStep,
			VerticalBase: pass.VerticalBase, VerticalAmplitude: pass.VerticalAmplitude,
			VerticalDivisor: pass.VerticalDivisor, VerticalStep: pass.VerticalStep,
		})
		if err != nil {
			r.Close()
			return nil, err
		}
		r.stages = append(r.stages, rowBandStage{surface: ebiten.NewImage(pass.Width, pass.Height), warp: warp})
		previousHeight = pass.Height
	}
	final := r.work
	if len(r.stages) > 0 {
		final = r.stages[len(r.stages)-1].surface
	}
	if c.Crop.Empty() {
		c.Crop = final.Bounds()
	}
	if !c.Crop.In(final.Bounds()) {
		r.Close()
		return nil, fmt.Errorf("scrolling: final row-band crop exceeds its surface")
	}
	r.visible = final.SubImage(c.Crop).(*ebiten.Image)
	r.config.Crop = c.Crop
	r.mapper = func(sample Sample, _ *ebiten.DrawImageOptions) bool {
		return sample.X >= -r.config.Advance && sample.X < float64(r.config.WorkWidth)+r.config.Advance
	}
	return r, nil
}

// CursorRune exposes the character at the current transport position so a
// production can trigger a scene cue without reimplementing text indexing.
func (r *RowBands) CursorRune() rune {
	if r == nil || len(r.runes) == 0 {
		return 0
	}
	index := int(r.position / r.config.Advance)
	if index < 0 || index >= len(r.runes) {
		return 0
	}
	return r.runes[index]
}

func (r *RowBands) Update(kit.Frame) error {
	if r == nil || r.work == nil {
		return nil
	}
	period := float64(len(r.runes)) * r.config.Advance
	next := math.Mod(r.position+r.config.TextSpeed, period)
	if !finite(next) {
		return fmt.Errorf("scrolling: row-band text clock overflows")
	}
	if next < 0 {
		next += period
	}
	r.position = next
	for i := range r.stages {
		if err := r.stages[i].warp.Step(); err != nil {
			return err
		}
	}
	return nil
}

func (r *RowBands) Draw(dst *ebiten.Image) {
	if r == nil || dst == nil || r.work == nil {
		return
	}
	r.work.Clear()
	first := int(r.position / r.config.Advance)
	offset := math.Mod(r.position, r.config.Advance)
	state := IdentityState()
	state.X = -offset - float64(first)*r.config.Advance
	state.First = first
	state.End = first + int(float64(r.config.WorkWidth)/r.config.Advance) + 3
	state.Cycle = true
	state.Map = r.mapper
	r.text.DrawAt(r.work, state)
	source := r.work
	for i := range r.stages {
		stage := &r.stages[i]
		stage.surface.Clear()
		stage.warp.DrawInto(stage.surface, source)
		source = stage.surface
	}
	op := ebiten.DrawImageOptions{}
	op.GeoM.Translate(r.config.X, r.config.Y)
	dst.DrawImage(r.visible, &op)
}

func (r *RowBands) Close() error {
	if r != nil {
		if r.work != nil {
			r.work.Deallocate()
			r.work = nil
		}
		for i := range r.stages {
			if r.stages[i].surface != nil {
				r.stages[i].surface.Deallocate()
				r.stages[i].surface = nil
			}
		}
		r.visible = nil
	}
	return nil
}
