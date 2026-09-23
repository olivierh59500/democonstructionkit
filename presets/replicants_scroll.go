package presets

import (
	"math"

	"github.com/olivierh59500/democonstructionkit/scrolling"
)

// ReplicantsRowWave extends the TeamG1 three-section profile with a repeated
// opening and two final sections. It returns independent editable samples.
func ReplicantsRowWave() []float64 {
	values := TeamG1ScrollWave()
	values = append(values, values[:389]...)
	step := 72.0 / 180.0 * math.Pi
	for i := 0; i < 36; i++ {
		values = append(values, 4*math.Sin(float64(i)*step))
	}
	step = 8.0 / 180.0 * math.Pi
	for i := 0; i < 189; i++ {
		values = append(values, 30*math.Sin(float64(i)*step))
	}
	return values
}

// DMA3DRowColumn preserves the historical batched sampling and clipped text
// window while allowing any atlas and message to drive the same transport.
func DMA3DRowColumn(font *scrolling.Atlas, text string) scrolling.RowColumnConfig {
	return scrolling.RowColumnConfig{
		Font: font, Text: text, Advance: 64,
		ViewportWidth: 640, Height: 50, WorkWidth: 640 + 512,
		RowHeight: 2, ColumnWidth: 16,
		Wave: ReplicantsRowWave(), RowBias: 64, RowPhaseStep: 1,
		TextSpeed: 4, TextRestartX: 640, InitialMultiplier: 1,
		ColumnY: 380, ColumnBias: 35, ColumnAmplitude: 35,
		ColumnSpatial: .1, ColumnSpeed: .1,
		ColumnWrap: 2 * math.Pi,
		Cull:       scrolling.RowColumnBounds, DrawMode: scrolling.RowColumnQuads,
	}
}

// ReplicantsRowColumn uses source-image strips and a variable-speed clock.
func ReplicantsRowColumn(font *scrolling.Atlas, text string) scrolling.RowColumnConfig {
	c := DMA3DRowColumn(font, text)
	c.ColumnY = 280
	c.ColumnBiasInside = true
	c.ColumnWrap = 0
	c.InitialMultiplier = 1.4
	c.Cull = scrolling.RowColumnMapper
	c.DrawMode = scrolling.RowColumnImages
	return c
}
