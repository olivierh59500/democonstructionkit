package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// AxisFlipConfig alternates image faces using either one bouncing scale lane or
// a saw cycle that toggles faces on each strict wrap. SwitchAt selects the back
// face in bounce mode. Back may be nil to flip one image.
type AxisFlipConfig struct {
	Front, Back           *ebiten.Image
	Motion                motion.BounceBankConfig
	Saw                   *motion.SawToggleConfig
	SwitchAt              float64
	FrontAngle, BackAngle float64 // Degrees, matching image transform conventions.
	ScaleX                float64 // Zero defaults to one.
	AnchorX, AnchorY      float64 // Pixel anchor; UseAnchor=false centers each selected face.
	UseAnchor             bool
	BackMirrorY           bool    // Mirror the back face before anchoring and scaling.
	BackMirrorShift       float64 // Zero uses the selected image height.
	Filter                ebiten.Filter
	Blend                 ebiten.Blend
}

// AxisFlipPose can be drawn with the built-in centered renderer or combined
// with another path, shader, mask or layer renderer.
type AxisFlipPose struct {
	Image                 *ebiten.Image
	ScaleX, ScaleY, Angle float64
	Back                  bool
}

// AxisFlip owns one reversible face-selection clock and borrows its images.
// Pose and DrawAt never advance motion; call Step once per simulation update.
type AxisFlip struct {
	config AxisFlipConfig
	motion *motion.BounceBank
	saw    *motion.SawToggle
}

func NewAxisFlip(config AxisFlipConfig) (*AxisFlip, error) {
	if config.Front == nil {
		return nil, fmt.Errorf("sprites: axis flip needs a front image")
	}
	for _, value := range [...]float64{config.SwitchAt, config.FrontAngle, config.BackAngle, config.ScaleX, config.AnchorX, config.AnchorY, config.BackMirrorShift} {
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
	if config.Saw != nil {
		if len(config.Motion.Start) != 0 || len(config.Motion.Velocity) != 0 {
			return nil, fmt.Errorf("sprites: choose bounce motion or saw cycle")
		}
		cycle, err := motion.NewSawToggle(*config.Saw)
		if err != nil {
			return nil, fmt.Errorf("sprites: axis flip saw: %w", err)
		}
		return &AxisFlip{config: config, saw: cycle}, nil
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
	scaleY, back := 0.0, false
	if flip.saw != nil {
		scaleY, back = flip.saw.At(), flip.saw.Alternate()
	} else {
		scaleY = flip.motion.At(0)
		back = scaleY <= flip.config.SwitchAt
	}
	pose := AxisFlipPose{Image: flip.config.Front, ScaleX: flip.config.ScaleX, ScaleY: scaleY, Angle: flip.config.FrontAngle, Back: back}
	if back {
		pose.Image = flip.config.Back
		pose.Angle = flip.config.BackAngle
	}
	return pose
}

func (flip *AxisFlip) Step() {
	if flip.saw != nil {
		flip.saw.Step()
	} else {
		flip.motion.Step()
	}
}
func (flip *AxisFlip) Reset() {
	if flip.saw != nil {
		flip.saw.Reset()
	} else {
		flip.motion.Reset()
	}
}

// DrawAt places the selected image at x,y. Without UseAnchor, each face is
// centered independently. Drawing leaves the flip clock unchanged.
func (flip *AxisFlip) DrawAt(dst *ebiten.Image, x, y float64) {
	pose := flip.Pose()
	op := flip.optionsForPose(pose, x, y, nil)
	dst.DrawImage(pose.Image, &op)
}

// DrawAtWith applies parent after the local image pose. It can place a native
// flip into a scaled stage or another composed coordinate system.
func (flip *AxisFlip) DrawAtWith(dst *ebiten.Image, x, y float64, parent ebiten.GeoM) {
	pose := flip.Pose()
	op := flip.optionsForPose(pose, x, y, &parent)
	dst.DrawImage(pose.Image, &op)
}

// OptionsAt exposes the same transform used by DrawAt and DrawAtWith, so a
// caller can submit the face through another renderer or change its material.
func (flip *AxisFlip) OptionsAt(x, y float64, parent *ebiten.GeoM) ebiten.DrawImageOptions {
	return flip.optionsForPose(flip.Pose(), x, y, parent)
}

func (flip *AxisFlip) optionsForPose(pose AxisFlipPose, x, y float64, parent *ebiten.GeoM) ebiten.DrawImageOptions {
	op := ebiten.DrawImageOptions{Filter: flip.config.Filter, Blend: flip.config.Blend}
	if pose.Back && flip.config.BackMirrorY {
		shift := flip.config.BackMirrorShift
		if shift == 0 {
			shift = float64(pose.Image.Bounds().Dy())
		}
		op.GeoM.Scale(1, -1)
		op.GeoM.Translate(0, shift)
	}
	anchorX, anchorY := float64(pose.Image.Bounds().Dx())/2, float64(pose.Image.Bounds().Dy())/2
	if flip.config.UseAnchor {
		anchorX, anchorY = flip.config.AnchorX, flip.config.AnchorY
	}
	op.GeoM.Translate(-anchorX, -anchorY)
	op.GeoM.Scale(pose.ScaleX, pose.ScaleY)
	op.GeoM.Rotate(pose.Angle * math.Pi / 180)
	op.GeoM.Translate(x, y)
	if parent != nil {
		op.GeoM.Concat(*parent)
	}
	return op
}
