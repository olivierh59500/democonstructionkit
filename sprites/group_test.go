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

func TestGroupUsesReusableCuedFormation(t *testing.T) {
	image := ebiten.NewImage(4, 4)
	defer image.Deallocate()
	formation, err := motion.NewCuedFormation(motion.CuedFormationConfig{
		Origin: motion.Point{X: 10, Y: 20}, Spacing: motion.Point{X: 30}, Count: 2,
		Cues: []motion.FormationCue{{Start: 0, Duration: 1, Stagger: .5, LeadIndex: 1, Fade: .1,
			Y: []motion.FormationHarmonic{{FirstAmplitude: 40, LastAmplitude: 40, Cycles: .5}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	group, err := NewGroup(GroupConfig{Frames: []*ebiten.Image{image}, Count: 2, Formation: formation.At, Speed: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := group.Update(kit.Frame{Time: .5}); err != nil {
		t.Fatal(err)
	}
	if got := group.Poses(); got[0].X != 10 || got[0].Y != 20 || got[1].X != 40 || got[1].Y != 60 {
		t.Fatalf("unexpected staged poses: %+v", got)
	}
	if _, err := NewGroup(GroupConfig{Frames: []*ebiten.Image{image}, Count: 2, Formation: formation.At, Weave: &motion.Weave{}}); err == nil {
		t.Fatal("two formation trajectories were accepted")
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

func TestGroupSamplesHarmonicFormationOncePerStateChange(t *testing.T) {
	image := ebiten.NewImage(4, 4)
	defer image.Deallocate()
	formation, err := motion.NewHarmonicFormation(motion.HarmonicFormationConfig{
		Origin: motion.Point{X: 10, Y: 20}, Spacing: motion.Point{X: 8},
		X: []motion.IndexedHarmonic{{Amplitude: 4, Rate: 1, IndexPhase: .5, Cos: true}},
		Y: []motion.IndexedHarmonic{{Amplitude: 2, Rate: 1, SecondaryClock: true, Envelope: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	group, err := NewGroup(GroupConfig{
		Frames: []*ebiten.Image{image}, Count: 3, Harmonic: formation,
		HarmonicClockStep: [2]float64{.2, .3},
		HarmonicEnvelope:  &motion.BounceBankConfig{Start: []float64{5}, Velocity: []float64{1}, Min: 0, Max: 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := group.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	for index, pose := range group.Poses() {
		want := formation.At(index, [2]float64{.2, .3}, 6)
		if pose.X != want.X || pose.Y != want.Y {
			t.Fatalf("pose %d = %+v, want %+v", index, pose, want)
		}
	}
	if _, err := NewGroup(GroupConfig{Frames: []*ebiten.Image{image}, Count: 1, Harmonic: formation, Circle: &motion.CircleFormation{}}); err == nil {
		t.Fatal("accepted two formation modes")
	}
	indexed, err := motion.NewHarmonicFormation(motion.HarmonicFormationConfig{
		IndexOffsets: []float64{0}, X: []motion.IndexedHarmonic{{Rate: 1, UseIndexOffsets: true}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewGroup(GroupConfig{Frames: []*ebiten.Image{image}, Count: 2, Harmonic: indexed}); err == nil {
		t.Fatal("accepted incomplete authored phase table")
	}
	if err := group.ResetHarmonics(); err != nil {
		t.Fatal(err)
	}
	if pose, want := group.Poses()[0], formation.At(0, [2]float64{}, 5); pose.X != want.X || pose.Y != want.Y {
		t.Fatalf("reset pose = %+v, want %+v", pose, want)
	}
	if allocations := testing.AllocsPerRun(100, func() {
		if err := group.Update(kit.Frame{}); err != nil {
			panic(err)
		}
	}); allocations != 0 {
		t.Fatalf("harmonic group update allocates %v times", allocations)
	}
}
