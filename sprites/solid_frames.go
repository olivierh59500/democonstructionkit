package sprites

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// SolidFrame describes one colored rectangle used as a sprite material.
type SolidFrame struct {
	Width, Height int
	Color         color.Color
}

// NewSolidFrames creates caller-owned cached images for stars, bars or sprite
// fields. Construct once and deallocate the returned images when the scene ends.
func NewSolidFrames(materials []SolidFrame) ([]*ebiten.Image, error) {
	if len(materials) == 0 || len(materials) > 1<<16 {
		return nil, fmt.Errorf("sprites: invalid solid frame bank")
	}
	frames := make([]*ebiten.Image, 0, len(materials))
	for _, material := range materials {
		if material.Width < 1 || material.Height < 1 || material.Width > 8192 || material.Height > 8192 || material.Color == nil {
			for _, frame := range frames {
				frame.Deallocate()
			}
			return nil, fmt.Errorf("sprites: invalid solid frame material")
		}
		frame := ebiten.NewImage(material.Width, material.Height)
		frame.Fill(material.Color)
		frames = append(frames, frame)
	}
	return frames, nil
}
