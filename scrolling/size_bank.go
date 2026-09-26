package scrolling

import (
	"errors"
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
)

// SizeBankLayer describes one rendering size of a synchronized text program.
// Source artwork, atlas order, scaling and vertical repetition are independent
// for every layer. SurfaceHeight zero uses CellHeight*ScaleY.
type SizeBankLayer struct {
	Name                         string
	Image                        *ebiten.Image
	CellWidth, CellHeight        int
	Columns, Count               int
	ScaleX, ScaleY               float64
	SurfaceHeight                int
	RepeatY, RepeatStep, Repeats int
	Lookup                       func(rune) int
}

// SizeBankConfig binds a controlled text program to one shared transport.
// BaseSpeeds are reference pixels per simulation tick for each active layer.
// Lookup maps a non-space character to an atlas tile when a layer has no
// override. The text and decoder are parsed once during construction.
type SizeBankConfig struct {
	Text            string
	Controls        scrolltext.Decoder
	InitialFont     string
	ViewportWidth   int
	Layers          []SizeBankLayer
	Lookup          func(rune) int
	BaseSpeeds      []float64
	StartOffset     float64
	SpeedMultiplier float64
	Lookahead       int
	WrapInclusive   bool
}

type sizeBankLayer struct {
	config   SizeBankLayer
	renderer *Scrolling
	canvas   *ebiten.Image
}

// SizeBank owns cached glyph views and scaled canvases. It advances and draws
// only the selected layer, while all layer offsets remain synchronized.
type SizeBank struct {
	clock  *motion.ScaledTextClock
	layers []sizeBankLayer
	clip   Mapper
}

func NewSizeBank(c SizeBankConfig) (*SizeBank, error) {
	if c.Text == "" || c.ViewportWidth < 1 || c.ViewportWidth > 8192 ||
		len(c.Layers) == 0 || len(c.Layers) > 256 || len(c.BaseSpeeds) != len(c.Layers) {
		return nil, fmt.Errorf("scrolling: invalid size bank text or layers")
	}
	program, err := scrolltext.NewFontProgram(c.Text, c.Controls, c.InitialFont)
	if err != nil {
		return nil, err
	}
	if program.Len() == 0 {
		return nil, fmt.Errorf("scrolling: empty size bank text")
	}
	names := make(map[string]int, len(c.Layers))
	scales := make([]float64, len(c.Layers))
	for i, layer := range c.Layers {
		if layer.Name == "" || layer.Image == nil || layer.CellWidth < 1 || layer.CellHeight < 1 ||
			layer.Repeats < 1 || layer.Repeats > 8192 || layer.RepeatStep < 0 ||
			!finite(layer.ScaleX) || layer.ScaleX <= 0 || !finite(layer.ScaleY) || layer.ScaleY <= 0 {
			return nil, fmt.Errorf("scrolling: invalid size bank layer %d", i)
		}
		if _, exists := names[layer.Name]; exists {
			return nil, fmt.Errorf("scrolling: duplicate size bank layer %q", layer.Name)
		}
		names[layer.Name] = i
		scales[i] = layer.ScaleX
	}
	initial, ok := names[c.InitialFont]
	if !ok {
		return nil, fmt.Errorf("scrolling: missing initial size bank font")
	}
	fontAt := make([]int, program.Len())
	for i := range fontAt {
		index, found := names[program.FontAt(i)]
		if !found {
			return nil, fmt.Errorf("scrolling: unknown size bank font at glyph %d", i)
		}
		fontAt[i] = index
	}
	clock, err := motion.NewScaledTextClock(motion.ScaledTextClockConfig{
		FontAt: fontAt, Scales: scales, BaseSpeeds: c.BaseSpeeds,
		ViewportWidth: float64(c.ViewportWidth), TileWidth: float64(c.Layers[0].CellWidth),
		StartOffset: c.StartOffset, SpeedMultiplier: c.SpeedMultiplier,
		InitialBank: initial, Lookahead: c.Lookahead, WrapInclusive: c.WrapInclusive,
	})
	if err != nil {
		return nil, err
	}
	bank := &SizeBank{clock: clock, layers: make([]sizeBankLayer, len(c.Layers))}
	limit := float64(c.ViewportWidth)
	bank.clip = func(s Sample, _ *ebiten.DrawImageOptions) bool { return s.X < limit }
	for i, layer := range c.Layers {
		if layer.CellWidth != c.Layers[0].CellWidth {
			bank.Close()
			return nil, fmt.Errorf("scrolling: size bank layers need a common source advance")
		}
		lookup := layer.Lookup
		if lookup == nil {
			lookup = c.Lookup
		}
		if lookup == nil {
			bank.Close()
			return nil, fmt.Errorf("scrolling: missing size bank atlas order")
		}
		columns := layer.Columns
		if columns == 0 {
			columns = layer.Image.Bounds().Dx() / layer.CellWidth
		}
		count := layer.Count
		if count == 0 {
			count = columns * (layer.Image.Bounds().Dy() / layer.CellHeight)
		}
		frames, frameErr := GridImages(layer.Image, image.Pt(layer.CellWidth, layer.CellHeight), columns, count)
		if frameErr != nil {
			bank.Close()
			return nil, frameErr
		}
		masked := []rune(program.MaskedText(layer.Name, ' '))
		if len(masked) != len(fontAt) {
			bank.Close()
			return nil, fmt.Errorf("scrolling: size bank masks changed character positions")
		}
		images := make([]*ebiten.Image, len(masked))
		for j, r := range masked {
			if r == ' ' {
				continue
			}
			index := lookup(r)
			if index >= 0 && index < len(frames) {
				images[j] = frames[index]
			}
		}
		renderer, renderErr := FromImages(images, float64(layer.CellWidth))
		if renderErr != nil {
			bank.Close()
			return nil, renderErr
		}
		height := layer.SurfaceHeight
		if height == 0 {
			scaledHeight := float64(layer.CellHeight) * layer.ScaleY
			if !finite(scaledHeight) || scaledHeight < 1 || scaledHeight > 8192 {
				bank.Close()
				return nil, fmt.Errorf("scrolling: invalid size bank surface height")
			}
			height = int(math.Round(scaledHeight))
		}
		if height < 1 || height > 8192 {
			bank.Close()
			return nil, fmt.Errorf("scrolling: invalid size bank surface height")
		}
		bank.layers[i] = sizeBankLayer{config: layer, renderer: renderer,
			canvas: ebiten.NewImage(c.ViewportWidth, height)}
	}
	return bank, nil
}

