package presets

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

// CuddlyDNAFeedbackScroll binds one independently fonted text lane to its
// wrapped DNA history. Tick is the authored scene counter, including its intro.
// Edit the returned dimensions, source transport, gradient or phase sampler to
// reuse the same composition with different artwork and placement.
func CuddlyDNAFeedbackScroll(font scrolling.BitmapGrid, gradient *ebiten.Image,
	message string, profile []float64, direction int) (scrolling.Config, error) {
	if direction != 1 && direction != -1 || len(profile) == 0 {
		return scrolling.Config{}, fmt.Errorf("presets: invalid feedback direction or profile")
	}
	rows := make([]int, len(profile))
	for i, value := range profile {
		if math.IsNaN(value) || math.IsInf(value, 0) || math.Abs(value) >= 1<<30 {
			return scrolling.Config{}, fmt.Errorf("presets: invalid feedback row sample")
		}
		rows[i] = int(value)
	}
	insertY := 25.0
	if direction < 0 {
		insertY = 37
	}
	return scrolling.Config{
		Recycled: &scrolling.RecycledConfig{Ring: scrolling.RingConfig{
			Text: message, Font: font, Viewport: 320, Speed: 4, Controls: true,
		}},
		Output: &scrolling.OutputConfig{Width: 320, Height: 25,
			Feedback: []scrolling.FeedbackLayer{{
				Config: scrolling.FeedbackDNAConfig{Width: 320, Height: 64,
					HorizontalSpeed: 4, VerticalSpeed: 2, ColumnWidth: 2,
					Direction: direction, InsertY: insertY, Profile: rows,
					Filter: ebiten.FilterLinear},
				Gradient: gradient, X: 52, Y: 175,
				PhaseAt: func(frame kit.Frame) int { return direction * (int(frame.Tick) / 2) },
			}},
		},
	}, nil
}

// CuddlySpreadpointFeedbackScroll shares the same text/history engine while
// retaining Spreadpoint's insertion height, fixed phase and lower placement.
func CuddlySpreadpointFeedbackScroll(font scrolling.BitmapGrid, gradient *ebiten.Image,
	message string, profile []float64) (scrolling.Config, error) {
	config, err := CuddlyDNAFeedbackScroll(font, gradient, message, profile, 1)
	if err != nil {
		return scrolling.Config{}, err
	}
	layer := &config.Output.Feedback[0]
	layer.Config.InsertY, layer.Y, layer.PhaseAt = 4, 150, nil
	return config, nil
}
