package scene

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	kit "github.com/olivierh59500/democonstructionkit"
	authored "github.com/olivierh59500/democonstructionkit/examples/authoring/scene"
)

// Config bounds profiling by update ticks. Zero Frames runs interactively.
// Capture reads the final logical image once; it is never part of normal drawing.
type Config struct {
	Eco                  bool
	Authoring            bool
	Frames               int
	Profile              string
	Capture              string
	CaptureFrame         int
	StopWhenDone         bool
	ContinueAfterProfile bool // Mobile can retain the report and resume ordinary playback.
}

type Memory struct {
	HeapAlloc  uint64 `json:"heap_alloc_bytes"`
	HeapInuse  uint64 `json:"heap_inuse_bytes"`
	TotalAlloc uint64 `json:"total_alloc_bytes"`
	Mallocs    uint64 `json:"mallocs"`
	NumGC      uint32 `json:"gc_cycles"`
}
type CPUTime struct {
	Count  int     `json:"count"`
	MeanUS float64 `json:"mean_us"`
	MaxUS  float64 `json:"max_us"`
}
type Report struct {
	Scene          string           `json:"scene"`
	Description    string           `json:"measurement"`
	Resolution     [2]int           `json:"logical_resolution"`
	Eco            bool             `json:"eco"`
	WaterRowHeight int              `json:"water_row_height"`
	WarmupFrames   int              `json:"warmup_frames"`
	ElapsedSeconds float64          `json:"elapsed_seconds"`
	ActualFPS      float64          `json:"ebitengine_actual_fps"`
	ActualTPS      float64          `json:"ebitengine_actual_tps"`
	MeasuredFPS    float64          `json:"measured_draws_per_second"`
	MeasuredTPS    float64          `json:"measured_updates_per_second"`
	RenderedScenes int              `json:"rendered_scenes"`
	Update         CPUTime          `json:"update_cpu"`
	Draw           CPUTime          `json:"draw_cpu_submission"`
	Before         Memory           `json:"memory_before"`
	After          Memory           `json:"memory_after"`
	AllocatedBytes uint64           `json:"allocated_bytes_during_measurement"`
	Allocations    uint64           `json:"allocations_during_measurement"`
	Surfaces       map[string]int64 `json:"logical_surface_bytes"`
	Capture        string           `json:"capture,omitempty"`
}

type timing struct {
	count      int
	total, max time.Duration
}

func (t *timing) add(elapsed time.Duration) {
	t.count++
	t.total += elapsed
	if elapsed > t.max {
		t.max = elapsed
	}
}
func (t timing) report() CPUTime {
	r := CPUTime{Count: t.count, MaxUS: float64(t.max) / 1000}
	if t.count > 0 {
		r.MeanUS = float64(t.total) / float64(t.count) / 1000
	}
	return r
}

// Game retains the rendered frame between updates. A bounded profiling run can
// exit, hold its last image, or resume ordinary playback after saving the report.
type Game struct {
	config                        Config
	scene                         *Scene
	authored                      *authored.Scene
	root                          kit.Effect
	output                        *ebiten.Image
	tick                          uint64
	lastDrawTick                  uint64
	renderedScenes                int
	frame                         kit.Frame
	started                       time.Time
	warmup                        int
	measuring, finished, captured bool
	closed                        bool
	before                        runtime.MemStats
	updateTime, drawTime          timing
	touches                       []ebiten.TouchID
	report                        *Report
	err                           error
}

