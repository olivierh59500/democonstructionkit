package presets

import (
	"image/color"

	"github.com/olivierh59500/democonstructionkit/palette"
)

// PhenomenaRasterStops is the editable purple-white-blue frame raster.
func PhenomenaRasterStops() []palette.GradientStop {
	return []palette.GradientStop{
		{Color: color.RGBA{R: 0x44, B: 0x44, A: 0xff}, Offset: 0},
		{Color: color.RGBA{R: 0xff, G: 0xdd, B: 0xff, A: 0xff}, Offset: .5},
		{Color: color.RGBA{R: 0x11, G: 0x11, B: 0x44, A: 0xff}, Offset: 1},
	}
}

// PhenomenaCoreStops colors the narrow red center of each DNA glyph.
func PhenomenaCoreStops() []palette.GradientStop {
	return []palette.GradientStop{
		{Color: color.RGBA{A: 0xff}, Offset: 0},
		{Color: color.RGBA{R: 0xff, G: 0x33, A: 0xff}, Offset: .5},
		{Color: color.RGBA{A: 0xff}, Offset: 1},
	}
}

// PhenomenaFrontStops colors the silver face of each rotating glyph.
func PhenomenaFrontStops() []palette.GradientStop {
	return []palette.GradientStop{
		{Color: color.RGBA{R: 0x55, G: 0x55, B: 0x55, A: 0xff}, Offset: 0},
		{Color: color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, Offset: .5},
		{Color: color.RGBA{R: 0x55, G: 0x55, B: 0x55, A: 0xff}, Offset: 1},
	}
}

// PhenomenaBackStops colors the purple face of each rotating glyph.
func PhenomenaBackStops() []palette.GradientStop {
	return []palette.GradientStop{
		{Color: color.RGBA{R: 0x34, G: 0x22, B: 0x55, A: 0xff}, Offset: 0},
		{Color: color.RGBA{R: 0x60, G: 0x4e, B: 0x98, A: 0xff}, Offset: .5},
		{Color: color.RGBA{R: 0x34, G: 0x22, B: 0x55, A: 0xff}, Offset: 1},
	}
}
