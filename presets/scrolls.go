package presets

import "github.com/olivierh59500/democonstructionkit/scrolling"

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
// both display the preceding byte. New texts can use ordinary Scrolling controls
// without these historical timing slots.
func TCBPlaneSlots(text string, advance float64) []scrolling.PlaneSlot {
	slots := make([]scrolling.PlaneSlot, len(text))
	for i := range text {
		letter, form := text[i], -1
		if letter == '^' && i+1 < len(text) && text[i+1] >= '0' && text[i+1] <= '7' {
			form = int(text[i+1] - '0')
			letter = text[(i-1+len(text))%len(text)]
		} else if i >= 2 && text[i-1] == '^' && letter >= '0' && letter <= '7' {
			letter = text[i-2]
		}
		slots[i] = scrolling.PlaneSlot{Rune: rune(letter), Advance: advance, Form: form}
	}
	return slots
}
