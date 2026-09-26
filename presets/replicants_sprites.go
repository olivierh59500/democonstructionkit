package presets

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// ReplicantsBouncingSprites gives two images fixed horizontal spacing and
// quarter-cycle-separated rectified vertical waves. Image, amplitude, phase,
// spacing and baseline remain editable on the returned TrainConfig.
func ReplicantsBouncingSprites(image *ebiten.Image) sprites.TrainConfig {
	return sprites.TrainConfig{
		Images: []*ebiten.Image{image, image}, Count: 2,
		X: sprites.TrainAxis{Offset: 32},
		Y: sprites.TrainAxis{Offset: 326, Wave: &motion.Wave{
			Amplitude: -24, Spatial: -math.Pi / 2, Speed: 1,
			Phase: math.Pi / 2, Rectify: true,
		}},
		Spacing: motion.Point{X: 480},
	}
}
