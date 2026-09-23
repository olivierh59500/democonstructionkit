package effects

import (
	"math"
	"testing"

	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

func TestJellyCubeMatchesCompleteLegacyController(t *testing.T) {
	for _, smooth := range []bool{false, true} {
		cube, err := NewJellyCube(DMAJellyCubeConfig())
		if err != nil {
			t.Fatal(err)
		}
		defer cube.Close()
		cube.SetSmoothTransitions(smooth)
		old := newJellyLegacy()
		old.smoothCubeTransitions = smooth
		var expected [8]geometry.Vec3
		// Two complete cycles cover every mode, entrance and handoff in both paths.
		for frame := 0; frame <= 3050; frame++ {
			if frame > 0 {
				if err := old.update(); err != nil {
					t.Fatal(err)
				}
			}
			if err := cube.Update(kit.Frame{Time: float64(frame) / 60}); err != nil {
				t.Fatal(err)
			}
			old.sampleCube(expected[:])
			got := cube.Pose()
			for i := range got {
				delta := got[i].Sub(expected[i])
				if math.Max(math.Abs(delta.X), math.Max(math.Abs(delta.Y), math.Abs(delta.Z))) > 1e-10 {
					t.Fatalf("smooth=%v frame=%d vertex=%d got=%v want=%v", smooth, frame, i, got[i], expected[i])
				}
			}
			if cube.zoom != old.zoom3d {
				t.Fatalf("frame %d zoom got=%v want=%v", frame, cube.zoom, old.zoom3d)
			}
		}
	}
}

func TestJellyCubeTransitionsMatchPreviousPose(t *testing.T) {
	c := DefaultJellyCubeConfig()
	cube, err := NewJellyCube(c)
	if err != nil {
		t.Fatal(err)
	}
	defer cube.Close()
	for frame := 1; frame <= 1500; frame++ {
		before := cube.Pose()
		if err := cube.Update(kit.Frame{Time: float64(frame) / 60}); err != nil {
			t.Fatal(err)
		}
		if frame%300 == 0 && cube.Pose() != before {
			t.Fatalf("mode handoff at frame %d moved its first pose", frame)
		}
	}
}

func TestJellyCubeSeeksAndInstancesAreIndependent(t *testing.T) {
	c := DefaultJellyCubeConfig()
	c.Steps = []JellyCubeStep{{JellyBounce, 1}, {JellyTumble, .5}, {JellySwing, 2}, {JellyNormal, 1}, {JellyPulsate, .25}}
	c.Size = 34
	c.Phase = 1.25
	c.Speed = .7
	a, err := NewJellyCube(c)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := NewJellyCube(c)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	for i := 0; i <= 731; i++ {
		if err := a.Update(kit.Frame{Time: float64(i) / 60}); err != nil {
			t.Fatal(err)
		}
	}
	if err := b.Update(kit.Frame{Time: 731.0 / 60}); err != nil {
		t.Fatal(err)
	}
	if a.Pose() != b.Pose() {
		t.Fatal("skipped updates changed the pose")
	}
	want := b.Pose()
	a.SetPosition(80, 80)
	if err := a.Update(kit.Frame{Time: 30}); err != nil {
		t.Fatal(err)
	}
	if b.Pose() != want {
		t.Fatal("one instance changed another")
	}
	if err := a.Update(kit.Frame{Time: 731.0 / 60}); err != nil {
		t.Fatal(err)
	}
	if a.Pose() != want {
		t.Fatal("backward seek changed the pose")
	}
	for i := 0; i < 4; i++ {
		a.Geometry()
		if err := a.Update(kit.Frame{Time: 731.0 / 60}); err != nil {
			t.Fatal(err)
		}
	}
	if a.Pose() != want {
		t.Fatal("repeated reads advanced motion")
	}
}

func TestJellyCubeConfigurationOwnershipAndValidation(t *testing.T) {
	c := DefaultJellyCubeConfig()
	cube, err := NewJellyCube(c)
	if err != nil {
		t.Fatal(err)
	}
	defer cube.Close()
	c.Steps[0].Mode = 99
	if err := cube.Update(kit.Frame{Time: 1}); err != nil {
		t.Fatal(err)
	}
	if cube.config.Steps[0].Mode != JellyNormal {
		t.Fatal("constructor retained caller-owned sequence")
	}
	for _, edit := range []func(*JellyCubeConfig){
		func(c *JellyCubeConfig) { c.Steps = nil }, func(c *JellyCubeConfig) { c.Steps[0].Duration = 0 },
		func(c *JellyCubeConfig) { c.Size = 0 }, func(c *JellyCubeConfig) { c.Phase = math.NaN() },
		func(c *JellyCubeConfig) { c.Deformation.Squash = 10 },
	} {
		c := DefaultJellyCubeConfig()
		edit(&c)
		if got, err := NewJellyCube(c); err == nil {
			got.Close()
			t.Fatal("invalid configuration accepted")
		}
	}
	if err := cube.Update(kit.Frame{Time: math.NaN()}); err == nil {
		t.Fatal("invalid time accepted")
	}
	cube.Close()
	cube.Close()
	if err := cube.Update(kit.Frame{}); err == nil {
		t.Fatal("closed effect updated")
	}
}

func TestJellyCubeSteadyStateAllocations(t *testing.T) {
	cube, err := NewJellyCube(DefaultJellyCubeConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cube.Close()
	frame := 0
	if n := testing.AllocsPerRun(1000, func() { frame++; _ = cube.Update(kit.Frame{Time: float64(frame) / 60}); cube.Geometry() }); n != 0 {
		t.Fatalf("updates/geometry allocated %g objects", n)
	}
}

func BenchmarkJellyCubeUpdateAndGeometry(b *testing.B) {
	cube, err := NewJellyCube(DefaultJellyCubeConfig())
	if err != nil {
		b.Fatal(err)
	}
	defer cube.Close()
	frame := 0
	b.ReportAllocs()
	for b.Loop() {
		frame++
		_ = cube.Update(kit.Frame{Time: float64(frame) / 60})
		cube.Geometry()
	}
}

func TestJellyCubeRejectsOversizedReplayWithoutChangingPose(t *testing.T) {
	c := DefaultJellyCubeConfig()
	c.MaxReplaySteps = 100
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	cube, err := NewJellyCube(c)
	if err != nil {
		t.Fatal(err)
	}
	defer cube.Close()
	if err := cube.Update(kit.Frame{Time: 1}); err != nil {
		t.Fatal(err)
	}
	pose := cube.Pose()
	if err := cube.Update(kit.Frame{Time: 1e12}); err == nil {
		t.Fatal("unbounded replay accepted")
	}
	if cube.Pose() != pose {
		t.Fatal("rejected replay changed pose")
	}
	c.Phase = 2
	if err := c.Validate(); err == nil {
		t.Fatal("initial phase exceeds replay budget")
	}
}

func TestJellyCubeRejectsUnrepresentableDerivedConfiguration(t *testing.T) {
	tests := map[string]func(*JellyCubeConfig){
		"underflowing scale":             func(c *JellyCubeConfig) { c.Size = math.SmallestNonzeroFloat64 },
		"underflowing radius":            func(c *JellyCubeConfig) { c.Size = 1e-300 },
		"overflowing radius":             func(c *JellyCubeConfig) { c.Size = math.MaxFloat64 },
		"unrepresentable GPU coordinate": func(c *JellyCubeConfig) { c.Size = math.MaxFloat32 },
		"overflowing gain":               func(c *JellyCubeConfig) { c.Deformation.Wobble = math.MaxFloat64 },
		"overflowing rotation":           func(c *JellyCubeConfig) { c.RotationSpeed = math.MaxFloat64 },
		"delay integer boundary":         func(c *JellyCubeConfig) { c.Delay = float64(math.MaxInt64) / 60 },
		"phase integer boundary":         func(c *JellyCubeConfig) { c.Phase = float64(math.MaxInt64) / 60; c.MaxReplaySteps = math.MaxInt64 },
		"step integer boundary":          func(c *JellyCubeConfig) { c.Steps[0].Duration = float64(math.MaxInt64) / 60 },
		"replay integer boundary":        func(c *JellyCubeConfig) { c.MaxReplaySteps = math.MaxInt64 },
	}
	for name, edit := range tests {
		t.Run(name, func(t *testing.T) {
			c := DefaultJellyCubeConfig()
			edit(&c)
			if err := c.Validate(); err == nil {
				t.Fatal("invalid derived configuration passed validation")
			}
			cube, err := NewJellyCube(c)
			if err == nil {
				cube.Close()
				t.Fatal("invalid derived configuration constructed")
			}
		})
	}
	cube, err := NewJellyCube(DefaultJellyCubeConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer cube.Close()
	pose := cube.Pose()
	if err := cube.Update(kit.Frame{Time: float64(math.MaxInt64) / 60}); err == nil {
		t.Fatal("integer-overflowing time accepted")
	}
	if cube.Pose() != pose {
		t.Fatal("invalid time modified the cube")
	}
}
