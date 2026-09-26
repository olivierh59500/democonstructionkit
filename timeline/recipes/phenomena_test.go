package recipes

import (
	"testing"

	"github.com/olivierh59500/democonstructionkit/timeline"
)

func TestPhenomenaPresentationMatchesIntroAndExitStages(t *testing.T) {
	program, err := timeline.NewScalarStages(PhenomenaPresentation(PhenomenaTextPage1))
	if err != nil {
		t.Fatal(err)
	}
	type referenceState struct {
		stage                       int
		rasterY, percent, direction float64
		photonY, velocity, rebound  float64
		finished                    bool
	}
	reference := referenceState{rasterY: -40, direction: 1, photonY: 184, rebound: -9.50}
	mainTicks, ended := 0, false
	for tick := 0; tick < 5000; tick++ {
		oldStage := reference.stage
		bounceDone := false
		finishCue := oldStage == PhenomenaMain && reference.finished
		switch oldStage {
		case PhenomenaTextPage1:
			reference.rasterY += 1.5
			if reference.rasterY >= 340 {
				reference.rasterY, reference.percent, reference.stage = 0, 0, PhenomenaTextPage2
			}
		case PhenomenaTextPage2:
			reference.percent += reference.direction
			if reference.percent > 200 {
				reference.direction, reference.percent = -1, 100
			}
			if reference.percent <= 0 && reference.direction == -1 {
				reference.percent, reference.stage = 0, PhenomenaShowLogo
			}
		case PhenomenaShowLogo:
			reference.percent += 4
			if reference.percent >= 200 {
				reference.percent, reference.stage = 0, PhenomenaShowUpperRaster
			}
		case PhenomenaShowUpperRaster:
			reference.percent += 4
			if reference.percent >= 100 {
				reference.percent, reference.stage = 0, PhenomenaShowLowerRaster
			}
		case PhenomenaShowLowerRaster:
			reference.percent += 4
			if reference.percent >= 100 {
				reference.percent, reference.stage = 0, PhenomenaDropPhoton
			}
		case PhenomenaDropPhoton:
			reference.velocity += .30
			reference.photonY += reference.velocity
			if reference.photonY > 445 {
				reference.velocity = reference.rebound
				reference.rebound *= .70
			}
			if reference.rebound >= -.70 {
				reference.percent, reference.stage = 100, PhenomenaPhotonFade
				bounceDone = true
			}
		case PhenomenaPhotonFade:
			reference.percent -= 4
			if reference.percent < 50 {
				reference.percent, reference.stage = 0, PhenomenaMain
			}
		case PhenomenaMain:
			if reference.finished {
				reference.percent, reference.direction, reference.stage = 50, 1, PhenomenaHideLogo
			}
		case PhenomenaHideLogo:
			if reference.direction > 0 {
				reference.percent += reference.direction * 4
				if reference.percent > 100 {
					reference.direction = -1
				}
			} else {
				reference.percent += reference.direction * 4
				if reference.percent < 0 {
					reference.percent, reference.direction, reference.stage = 100, -1, PhenomenaHideLowerRaster
				}
			}
		case PhenomenaHideLowerRaster:
			reference.percent += reference.direction * 4
			if reference.percent < 0 {
				reference.percent, reference.direction, reference.stage = 100, -1, PhenomenaHideUpperRaster
			}
		case PhenomenaHideUpperRaster:
			reference.percent += reference.direction * 4
			if reference.percent < 0 {
				reference.stage = PhenomenaEnd
			}
		}
		if bounceDone {
			program.Signal("photon-landed")
		}
		if finishCue {
			program.Signal("finish")
		}
		program.Step()
		got := program.State()
		wantValue := reference.percent
		if reference.stage == PhenomenaTextPage1 {
			wantValue = reference.rasterY
		}
		if got.Stage != reference.stage || got.Value != wantValue || got.Direction != reference.direction {
			t.Fatalf("tick %d: stage/value/direction = %+v, want %d/%v/%v", tick, got,
				reference.stage, wantValue, reference.direction)
		}
		if oldStage == PhenomenaMain && reference.stage == PhenomenaMain {
			mainTicks++
			if mainTicks == 120 {
				reference.finished = true // Input arrives after this tick's update.
			}
		}
		if reference.stage == PhenomenaEnd {
			ended = true
			break
		}
	}
	if !ended || mainTicks < 120 {
		t.Fatalf("sequence did not traverse main and exit: stage=%d main ticks=%d", reference.stage, mainTicks)
	}
	if got := testing.AllocsPerRun(100, func() {
		program.Reset()
		for range 300 {
			program.Step()
		}
	}); got != 0 {
		t.Fatalf("scalar stage program allocated %.2f objects", got)
	}
}

func TestPhenomenaEmbeddedPanelStartsAtMainWithoutIntro(t *testing.T) {
	program, err := timeline.NewScalarStages(PhenomenaPresentation(PhenomenaMain))
	if err != nil {
		t.Fatal(err)
	}
	for range 1200 {
		program.Step()
	}
	if state := program.State(); state.Stage != PhenomenaMain || state.Value != 0 || state.Direction != 1 {
		t.Fatalf("embedded panel entered intro/outro: %+v", state)
	}
}