func NewGame(c Config) (*Game, error) {
	if c.Frames < 0 || c.CaptureFrame < 0 {
		return nil, fmt.Errorf("effectslab: negative frame limit")
	}
	if c.Profile != "" && c.Frames == 0 {
		c.Frames = 600
	}
	if c.Capture != "" && c.CaptureFrame == 0 {
		c.CaptureFrame = 300
	}
	if c.Capture == "" {
		c.CaptureFrame = 0
	}
	if c.Frames > 0 && c.CaptureFrame > c.Frames {
		return nil, fmt.Errorf("effectslab: capture frame exceeds run length")
	}
	g := &Game{config: c, warmup: 60, touches: make([]ebiten.TouchID, 0, 10)}
	if c.Frames > 0 {
		g.warmup = min(60, c.Frames/4)
	}
	return g, nil
}
func (g *Game) initialize() error {
	if g.config.Authoring {
		width, height := 640, 360
		if g.config.Eco {
			width, height = 320, 180
		}
		s, err := authored.NewSize(width, height)
		if err != nil {
			return err
		}
		g.authored, g.root = s, s
		g.output = ebiten.NewImage(width, height)
		return nil
	}
	s, err := New(g.config.Eco)
	if err != nil {
		return err
	}
	g.scene = s
	g.root = s
	g.output = ebiten.NewImage(s.Width, s.Height)
	return nil
}
func (g *Game) Update() error {
	if g.closed {
		return ebiten.Termination
	}
	if g.err != nil {
		return g.err
	}
	if g.root == nil {
		if err := g.initialize(); err != nil {
			return err
		}
	}
	if g.finished {
		if g.config.StopWhenDone {
			return ebiten.Termination
		}
		return nil
	}
	if g.config.Frames > 0 && int(g.tick) >= g.config.Frames {
		if g.config.Capture != "" && !g.captured {
			return nil
		}
		if err := g.finish(); err != nil {
			return err
		}
		if g.config.StopWhenDone {
			return ebiten.Termination
		}
		if g.config.ContinueAfterProfile {
			g.finished, g.measuring = false, false
			g.config.Frames = 0
		}
		return nil
	}
	if !g.measuring && g.report == nil && int(g.tick) >= g.warmup {
		runtime.ReadMemStats(&g.before)
		g.started = time.Now()
		g.measuring = true
	}
	start := time.Now()
	if err := g.handleInput(); err != nil {
		return err
	}
	g.tick++
	g.frame = kit.Frame{Tick: g.tick, Time: float64(g.tick) / 60, Delta: 1.0 / 60}
	if err := g.root.Update(g.frame); err != nil {
		return err
	}
	if g.measuring {
		g.updateTime.add(time.Since(start))
	}
	return nil
}
func (g *Game) Draw(dst *ebiten.Image) {
	if g.root == nil {
		return
	}
	start := time.Now()
	if !g.finished && g.lastDrawTick != g.tick {
		g.output.Clear()
		g.root.Draw(g.output)
		g.lastDrawTick = g.tick
		if g.measuring {
			g.renderedScenes++
		}
	}
	dst.DrawImage(g.output, nil)
	if g.measuring && !g.finished {
		g.drawTime.add(time.Since(start))
	}
	if !g.captured && g.config.Capture != "" && int(g.tick) >= g.config.CaptureFrame {
		g.captured = true
		g.err = writeCapture(g.config.Capture, g.output)
	}
}
func (g *Game) Layout(int, int) (int, int) {
	if g.authored != nil {
		return g.authored.Width, g.authored.Height
	}
	if g.scene != nil {
		return g.scene.Width, g.scene.Height
	}
	if g.config.Eco {
		return 320, 180
	}
	return 640, 360
}