func (b *SizeBank) Update(kit.Frame) error {
	b.clock.Step()
	index := b.clock.ActiveBank()
	layer := &b.layers[index]
	layer.canvas.Clear()
	state := IdentityState()
	state.X = b.clock.Offset(index)
	state.ScaleX, state.ScaleY = layer.config.ScaleX, layer.config.ScaleY
	if state.X < 0 {
		state.First = int(math.Floor(-state.X / (float64(layer.config.CellWidth) * state.ScaleX)))
	}
	state.Map = b.clip
	layer.renderer.DrawAt(layer.canvas, state)
	return nil
}

// Draw repeats the active cached text strip using its authored row positions.
// The destination is caller-owned and is not cleared by this effect.
func (b *SizeBank) Draw(dst *ebiten.Image) {
	if b == nil || dst == nil {
		return
	}
	layer := &b.layers[b.clock.ActiveBank()]
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(0, float64(layer.config.RepeatY))
	for range layer.config.Repeats {
		dst.DrawImage(layer.canvas, &op)
		op.GeoM.Translate(0, float64(layer.config.RepeatStep))
	}
}

func (b *SizeBank) ActiveBank() int                        { return b.clock.ActiveBank() }
func (b *SizeBank) Offset(index int) float64               { return b.clock.Offset(index) }
func (b *SizeBank) Clock() *motion.ScaledTextClock         { return b.clock }
func (b *SizeBank) SetSpeedMultiplier(value float64) error { return b.clock.SetSpeedMultiplier(value) }
func (b *SizeBank) Canvas() *ebiten.Image                  { return b.layers[b.clock.ActiveBank()].canvas }

func (b *SizeBank) Close() error {
	if b == nil {
		return nil
	}
	var closeErr error
	for i := range b.layers {
		closeErr = errors.Join(closeErr, kit.Close(b.layers[i].renderer))
		if b.layers[i].canvas != nil {
			b.layers[i].canvas.Deallocate()
			b.layers[i].canvas = nil
		}
	}
	return closeErr
}
