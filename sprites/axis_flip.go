package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// AxisFlipConfig alternates image faces while a signed vertical scale crosses
// SwitchAt. Back may be nil to flip one image. Motion must contain one bounce
// lane; its bounds and velocity define the turn points and cycle speed.
type AxisFlipConfig struct {
	Front, Back           *ebiten.Image
	Motion                motion.BounceBankConfig
	SwitchAt              float64
	FrontAngle, BackAngle float64 // Degrees, matching image transform conventions.
	ScaleX                float64 // Zero defaults to one.
	Filter                ebiten.Filter
	Blend                 ebiten.Blend
}

// AxisFlipPose can be drawn with the built-in centered renderer or combined
// with another path, shader, mask or layer renderer.
type AxisFlipPose struct {
	Image                 *ebiten.Image
	ScaleX, ScaleY, Angle float64
}

// AxisFlip owns one reversible face-selection clock and borrows its images.
// Pose and DrawAt never advance motion; call Step once per simulation update.
type AxisFlip struct {
	config AxisFlipConfig
	motion *motion.BounceBank
}

func NewAxisFlip(config AxisFlipConfig) (*AxisFlip, error) {
	if config.Front == nil {
		return nil, fmt.Errorf("sprites: axis flip needs a front image")
	}
	for _, value := range [...]float64{config.SwitchAt, config.FrontAngle, config.BackAngle, config.ScaleX} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("sprites: nonfinite axis flip setting")
		}
	}
	if config.Back == nil {
		config.Back = config.Front
	}
	if config.ScaleX == 0 {
		config.ScaleX = 1
	}
	bank, err := motion.NewBounceBank(config.Motion)
	if err != nil {
		return nil, fmt.Errorf("sprites: axis flip motion: %w", err)
	}
	if bank.Len() != 1 {
		return nil, fmt.Errorf("sprites: axis flip needs one motion lane")
	}
	return &AxisFlip{config: config, motion: bank}, nil
}

// Pose returns the current face, signed scale and angle before the next Step.
func (flip *AxisFlip) Pose() AxisFlipPose {
	scaleY := flip.motion.At(0)
	pose := AxisFlipPose{Image: flip.config.Front, ScaleX: flip.config.ScaleX, ScaleY: scaleY, Angle: flip.config.FrontAngle}
	if scaleY <= flip.config.SwitchAt {
		pose.Image = flip.config.Back
		pose.Angle = flip.config.BackAngle
	}
	return pose
}

func (flip *AxisFlip) Step()  { flip.motion.Step() }
func (flip *AxisFlip) Reset() { flip.motion.Reset() }

// DrawAt centers the selected image at x,y. Position may come from any DCK
// motion path or from the caller; drawing leaves the flip clock unchanged.
func (flip *AxisFlip) DrawAt(dst *ebiten.Image, x, y float64) {
	pose := flip.Pose()
	op := ebiten.DrawImageOptions{Filter: flip.config.Filter, Blend: flip.config.Blend}
	op.GeoM.Translate(-float64(pose.Image.Bounds().Dx())/2, -float64(pose.Image.Bounds().Dy())/2)
	op.GeoM.Scale(pose.ScaleX, pose.ScaleY)
	op.GeoM.Rotate(pose.Angle * math.Pi / 180)
	op.GeoM.Translate(x, y)
	dst.DrawImage(pose.Image, &op)
}
