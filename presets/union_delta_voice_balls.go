package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/modulation"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// UnionDeltaVoiceBalls keeps the authored three-voice crop, palette, decay and
// placement while using a reusable signal-driven sprite bank. The atlas is
// borrowed; change channels, positions or envelopes before constructing it.
func UnionDeltaVoiceBalls(image *ebiten.Image) (sprites.SignalFrameBankConfig, error) {
	atlas, err := sprites.NewAtlas(sprites.AtlasConfig{
		Image: image, TileW: 96, TileH: 114, Columns: 16, Count: 16,
	})
	if err != nil {
		return sprites.SignalFrameBankConfig{}, err
	}
	return sprites.SignalFrameBankConfig{
		Atlas:    atlas,
		Envelope: modulation.DecayConfig{Peak: 7, Rate: 1},
		Slots: []sprites.SignalFrameSlot{
			{Channel: 0, X: 244, Y: 177},
			{Channel: 1, X: 355, Y: 177},
			{Channel: 2, X: 464, Y: 177},
		},
	}, nil
}
