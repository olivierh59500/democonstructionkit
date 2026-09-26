package presets

import (
	"github.com/olivierh59500/democonstructionkit/timeline"
	timelinerecipes "github.com/olivierh59500/democonstructionkit/timeline/recipes"
)

const (
	PhenomenaTextPage1       = timelinerecipes.PhenomenaTextPage1
	PhenomenaTextPage2       = timelinerecipes.PhenomenaTextPage2
	PhenomenaShowLogo        = timelinerecipes.PhenomenaShowLogo
	PhenomenaShowUpperRaster = timelinerecipes.PhenomenaShowUpperRaster
	PhenomenaShowLowerRaster = timelinerecipes.PhenomenaShowLowerRaster
	PhenomenaDropPhoton      = timelinerecipes.PhenomenaDropPhoton
	PhenomenaPhotonFade      = timelinerecipes.PhenomenaPhotonFade
	PhenomenaMain            = timelinerecipes.PhenomenaMain
	PhenomenaHideLogo        = timelinerecipes.PhenomenaHideLogo
	PhenomenaHideLowerRaster = timelinerecipes.PhenomenaHideLowerRaster
	PhenomenaHideUpperRaster = timelinerecipes.PhenomenaHideUpperRaster
	PhenomenaEnd             = timelinerecipes.PhenomenaEnd
)

// PhenomenaPresentation exposes the pure editor-ready stage recipe alongside
// other DCK production presets.
func PhenomenaPresentation(start int) timeline.ScalarStagesConfig {
	return timelinerecipes.PhenomenaPresentation(start)
}
