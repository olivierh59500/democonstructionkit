package composite

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/palette"
)

// NewGradientImage uploads one configurable gradient at construction time.
// The caller owns and deallocates the returned GPU image.
func NewGradientImage(c palette.GradientConfig) (*ebiten.Image, error) {
	img, err := palette.NewGradient(c)
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

// NewUniformGradientImage uploads one evenly spaced color bank material.
func NewUniformGradientImage(c palette.UniformGradientConfig) (*ebiten.Image, error) {
	img, err := palette.NewUniformGradient(c)
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

// NewWhiteSilhouette uploads a source's alpha as a white tintable material.
func NewWhiteSilhouette(source image.Image) (*ebiten.Image, error) {
	img, err := palette.WhiteSilhouette(source)
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

// NewInvertedImage uploads unpremultiplied inverted RGB with unchanged alpha.
func NewInvertedImage(source image.Image) (*ebiten.Image, error) {
	img, err := palette.InvertRGB(source)
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}
