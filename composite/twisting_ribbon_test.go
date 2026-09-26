package composite

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestTwistingRibbonKeepsDrawBeforeAdvanceFirstFrame(t *testing.T) {
	front, back := ebiten.NewImage(320, 25), ebiten.NewImage(320, 25)
	defer front.Deallocate()
	defer back.Deallocate()
	config := motion.TwistingRibbonConfig{
		Width: 320, StripWidth: 16, IndexStep: 6, PhaseStep: 8, PhaseWrap: 3840,
		FarStart: 1280, NearAmplitude: 15, FarAmplitude: 30, NearScale: 1.5, FarScale: 1,
		AngleMultiplier: 8, NearDivisor: 1280, FarDivisor: 2560,
		BaseAngle: math.Pi + .4, FaceGap: 1.12, BaseY: 30,
		BackVisibleBelow: 1.5, BackVisibleAbove: 3.6,
		FrontVisibleAbove: .5, FrontVisibleBelow: 4.6,
		UpperClipStart: .5, UpperClipEnd: 1.5,
		LowerClipStart: 3.6, LowerClipEnd: 4.6, MinimumHeight: .075,
	}
	ribbon, err := NewTwistingRibbon(TwistingRibbonConfig{
		Front: front, Back: back, Motion: config, SourceHeight: 25,
		FirstUpdateHolds: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := ribbon.Update(kit.Frame{}); err != nil || ribbon.Phase() != 0 {
		t.Fatalf("first update changed phase: %d, %v", ribbon.Phase(), err)
	}
	if err := ribbon.Update(kit.Frame{}); err != nil || ribbon.Phase() != 8 {
		t.Fatalf("second update = phase %d, error %v", ribbon.Phase(), err)
	}
	if len(ribbon.Poses()) != 20 {
		t.Fatal("ribbon lost source strips")
	}
}
