package presets

import (
	"fmt"
	"math"

	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

// PhenomenaDNAProgram compiles the authored character order, two-pixel strips,
// nonzero loop start and four pause/rotation commands into editable DCK data.
// The message and glyph lookup remain supplied by the production.
func PhenomenaDNAProgram(message string, glyphIndex func(rune) (int, bool)) (scrolling.SliceProgramConfig, error) {
	if glyphIndex == nil || len(message) <= 90 {
		return scrolling.SliceProgramConfig{}, fmt.Errorf("presets: invalid Phenomena DNA message or font lookup")
	}
	tokens := make([]scrolling.SliceToken, 0, len(message))
	for _, ch := range message {
		if ch == '^' || ch == '#' || ch == '&' || ch == '%' {
			tokens = append(tokens, scrolling.SliceToken{Control: string(ch)})
			continue
		}
		index, ok := glyphIndex(ch)
		if !ok {
			index = 0
		}
		tokens = append(tokens, scrolling.SliceToken{Glyph: index, Width: 16})
	}
	if len(tokens) <= 90 {
		return scrolling.SliceProgramConfig{}, fmt.Errorf("presets: DNA loop start exceeds token count")
	}
	offsets := make([]float64, 240)
	for i := range offsets {
		offsets[i] = math.Sin(float64(i)*0.05) * 15
	}
	return scrolling.SliceProgramConfig{
		Stream: scrolling.SliceStreamConfig{
			Tokens: tokens, Capacity: len(offsets), SliceWidth: 2, Repeat: true, LoopStart: 90,
		},
		Clock: motion.CuedScrollClockConfig{
			InitialTextStep: 1, InitialPauseTicks: 250, InitialRotationStep: .35,
			ResumeTextStep: 1, ResumeRotationStep: .35, RotationFrames: 30,
			Cues: map[string]motion.ScrollCue{
				"^": {PauseTicks: 275, SetRotation: true, RotationStep: -1},
				"&": {PauseTicks: 275, SetRotation: true, RotationStep: 1},
				"#": {PauseTicks: 250, SetRotation: true, RotationStep: -1},
				"%": {PauseTicks: 225, SetRotation: true, RotationStep: -1},
			},
		},
		Offsets: offsets,
	}, nil
}
