package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/effects"
)

// DOCCheckerboard uses the separate XOR mask and updates the oscillator before
// moving the floor. The returned configuration remains fully editable.
func DOCCheckerboard() effects.PerspectiveCheckerboardConfig {
	return effects.DefaultPerspectiveCheckerboardConfig()
}

// CuddlyDOCCheckerboard uses direct per-band XOR and advances the oscillator
// after moving the floor, retaining that screen's authored phase and opacity.
func CuddlyDOCCheckerboard() effects.PerspectiveCheckerboardConfig {
	c := effects.DefaultPerspectiveCheckerboardConfig()
	c.Composition = effects.CheckerboardDirectXOR
	c.ClockOrder = effects.CheckerboardAfterMove
	c.Wrap = effects.CheckerboardSingleWrap
	c.AntiAlias = true
	c.Unmanaged = true
	c.Filter = ebiten.FilterLinear
	c.Opacity = .3
	return c
}
