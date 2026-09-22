package scene

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

func TestGameCloseBeforeInitializationIsIdempotent(t *testing.T) {
	g, err := NewGame(Config{})
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := g.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.Update(); err != ebiten.Termination {
		t.Fatalf("closed game restarted: %v", err)
	}
}

func TestBoundedProfilingConfiguration(t *testing.T) {
	g, err := NewGame(Config{Profile: "profile.json"})
	if err != nil {
		t.Fatal(err)
	}
	if g.config.Frames != 600 || g.warmup != 60 {
		t.Fatal("profiling must have a bounded default and warmup")
	}
	g, err = NewGame(Config{Frames: 40, CaptureFrame: 300})
	if err != nil || g.warmup != 10 {
		t.Fatalf("short run warmup or unused capture option: %v", err)
	}
	for _, c := range []Config{{Frames: -1}, {CaptureFrame: -1}, {Frames: 40, Capture: "capture.png", CaptureFrame: 50}} {
		if _, err := NewGame(c); err == nil {
			t.Fatalf("accepted invalid run: %+v", c)
		}
	}
}

func TestMobileProfileReturnsToContinuousPlayback(t *testing.T) {
	g, err := NewGame(Config{Authoring: true, Frames: 8, ContinueAfterProfile: true})
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	for i := 0; i < 12; i++ {
		if err := g.Update(); err != nil {
			t.Fatal(err)
		}
	}
	if g.Report() == nil || g.finished || g.config.Frames != 0 || g.tick <= 8 {
		t.Fatal("profiling froze interactive playback")
	}
	report := *g.Report()
	for i := 0; i < 5; i++ {
		g.Update()
	}
	if g.Report().Update != report.Update || g.measuring {
		t.Fatal("continuous playback extended the finished report")
	}
}

func TestQualityProfilesKeepTimingAndReduceSurfaceStorage(t *testing.T) {
	high, err := New(false)
	if err != nil {
		t.Fatal(err)
	}
	defer high.Close()
	eco, err := New(true)
	if err != nil {
		t.Fatal(err)
	}
	defer eco.Close()
	for _, s := range []*Scene{high, eco} {
		for _, sample := range []struct {
			seconds float64
			path    int
		}{{0, 0}, {8, 1}, {16, 2}, {24, 0}} {
			if err := s.Update(kit.Frame{Time: sample.seconds}); err != nil {
				t.Fatal(err)
			}
			if s.Path != sample.path {
				t.Fatalf("quality changed path timing: at %g got %d", sample.seconds, s.Path)
			}
		}
		_, alpha, on := s.lensWindow.At(6)
		if !on || alpha != 1 {
			t.Fatal("lens must be fully visible during the bounded capture interval")
		}
		_, _, on = s.lensWindow.At(11)
		if on {
			t.Fatal("lens must release its pass outside the active interval")
		}
	}
	h, e := high.SurfaceInventory(), eco.SurfaceInventory()
	for _, key := range []string{"pipeline_ping_pong_rgba", "layer_fade_surface_rgba", "background_rgba"} {
		if h[key] != 4*e[key] {
			t.Fatalf("economy %s does not quarter logical storage", key)
		}
	}
	if high.water.Config().RowHeight != 0 || eco.water.Config().RowHeight != 4 {
		t.Fatal("unexpected reflection quality settings")
	}
}
