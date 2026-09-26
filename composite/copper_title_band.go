package composite

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// CopperTitleMode selects the final composition path. Surface retains an
// isolated, bounded band; Direct paints into the destination at its origin.
type CopperTitleMode uint8

const (
	CopperTitleSurface CopperTitleMode = iota
	CopperTitleDirect
)

// CopperTitleBandConfig combines a borrowed title with optional copper bars
// and a horizontal motion clock. Copper configuration controls material,
// geometry and wrap; TitleMotion controls phase, wave and initial hold.
type CopperTitleBandConfig struct {
	Title                            *ebiten.Image
	Copper                           *CopperBarsConfig
	TitleMotion                      motion.WaveClockConfig
	Width, Height                    int
	Mode                             CopperTitleMode
	Background                       color.Color
	TitleY, TitleScaleX, TitleScaleY float64
	TitleFilter                      ebiten.Filter
	TitleBlend                       ebiten.Blend
}

// CopperTitleBand owns both animation clocks and the optional band surface.
// The source images remain caller-owned. Draw never advances either clock.
type CopperTitleBand struct {
	config   CopperTitleBandConfig
	clock    *motion.WaveClock
	copper   *CopperBars
	layer    *SurfaceLayer
	baseStep float64
}

func NewCopperTitleBand(config CopperTitleBandConfig) (*CopperTitleBand, error) {
	if config.Title == nil || config.Width < 1 || config.Height < 1 ||
		config.Width > 8192 || config.Height > 8192 || config.Mode > CopperTitleDirect {
		return nil, fmt.Errorf("composite: invalid copper title size, image or mode")
	}
	if config.TitleScaleX == 0 {
		config.TitleScaleX = 1
	}
	if config.TitleScaleY == 0 {
		config.TitleScaleY = float64(config.Height) / float64(config.Title.Bounds().Dy())
	}
	for _, value := range [...]float64{config.TitleY, config.TitleScaleX, config.TitleScaleY} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("composite: nonfinite copper title transform")
		}
	}
	if config.TitleScaleX <= 0 || config.TitleScaleY <= 0 {
		return nil, fmt.Errorf("composite: invalid copper title scale")
	}
	if config.Background == nil {
		config.Background = color.Black
	}
	clock, err := motion.NewWaveClock(config.TitleMotion)
	if err != nil {
		return nil, err
	}
	band := &CopperTitleBand{config: config, clock: clock, baseStep: config.TitleMotion.Step}
	if config.Copper != nil {
		band.copper, err = NewCopperBars(*config.Copper)
		if err != nil {
			return nil, err
		}
	}
	if config.Mode == CopperTitleSurface {
		var sources []kit.Effect
		if band.copper != nil {
			sources = []kit.Effect{band.copper}
		}
		band.layer, err = NewSurfaceLayer(SurfaceLayerConfig{
			Width: config.Width, Height: config.Height, Background: config.Background,
			Sources: sources,
			Passes: []SurfaceImagePass{{Image: config.Title, X: clock.At(0), Y: config.TitleY,
				ScaleX: config.TitleScaleX, ScaleY: config.TitleScaleY,
				Filter: config.TitleFilter, Blend: config.TitleBlend}},
			Outputs: []SurfaceOutput{{}},
		})
		if err != nil {
			return nil, err
		}
	}
	return band, nil
}

// Advance applies one caller-selected speed to both clocks. Zero freezes the
// effect; a later speed change retains their current phases.
func (band *CopperTitleBand) Advance(speed float64) error {
	if band == nil || math.IsNaN(speed) || math.IsInf(speed, 0) || speed < 0 {
		return fmt.Errorf("composite: invalid copper title speed")
	}
	if band.config.Mode == CopperTitleSurface && band.layer == nil {
		return fmt.Errorf("composite: closed copper title surface")
	}
	if band.copper != nil {
		if err := band.copper.SetSpeed(speed); err != nil {
			return err
		}
		if band.layer != nil {
			if err := band.layer.Update(kit.Frame{}); err != nil {
				return err
			}
		} else if err := band.copper.Update(kit.Frame{}); err != nil {
			return err
		}
	}
	if err := band.clock.SetStep(band.baseStep * speed); err != nil {
		return err
	}
	band.clock.Step()
	if band.layer != nil {
		return band.layer.SetPassPosition(0, band.clock.At(0), band.config.TitleY)
	}
	return nil
}

func (band *CopperTitleBand) Draw(dst *ebiten.Image) {
	if band == nil || dst == nil {
		return
	}
	if band.config.Mode == CopperTitleSurface {
		if band.layer != nil {
			band.layer.Draw(dst)
		}
		return
	}
	c := band.config
	vector.DrawFilledRect(dst, 0, 0, float32(c.Width), float32(c.Height), c.Background, false)
	if band.copper != nil {
		band.copper.Draw(dst)
	}
	var op ebiten.DrawImageOptions
	op.Filter, op.Blend = c.TitleFilter, c.TitleBlend
	op.GeoM.Scale(c.TitleScaleX, c.TitleScaleY)
	op.GeoM.Translate(band.clock.At(0), c.TitleY)
	dst.DrawImage(c.Title, &op)
}

func (band *CopperTitleBand) Copper() *CopperBars      { return band.copper }
func (band *CopperTitleBand) Clock() *motion.WaveClock { return band.clock }
func (band *CopperTitleBand) Close() error {
	if band == nil || band.layer == nil {
		return nil
	}
	err := band.layer.Close()
	band.layer = nil
	return err
}
