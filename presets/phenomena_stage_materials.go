package presets

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/timeline"
)

// PhenomenaStageImages are borrowed; page images may come from BitmapPage.
type PhenomenaStageImages struct {
	RasterBar, Page1, Page2, Logo, LogoMask *ebiten.Image
	Middle, Raster, Photon, PhotonMask      *ebiten.Image
}

// PhenomenaStageMaterials describes the exact ordered image passes of the
// intro and outro. The main scroller, its live photon material and input remain
// caller-owned; another demo can replace every borrowed image or tint formula.
func PhenomenaStageMaterials(director *timeline.ScalarStages, images PhenomenaStageImages) composite.ScalarStagePainterConfig {
	formulas := PhenomenaStageColorFormulas()
	value, percent := formulas.Value, formulas.Percent
	brightDown, whiteMask := formulas.BrightDown, formulas.WhiteMask
	red, green, secondary := formulas.Percent, formulas.Green, formulas.Secondary
	stage := func(index int) composite.ScalarStageRule {
		return composite.ScalarStageRule{From: index, To: index}
	}
	stages := func(first, last int) composite.ScalarStageRule {
		return composite.ScalarStageRule{From: first, To: last}
	}
	lessOrEqual := func(index int) composite.ScalarStageRule {
		return composite.ScalarStageRule{From: index, To: index,
			Compare: timeline.ScalarLessEqual, Threshold: 100}
	}
	greater := func(index int) composite.ScalarStageRule {
		return composite.ScalarStageRule{From: index, To: index,
			Compare: timeline.ScalarGreater, Threshold: 100}
	}
	brightness := func(expression *motion.FormulaExpr) composite.ScalarTintConfig {
		return composite.ScalarTintConfig{Mode: composite.ScalarTintRGB,
			R: expression, G: expression, B: expression}
	}
	alpha := func(expression *motion.FormulaExpr) composite.ScalarTintConfig {
		return composite.ScalarTintConfig{Mode: composite.ScalarTintAlpha, A: expression}
	}
	backgrounds := make([]color.Color, PhenomenaEnd+1)
	for i := range backgrounds {
		backgrounds[i] = color.Black
	}
	backgrounds[PhenomenaShowLogo] = color.RGBA{R: 0, G: 1, B: 17, A: 255}
	return composite.ScalarStagePainterConfig{
		Director: director, Backgrounds: backgrounds,
		Passes: []composite.ScalarImagePass{
			{Rule: stage(PhenomenaTextPage1), Image: images.RasterBar, YFormula: &value},
			{Rule: stage(PhenomenaTextPage1), Image: images.Page1},
			{Rule: lessOrEqual(PhenomenaTextPage2), Image: images.Page2, Tint: brightness(&percent)},
			{Rule: greater(PhenomenaTextPage2), Image: images.Page2, Tint: brightness(&brightDown)},
			{Rule: lessOrEqual(PhenomenaShowLogo), Image: images.Logo, Tint: brightness(&percent)},
			{Rule: greater(PhenomenaShowLogo), Image: images.Logo},
			{Rule: greater(PhenomenaShowLogo), Image: images.LogoMask, Tint: alpha(&whiteMask)},
			{Rule: stages(PhenomenaShowUpperRaster, PhenomenaPhotonFade), Image: images.Middle, Y: 130},
			{Rule: stages(PhenomenaShowUpperRaster, PhenomenaPhotonFade), Image: images.Logo},
			{Rule: stage(PhenomenaShowUpperRaster), Image: images.Raster, Y: 129, Tint: alpha(&percent)},
			{Rule: stages(PhenomenaShowLowerRaster, PhenomenaPhotonFade), Image: images.Raster, Y: 129},
			{Rule: stage(PhenomenaShowLowerRaster), Image: images.Raster, Y: 430, Tint: alpha(&percent)},
			{Rule: stages(PhenomenaDropPhoton, PhenomenaPhotonFade), Image: images.Raster, Y: 430},
			{Rule: stage(PhenomenaDropPhoton), Image: images.Photon, X: 285, YFormula: &secondary},
			{Rule: stage(PhenomenaPhotonFade), Image: images.Photon, X: 285, Y: 445,
				Tint: composite.ScalarTintConfig{Mode: composite.ScalarTintRGB, R: &red, G: &green, B: &green}},
			{Rule: stage(PhenomenaMain), Image: images.Middle, Y: 130},
			{Rule: stage(PhenomenaMain), Image: images.Logo},
			{Rule: stage(PhenomenaMain), Image: images.Raster, Y: 129},
			{Rule: stage(PhenomenaMain), Image: images.Raster, Y: 430},
			{Rule: stage(PhenomenaMain), Image: images.PhotonMask, X: 285, Y: 445,
				Tint: composite.ScalarTintConfig{Mode: composite.ScalarTintHSL, H: &secondary, Saturation: 1, Lightness: .5}},
			{Rule: composite.ScalarStageRule{From: PhenomenaHideLogo, To: PhenomenaHideLogo, DirectionSign: 1}, Image: images.Logo},
			{Rule: stage(PhenomenaHideLogo), Image: images.LogoMask, Tint: alpha(&percent)},
			{Rule: stage(PhenomenaHideLowerRaster), Image: images.Raster, Y: 430, Tint: alpha(&percent)},
			{Rule: stage(PhenomenaHideUpperRaster), Image: images.Raster, Y: 129, Tint: alpha(&percent)},
		},
	}
}
