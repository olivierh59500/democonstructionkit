package sprites

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/modulation"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestGroupDistinguishesPathSpacingTimeDelayAndScreenSpacing(t *testing.T) {
	image := ebiten.NewImage(4, 4)
	defer image.Deallocate()
	g, err := NewGroup(GroupConfig{Frames: []*ebiten.Image{image, image, image}, Count: 3, Points: []motion.Point{{X: 0}, {X: 1000}}, Speed: 10, Phase: 100, PhaseSpacing: 20, Delay: 1, Spacing: motion.Point{Y: 7}, FPS: 2, FrameStride: 1})
	if err != nil {
		t.Fatal(err)
	}
	g.Update(kit.Frame{Time: 3})
	for i, p := range g.Poses() {
		if math.Abs(p.X-(130+float64(i)*10)) > 1e-12 || p.Y != float64(i)*7 || p.Frame != []int{0, 2, 1}[i] {
			t.Fatal(i, p)
		}
	}
}
func TestGroupSamplesMusicOnceAndRetainsPreparedPoses(t *testing.T) {
	image := ebiten.NewImage(4, 4)
	defer image.Deallocate()
	scale, _ := modulation.New(modulation.Spec{Base: 1, Inputs: []modulation.Input{{Name: "beat", Gain: .5}}})
	calls := 0
	inputs := map[string]float64{"beat": 1}
	g, err := NewGroup(GroupConfig{Frames: []*ebiten.Image{image}, Count: 40, Velocity: motion.Point{X: 5}, Signals: GroupSignals{ScaleX: scale, ScaleY: scale}, Context: func(f kit.Frame) modulation.Context {
		calls++
		return modulation.Context{Seconds: f.Time, Inputs: inputs}
	}})
	if err != nil {
		t.Fatal(err)
	}
	calls = 0
	g.Update(kit.Frame{Time: 2})
	if calls != 1 || g.Poses()[0].ScaleX != 1.5 || g.Poses()[0].X != 10 {
		t.Fatal(calls, g.Poses()[0])
	}
	dst := ebiten.NewImage(20, 20)
	defer dst.Deallocate()
	g.Draw(dst)
	g.Draw(dst)
	if calls != 1 {
		t.Fatal("drawing sampled audio or advanced motion")
	}
	if allocs := testing.AllocsPerRun(100, func() { g.Update(kit.Frame{Time: 2}) }); allocs != 0 {
		t.Fatal(allocs)
	}
}
