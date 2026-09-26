package composite

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/render"
)

// WindowedImageBankConfig draws one borrowed image into ordered, independently
// clipped windows. Each optional clock has one value; its phase is added to
// every window on that axis. Zero scale defaults to one.
type WindowedImageBankConfig struct {
	Image          *ebiten.Image
	Windows        []WindowedImageWindow
	MotionX        *motion.WrapBankConfig
	MotionY        *motion.WrapBankConfig
	ScaleX, ScaleY float64
	Filter         ebiten.Filter
	Blend          ebiten.Blend
}

// WindowedImageBank borrows its source and destination. It owns one bounded
// surface per window, preserving the same two-pass blend/filter behavior as
// image strips rendered to transparent surfaces before stage composition.
// Draw samples the current phases without advancing them.
type WindowedImageBank struct {
	image    *ebiten.Image
	windows  []WindowedImageWindow
	surfaces []*ebiten.Image
	x, y     *motion.WrapBank
	scaleX   float64
	scaleY   float64
	filter   ebiten.Filter
	blend    ebiten.Blend
}

func NewWindowedImageBank(config WindowedImageBankConfig) (*WindowedImageBank, error) {
	if config.Image == nil || len(config.Windows) == 0 || len(config.Windows) > 4096 {
		return nil, fmt.Errorf("composite: invalid windowed image source or window count")
	}
	if config.ScaleX == 0 {
		config.ScaleX = 1
	}
	if config.ScaleY == 0 {
		config.ScaleY = 1
	}
	for _, scale := range [...]float64{config.ScaleX, config.ScaleY} {
		if math.IsNaN(scale) || math.IsInf(scale, 0) || scale <= 0 {
			return nil, fmt.Errorf("composite: invalid windowed image scale")
		}
	}
	for _, window := range config.Windows {
		if window.Clip.Empty() || math.IsNaN(window.Offset.X) || math.IsNaN(window.Offset.Y) ||
			math.IsInf(window.Offset.X, 0) || math.IsInf(window.Offset.Y, 0) {
			return nil, fmt.Errorf("composite: invalid image window")
		}
	}
	bank := &WindowedImageBank{
		image: config.Image, windows: append([]WindowedImageWindow(nil), config.Windows...),
		scaleX: config.ScaleX, scaleY: config.ScaleY, filter: config.Filter, blend: config.Blend,
	}
	for _, axis := range []struct {
		config *motion.WrapBankConfig
		clock  **motion.WrapBank
	}{{config.MotionX, &bank.x}, {config.MotionY, &bank.y}} {
		if axis.config == nil {
			continue
		}
		clock, err := motion.NewWrapBank(*axis.config)
		if err != nil {
			return nil, err
		}
		if clock.Len() != 1 {
			return nil, fmt.Errorf("composite: windowed image clock must contain one phase")
		}
		*axis.clock = clock
	}
	for _, window := range bank.windows {
		bank.surfaces = append(bank.surfaces, render.NewSurface(window.Clip.Dx(), window.Clip.Dy()))
	}
	return bank, nil
}

// XClock and YClock expose the owned phases for live speed changes.
func (bank *WindowedImageBank) XClock() *motion.WrapBank { return bank.x }
func (bank *WindowedImageBank) YClock() *motion.WrapBank { return bank.y }

func (bank *WindowedImageBank) Step() {
	if bank.x != nil {
		bank.x.Step()
	}
	if bank.y != nil {
		bank.y.Step()
	}
}

func (bank *WindowedImageBank) Reset() {
	if bank.x != nil {
		bank.x.Reset()
	}
	if bank.y != nil {
		bank.y.Reset()
	}
}

// Draw clears and refills each bounded window in order before compositing it
// onto the destination. This preserves intermediate alpha and clipping.
func (bank *WindowedImageBank) Draw(dst *ebiten.Image) {
	if bank == nil || dst == nil || len(bank.surfaces) != len(bank.windows) {
		return
	}
	phaseX, phaseY := 0.0, 0.0
	if bank.x != nil {
		phaseX = bank.x.At(0)
	}
	if bank.y != nil {
		phaseY = bank.y.At(0)
	}
	for i, window := range bank.windows {
		surface := bank.surfaces[i]
		surface.Clear()
		origin := windowedImageLocalOrigin(window, phaseX, phaseY)
		op := ebiten.DrawImageOptions{Filter: bank.filter, Blend: bank.blend}
		op.GeoM.Scale(bank.scaleX, bank.scaleY)
		op.GeoM.Translate(origin.X, origin.Y)
		surface.DrawImage(bank.image, &op)
		output := ebiten.DrawImageOptions{Filter: bank.filter, Blend: bank.blend}
		output.GeoM.Translate(float64(window.Clip.Min.X), float64(window.Clip.Min.Y))
		dst.DrawImage(surface, &output)
	}
}

// Close releases the owned window surfaces; borrowed images remain alive.
func (bank *WindowedImageBank) Close() error {
	if bank == nil {
		return nil
	}
	for _, surface := range bank.surfaces {
		surface.Deallocate()
	}
	bank.surfaces = nil
	return nil
}