func (g *Game) handleInput() error {
	if g.authored != nil {
		if g.config.StopWhenDone && inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return ebiten.Termination
		}
		return nil
	}
	action := -1
	for i, key := range []ebiten.Key{ebiten.KeyP, ebiten.KeyW, ebiten.KeyL, ebiten.KeyQ} {
		if inpututil.IsKeyJustPressed(key) {
			action = i
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && g.config.StopWhenDone {
		return ebiten.Termination
	}
	g.touches = inpututil.AppendJustPressedTouchIDs(g.touches[:0])
	for _, touch := range g.touches {
		x, y := ebiten.TouchPosition(touch)
		if y >= g.scene.Height-24 {
			action = min(3, max(0, x*4/g.scene.Width))
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if y >= g.scene.Height-24 {
			action = min(3, max(0, x*4/g.scene.Width))
		}
	}
	switch action {
	case 0:
		g.scene.AutoPath = false
		g.scene.Path = (g.scene.Path + 1) % len(g.scene.paths)
	case 1:
		g.scene.Water = !g.scene.Water
	case 2:
		g.scene.Lens = !g.scene.Lens
	case 3:
		if g.config.Frames > 0 {
			return nil
		} // Keep a bounded measurement at one fixed resolution.
		previous := g.scene
		next, err := New(!previous.Eco)
		if err != nil {
			return err
		}
		next.Path, next.Water, next.Lens = previous.Path, previous.Water, previous.Lens
		next.AutoPath = previous.AutoPath
		g.scene = next
		g.root = next
		_ = previous.Close()
		g.output.Deallocate()
		g.output = ebiten.NewImage(next.Width, next.Height)
	}
	return nil
}

func (g *Game) finish() error {
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	elapsed := time.Since(g.started).Seconds()
	if !g.measuring {
		elapsed = 0
	}
	width, height := g.Layout(0, 0)
	eco, waterRows, sceneName := g.config.Eco, 0, "authoring"
	var surfaces map[string]int64
	if g.authored != nil {
		surfaces = map[string]int64{"authoring_scene_rgba": g.authored.SurfaceBytes()}
	} else {
		eco, waterRows, sceneName = g.scene.Eco, 1, "live-effects"
		if eco {
			waterRows = 4
		}
		surfaces = g.scene.SurfaceInventory()
	}
	r := Report{Scene: sceneName, Description: "CPU wall time around Update and Draw submission; not GPU execution time, power consumption, or total process memory. Logical RGBA image sizes exclude backend atlases and driver buffers.",
		Resolution: [2]int{width, height}, Eco: eco, WaterRowHeight: waterRows, WarmupFrames: g.warmup, ElapsedSeconds: elapsed,
		ActualFPS: ebiten.ActualFPS(), ActualTPS: ebiten.ActualTPS(), Update: g.updateTime.report(), Draw: g.drawTime.report(), Before: memoryOf(g.before), After: memoryOf(after),
		AllocatedBytes: after.TotalAlloc - g.before.TotalAlloc, Allocations: after.Mallocs - g.before.Mallocs, Surfaces: surfaces, Capture: g.config.Capture}
	r.RenderedScenes = g.renderedScenes
	r.Surfaces["final_retained_frame_rgba"] = int64(width * height * 4)
	if elapsed > 0 {
		r.MeasuredFPS = float64(g.drawTime.count) / elapsed
		r.MeasuredTPS = float64(g.updateTime.count) / elapsed
	}
	g.report = &r
	g.finished = true
	if g.config.Profile != "" {
		if err := os.MkdirAll(filepath.Dir(g.config.Profile), 0755); err != nil {
			return err
		}
		data, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			return err
		}
		if err = os.WriteFile(g.config.Profile, append(data, '\n'), 0644); err != nil {
			return err
		}
	}
	return nil
}
func memoryOf(m runtime.MemStats) Memory {
	return Memory{HeapAlloc: m.HeapAlloc, HeapInuse: m.HeapInuse, TotalAlloc: m.TotalAlloc, Mallocs: m.Mallocs, NumGC: m.NumGC}
}
func (g *Game) Report() *Report { return g.report }
func (g *Game) Close() error {
	if g == nil || g.closed {
		return nil
	}
	g.closed = true
	if g.output != nil {
		g.output.Deallocate()
		g.output = nil
	}
	err := kit.Close(g.root)
	g.scene = nil
	g.root = nil
	g.authored = nil
	return err
}
func writeCapture(path string, source *ebiten.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	pixels := image.NewRGBA(source.Bounds())
	source.ReadPixels(pixels.Pix)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	if err = png.Encode(file, pixels); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
