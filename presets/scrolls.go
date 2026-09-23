package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

// TCBPlanes returns the authored multi-plane transport recipe. Both its forms
// and projection can be edited before construction. Advance belongs to the
// chosen font; text and images are supplied by the caller.
func TCBPlanes(text string, advance float64) scrolling.PlanesConfig {
	return scrolling.PlanesConfig{
		Slots: TCBPlaneSlots(text, advance), Forms: TCBScrollForms(),
		Visible: 30, PhaseStep: .02,
		Projection: scrolling.PlaneProjection{Focal: 250, Depth: 150, OriginX: -450, CenterX: 160, CenterY: 100, XBias: -16, YBias: -14, VerticalOffset: -4},
	}
}

// TCBProjectedScroll is a ready-to-customize scrolling.New configuration. Draw
// placement defaults to native 320x200 coordinates; callers can set its origin,
// scale, speed, projection, forms or ordered output passes independently.
func TCBProjectedScroll(text string, advance float64, face scrolling.Face, raster *ebiten.Image) scrolling.Config {
	return scrolling.Config{Projected: &scrolling.ProjectedConfig{
		Planes: TCBPlanes(text, advance), Face: face, Raster: raster,
		PixelsPerUpdate: 4, Draw: scrolling.PlaneDraw{ScaleX: 1, ScaleY: 1},
	}}
}

// TCBScrollForms returns independent normal, bounce, sine, static zoom, animated
// zoom and three perspective forms. Speeds match the original .02 phase/tick;
// use Speed*1.2 with seconds at 60 Hz in a regular Scrolling mode.
func TCBScrollForms() []scrolling.PlaneForm {
	return []scrolling.PlaneForm{
		{Height: 55, VerticalPhase: 1.5},
		{Height: 55, VerticalSpeed: 2, VerticalPhase: 1.5},
		{Height: 55, VerticalStep: .2, VerticalSpeed: 2, VerticalPhase: 1.5},
		{DepthAmplitude: 200, DepthPhase: 5, Height: 55, VerticalStep: .2, VerticalSpeed: 2, VerticalPhase: 1.5},
		{DepthAmplitude: 200, DepthSpeed: 4, DepthPhase: 5, Height: 55, VerticalStep: .2, VerticalSpeed: 2, VerticalPhase: 1.5},
		{DepthAmplitude: 200, DepthStep: -.3, DepthSpeed: 4, Height: 55, VerticalStep: .3, VerticalSpeed: 2, VerticalPhase: 1.5},
		{DepthAmplitude: 200, DepthStep: .4, DepthSpeed: -4, DepthPhase: 5, Height: -70, VerticalStep: .4, VerticalSpeed: -4, VerticalPhase: 1.5},
		{DepthAmplitude: 150, DepthStep: .2, DepthSpeed: -3, DepthPhase: 5, Height: 55, VerticalStep: .2, VerticalSpeed: 2, VerticalPhase: 1.5},
	}
}

// TCBPlaneSlots retains the production's two visible slots for each ^N command:
// both display the preceding glyph. New texts can use ordinary Scrolling controls
// without these historical timing slots.
func TCBPlaneSlots(text string, advance float64) []scrolling.PlaneSlot {
	runes := []rune(text)
	slots := make([]scrolling.PlaneSlot, len(runes))
	for i := range runes {
		letter, form := runes[i], -1
		if letter == '^' && i+1 < len(runes) && runes[i+1] >= '0' && runes[i+1] <= '7' {
			form = int(runes[i+1] - '0')
			letter = runes[(i-1+len(runes))%len(runes)]
		} else if i >= 2 && runes[i-1] == '^' && letter >= '0' && letter <= '7' {
			letter = runes[i-2]
		}
		slots[i] = scrolling.PlaneSlot{Rune: letter, Advance: advance, Form: form}
	}
	return slots
}
