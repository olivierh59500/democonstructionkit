package sprites

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/modulation"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestCircleGroupAccumulatesIndependentPoses(t *testing.T) {
	img := ebiten.NewImage(4, 4)
	defer img.Deallocate()
	circle := motion.CircleFormation{
		RadiusX: 150, RadiusY: 150, IndexCount: 3,
		XAmplitude: 20, XRate: 2, XIndexPhase: 1,
		YAmplitude: 20, YRate: 2, YIndexPhase: 1,
		ScaleBase: .5, ScaleAmplitude: .5, ScaleRate: 1, ScaleIndexPhase: .5,
	}
	g, err := NewGroup(GroupConfig{
		Frames: []*ebiten.Image{img}, Count: 3, Circle: &circle,
		Origin: motion.Point{X: 320, Y: 200}, PhaseStep: .02,
		AnchorX: .5, AnchorY: .5,
	})
	if err != nil {
		t.Fatal(err)
	}
	phase := 0.0
	for tick := 0; tick < 1200; tick++ {
		phase += .02
		if err := g.Update(kit.Frame{}); err != nil {
			t.Fatal(err)
		}
		for i, got := range g.Poses() {
			angle := phase + float64(i)*math.Pi*2/3
			wantX := math.Cos(angle)*150 + math.Sin(phase*2+float64(i))*20 + 320
			wantY := math.Sin(angle)*150 + math.Cos(phase*2+float64(i))*20 + 200
			wantScale := .5 + .5*math.Sin(phase+float64(i)*.5)
			if math.Abs(got.X-wantX) > 1e-10 || math.Abs(got.Y-wantY) > 1e-10 || math.Abs(got.ScaleX-wantScale) > 1e-10 || got.ScaleX != got.ScaleY {
				t.Fatalf("tick %d sprite %d = %+v, want (%v, %v, %v)", tick, i, got, wantX, wantY, wantScale)
			}
		}
	}
}

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
