package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// DOMBackgroundStrips reproduces the vertically sampled eleven-band scenery.
// The source image is borrowed; all sampling and timing values are editable.
func DOMBackgroundStrips(image *ebiten.Image) composite.VerticalStripTrainConfig {
	return composite.VerticalStripTrainConfig{
		Image: image, SourceStride: 2, StripHeight: 36,
		SampleStep: 4, Count: 11, OutputY: 62, RowStep: 36,
		Phase: 0, Velocity: 2,
		Wrap: &composite.RasterWrap{Boundary: 654, Restart: 0, Inclusive: true},
	}
}
