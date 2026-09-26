package recipes

import "github.com/olivierh59500/democonstructionkit/timeline"

// Phenomena stage indices retain the original draw-order grouping in both the
// standalone intro and the Multiscreen panel.
const (
	PhenomenaTextPage1 = iota
	PhenomenaTextPage2
	PhenomenaShowLogo
	PhenomenaShowUpperRaster
	PhenomenaShowLowerRaster
	PhenomenaDropPhoton
	PhenomenaPhotonFade
	PhenomenaMain
	PhenomenaHideLogo
	PhenomenaHideLowerRaster
	PhenomenaHideUpperRaster
	PhenomenaEnd
)

// PhenomenaPresentation describes the intro and exit as editable scalar stages.
// Start may be Main for an embedded panel that skips the standalone opening.
func PhenomenaPresentation(start int) timeline.ScalarStagesConfig {
	initialValue := -40.0
	if start != PhenomenaTextPage1 {
		initialValue = 0
	}
	nextAt := func(compare timeline.ScalarCompare, threshold, value float64, next int) timeline.ScalarRule {
		return timeline.ScalarRule{Compare: compare, Threshold: threshold,
			SetValue: true, Value: value, ChangeStage: true, NextStage: next}
	}
	return timeline.ScalarStagesConfig{
		InitialStage: start, InitialValue: initialValue, InitialDirection: 1,
		Stages: []timeline.ScalarStage{
			{Name: "text-page-one", Delta: 1.5, Rules: []timeline.ScalarRule{
				nextAt(timeline.ScalarGreaterEqual, 340, 0, PhenomenaTextPage2),
			}},
			{Name: "text-page-two", Delta: 1, UseDirection: true, Rules: []timeline.ScalarRule{
				{Compare: timeline.ScalarGreater, Threshold: 200,
					SetValue: true, Value: 100, SetDirection: true, Direction: -1},
				{Compare: timeline.ScalarLessEqual, Threshold: 0, DirectionSign: -1,
					SetValue: true, Value: 0, ChangeStage: true, NextStage: PhenomenaShowLogo},
			}},
			{Name: "show-logo", Delta: 4, Rules: []timeline.ScalarRule{
				nextAt(timeline.ScalarGreaterEqual, 200, 0, PhenomenaShowUpperRaster),
			}},
			{Name: "show-upper-raster", Delta: 4, Rules: []timeline.ScalarRule{
				nextAt(timeline.ScalarGreaterEqual, 100, 0, PhenomenaShowLowerRaster),
			}},
			{Name: "show-lower-raster", Delta: 4, Rules: []timeline.ScalarRule{
				nextAt(timeline.ScalarGreaterEqual, 100, 0, PhenomenaDropPhoton),
			}},
			{Name: "drop-photon", Rules: []timeline.ScalarRule{
				{Event: "photon-landed", SetValue: true, Value: 100,
					ChangeStage: true, NextStage: PhenomenaPhotonFade},
			}},
			{Name: "photon-fade", Delta: -4, Rules: []timeline.ScalarRule{
				nextAt(timeline.ScalarLess, 50, 0, PhenomenaMain),
			}},
			{Name: "main", Rules: []timeline.ScalarRule{
				{Event: "finish", SetValue: true, Value: 50,
					SetDirection: true, Direction: 1, ChangeStage: true, NextStage: PhenomenaHideLogo},
			}},
			{Name: "hide-logo", Delta: 4, UseDirection: true, Rules: []timeline.ScalarRule{
				{Compare: timeline.ScalarGreater, Threshold: 100, DirectionSign: 1,
					SetDirection: true, Direction: -1},
				{Compare: timeline.ScalarLess, Threshold: 0, DirectionSign: -1,
					SetValue: true, Value: 100, SetDirection: true, Direction: -1,
					ChangeStage: true, NextStage: PhenomenaHideLowerRaster},
			}},
			{Name: "hide-lower-raster", Delta: 4, UseDirection: true, Rules: []timeline.ScalarRule{
				{Compare: timeline.ScalarLess, Threshold: 0, SetValue: true, Value: 100,
					SetDirection: true, Direction: -1,
					ChangeStage: true, NextStage: PhenomenaHideUpperRaster},
			}},
			{Name: "hide-upper-raster", Delta: 4, UseDirection: true, Rules: []timeline.ScalarRule{
				{Compare: timeline.ScalarLess, Threshold: 0, ChangeStage: true, NextStage: PhenomenaEnd},
			}},
			{Name: "end"},
		},
	}
}
