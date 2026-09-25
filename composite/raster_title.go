package composite

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// RasterTitleMode selects a small precomposed title surface or a clipped
// destination draw. Direct mode avoids an additional GPU surface.
type RasterTitleMode uint8

const (
	RasterTitleCanvas RasterTitleMode = iota
	RasterTitleDirect
)

// RasterTitleConfig layers three moving raster copies behind one title image.
// RasterScale and TitleScale are independent of the output position; the third
// raster follows motion lane one with an additional vertical source offset.
type RasterTitleConfig struct {
	Title, Raster                 *ebiten.Image
	TitleSource, RasterSource     image.Rectangle
	Motion                        motion.WrapBankConfig
	Mode                          RasterTitleMode
	CanvasSize                    image.Point
	Clip                          image.Rectangle
	UnderlayWidth, UnderlayHeight float64
	RasterScaleX, RasterScaleY    float64
	TitleScaleX, TitleScaleY      float64
	ThirdOffsetY                  float64
	Fill                          color.Color
	Filter                        ebiten.Filter
	Blend                         ebiten.Blend
}

// RasterTitle borrows its artwork and owns only a small canvas in canvas mode.
// Motion is advanced in Update/Step, so repeated DrawAt calls keep one phase.
type RasterTitle struct {
	config        RasterTitleConfig
	motion        *motion.WrapBank
	title, raster *ebiten.Image
	canvas        *ebiten.Image
}

func NewRasterTitle(config RasterTitleConfig) (*RasterTitle, error) {
	if config.Mode != RasterTitleCanvas && config.Mode != RasterTitleDirect || config.Title == nil || config.Raster == nil {
		return nil, fmt.Errorf("composite: invalid raster title mode or images")
	}
	for _, value := range [...]float64{
		config.UnderlayWidth, config.UnderlayHeight, config.RasterScaleX, config.RasterScaleY,
		config.TitleScaleX, config.TitleScaleY, config.ThirdOffsetY,
	} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("composite: invalid raster title scale or offset")
		}
	}
	if config.UnderlayWidth < 0 || config.UnderlayHeight < 0 {
		return nil, fmt.Errorf("composite: negative raster title underlay size")
	}
	if config.RasterScaleX == 0 {
		config.RasterScaleX = 1
	}
	if config.RasterScaleY == 0 {
		config.RasterScaleY = 1
	}
	if config.TitleScaleX == 0 {
		config.TitleScaleX = 1
	}
	if config.TitleScaleY == 0 {
		config.TitleScaleY = 1
	}
	if config.Fill == nil {
		config.Fill = color.Black
	}
	if config.Mode == RasterTitleDirect && config.Clip.Empty() {
		return nil, fmt.Errorf("composite: direct raster title needs a clip rectangle")
	}
	view := func(source *ebiten.Image, crop image.Rectangle) (*ebiten.Image, error) {
		if crop.Empty() {
			return source, nil
		}
		clip := crop.Intersect(source.Bounds())
		if clip.Empty() {
			return nil, fmt.Errorf("composite: empty raster title source crop")
		}
		return source.SubImage(clip).(*ebiten.Image), nil
	}
	var err error
	title, err := view(config.Title, config.TitleSource)
	if err != nil {
		return nil, err
	}
	raster, err := view(config.Raster, config.RasterSource)
	if err != nil {
		return nil, err
	}
	bank, err := motion.NewWrapBank(config.Motion)
	if err != nil {
		return nil, err
	}
	if bank.Len() != 2 {
		return nil, fmt.Errorf("composite: raster title needs two motion lanes")
	}
	effect := &RasterTitle{config: config, motion: bank, title: title, raster: raster}
	if config.Mode == RasterTitleCanvas {
		if config.CanvasSize == (image.Point{}) {
			config.CanvasSize = title.Bounds().Size()
		}
		if config.CanvasSize.X <= 0 || config.CanvasSize.Y <= 0 {
			return nil, fmt.Errorf("composite: invalid raster title canvas")
		}
		effect.config.CanvasSize = config.CanvasSize
		effect.canvas = ebiten.NewImage(config.CanvasSize.X, config.CanvasSize.Y)
	}
	return effect, nil
}

func (effect *RasterTitle) Step()                    { effect.motion.Step() }
func (effect *RasterTitle) Update(kit.Frame) error   { effect.Step(); return nil }
func (effect *RasterTitle) Motion() *motion.WrapBank { return effect.motion }

func (effect *RasterTitle) DrawAt(dst *ebiten.Image, x, y float64) {
	if effect == nil || dst == nil {
		return
	}
	c := effect.config
	target := effect.canvas
	drawX, drawY := 0.0, 0.0
	if c.Mode == RasterTitleCanvas {
		target.Fill(c.Fill)
	} else {
		if c.UnderlayWidth > 0 && c.UnderlayHeight > 0 {
			vector.DrawFilledRect(dst, float32(x), float32(y), float32(c.UnderlayWidth), float32(c.UnderlayHeight), c.Fill, false)
		}
		clip := c.Clip.Intersect(dst.Bounds())
		if clip.Empty() {
			return
		}
		target = dst.SubImage(clip).(*ebiten.Image)
		drawX, drawY = x, y
	}
	for _, phase := range [...]float64{effect.motion.At(0), effect.motion.At(1), effect.motion.At(1) + c.ThirdOffsetY} {
		op := ebiten.DrawImageOptions{Filter: c.Filter, Blend: c.Blend}
		op.GeoM.Scale(c.RasterScaleX, c.RasterScaleY)
		op.GeoM.Translate(drawX, drawY+c.RasterScaleY*phase)
		target.DrawImage(effect.raster, &op)
	}
	titleOptions := ebiten.DrawImageOptions{Filter: c.Filter, Blend: c.Blend}
	titleOptions.GeoM.Scale(c.TitleScaleX, c.TitleScaleY)
	titleOptions.GeoM.Translate(drawX, drawY)
	target.DrawImage(effect.title, &titleOptions)
	if effect.canvas != nil {
		op := ebiten.DrawImageOptions{Filter: c.Filter, Blend: c.Blend}
		op.GeoM.Translate(x, y)
		dst.DrawImage(effect.canvas, &op)
	}
}

func (effect *RasterTitle) Close() error {
	if effect != nil && effect.canvas != nil {
		effect.canvas.Deallocate()
		effect.canvas = nil
	}
	return nil
}
