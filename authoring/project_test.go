package authoring

import (
	"bytes"
	"image"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	bitmap "github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/modulation"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

func testProject() Project {
	return Project{Version: 1, Units: DefaultUnits(), Canvas: Canvas{Width: 64, Height: 48, TPS: 60}, Assets: map[string]string{"tile": "image", "letters": "font"}, Layers: []Layer{{ID: "background", Kind: "background", Background: &Background{Image: "tile", Period: Point{X: 8, Y: 8}}}}}
}
func TestProjectRoundTrip(t *testing.T) {
	p := testProject()
	p.Signals = map[string]modulation.Spec{"pulse": {Base: 1, Oscillators: []modulation.Oscillator{{Shape: modulation.Sine, Amplitude: .25, Frequency: 2}}}}
	var encoded bytes.Buffer
	if err := Encode(&encoded, p); err != nil {
		t.Fatal(err)
	}
	got, err := Decode(&encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*got, p) {
		t.Fatalf("round trip changed project\ngot %+v\nwant %+v", got, p)
	}
}
func TestDecodeRejectsAmbiguousDocuments(t *testing.T) {
	var encoded bytes.Buffer
	if err := Encode(&encoded, testProject()); err != nil {
		t.Fatal(err)
	}
	valid := encoded.String()
	for _, input := range []string{
		strings.Replace(valid, "\"version\": 1", "\"version\": 1, \"version\": 1", 1),
		strings.Replace(valid, "\"version\": 1", "\"version\": 1, \"Version\": 2", 1),
		strings.Replace(valid, "\"canvas\":", "\"Canvas\":", 1),
		strings.Replace(valid, "\"width\": 64", "\"width\": 64, \"Width\": 8", 1),
		strings.Replace(valid, "\"tps\": 60", "\"tps\": null", 1),
		strings.Replace(valid, "\"period\": {", "\"period\": {\"X\":1,", 1),
		strings.Replace(valid, "\"version\": 1", "\"version\": 1, \"unused\": 0", 1),
		strings.Replace(valid, "\"image\": \"tile\"", "\"image\": \"tile\", \"unknown\": true", 1),
		valid + " {}", valid[:len(valid)/2], "null", "[]",
	} {
		if _, err := Decode(strings.NewReader(input)); err == nil {
			t.Fatalf("accepted malformed document: %s", input)
		}
	}
}
func TestValidateRejectsUnsupportedConfigurations(t *testing.T) {
	for name, mutate := range map[string]func(*Project){
		"version": func(p *Project) { p.Version = 2 }, "units": func(p *Project) { p.Units.Time = "ticks" }, "canvas": func(p *Project) { p.Canvas.Width = 0 }, "tps": func(p *Project) { p.Canvas.TPS = 1000 },
		"unknown kind": func(p *Project) { p.Layers[0].Kind = "unknown" }, "asset": func(p *Project) { p.Layers[0].Background.Image = "absent" }, "asset kind": func(p *Project) { p.Layers[0].Background.Image = "letters" },
		"filter": func(p *Project) { p.Layers[0].Background.Filter = "unknown" }, "blend": func(p *Project) { p.Layers[0].Blend = "unknown" }, "negative period": func(p *Project) { p.Layers[0].Background.Period.X = -1 },
		"nan": func(p *Project) { p.Layers[0].Background.Camera.X = math.NaN() }, "duration": func(p *Project) { p.Layers[0].Window.Duration = -1 }, "fade": func(p *Project) { p.Layers[0].Window.FadeOut = 1 },
		"two configs": func(p *Project) { p.Layers[0].Scroll = &Scroll{} }, "mismatch": func(p *Project) { p.Layers[0].Kind = "sprites" }, "duplicate id": func(p *Project) { p.Layers = append(p.Layers, p.Layers[0]) },
		"beat tempo": func(p *Project) { p.Signals = map[string]modulation.Spec{"x": {TimeBase: modulation.Beats}} },
	} {
		t.Run(name, func(t *testing.T) {
			p := testProject()
			mutate(&p)
			if err := p.Validate(); err == nil {
				t.Fatal("accepted invalid configuration")
			}
		})
	}
}
func TestScrollAndSpriteValidation(t *testing.T) {
	p := testProject()
	base := Scroll{Text: "AB{font:other}B{shape:wave}A", Fonts: map[string]string{"main": "letters", "other": "letters"}, Font: "main", Controls: "braces", Modes: map[string]Mode{"wave": {Kind: "sine", Wave: &Wave{Amplitude: 4, Speed: 2}}}, Repeat: true, Page: &Page{Width: 60, LineHeight: 8, Align: "center"}}
	p.Layers = []Layer{{ID: "text", Kind: "scroll", Scroll: &base}}
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"{font:missing}", "{shape:missing}", "{speed:NaN}", "{pause:-1}", "{scale:0,1}", "{effect:gold}"} {
		base.Text = text
		if err := p.Validate(); err == nil {
			t.Fatalf("accepted %s", text)
		}
	}
	p.Signals = map[string]modulation.Spec{"x": {Base: 1}}
	group := SpriteGroup{Images: []string{"tile"}, Count: 4, Points: []Point{{}, {X: 40, Y: 20}}, Signals: Bindings{"scaleX": "x"}}
	p.Layers = []Layer{{ID: "sprites", Kind: "sprites", Sprites: &group}}
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	group.Signals["unknown"] = "x"
	if err := p.Validate(); err == nil {
		t.Fatal("accepted unknown binding")
	}
	delete(group.Signals, "unknown")
	group.Signals["scaleX"] = "missing"
	if err := p.Validate(); err == nil {
		t.Fatal("accepted unknown signal")
	}
	group.Signals = nil
	group.Orbit = &Orbit{}
	if err := p.Validate(); err == nil {
		t.Fatal("accepted competing trajectories")
	}
}
func TestModesUseIndependentLibraryConfigurations(t *testing.T) {
	for _, mode := range []Mode{{Kind: "normal"}, {Kind: "sine", Wave: &Wave{}}, {Kind: "bounce", Wave: &Wave{}}, {Kind: "zoom", Zoom: &Zoom{Base: Point{X: 1, Y: 1}}}, {Kind: "perspective", Perspective: &Perspective{Focal: 100, Near: 1}}, {Kind: "path", Path: &Path{Points: []Point{{}, {X: 100, Y: 20}}, SplineSamples: 8}}} {
		if _, err := compileMode(mode, false); err != nil {
			t.Fatal(mode.Kind, err)
		}
	}
	for _, mode := range []Mode{{Kind: "unknown"}, {Kind: "normal", Wave: &Wave{}}, {Kind: "zoom", Wave: &Wave{}}, {Kind: "sine", Wave: &Wave{}, Zoom: &Zoom{}}, {Kind: "bounce", Wave: &Wave{Spatial: 1}}, {Kind: "perspective", Perspective: &Perspective{Focal: 0, Near: 1}}, {Kind: "path", Path: &Path{Points: []Point{{}}}}} {
		if _, err := compileMode(mode, false); err == nil {
			t.Fatal("accepted", mode)
		}
	}
}

func testAssets(t *testing.T) Assets {
	t.Helper()
	tile := ebiten.NewImage(8, 8)
	atlas := ebiten.NewImage(16, 8)
	t.Cleanup(tile.Deallocate)
	t.Cleanup(atlas.Deallocate)
	metrics, err := bitmap.NewGrid(bitmap.Grid{Bounds: atlas.Bounds(), Cell: image.Pt(8, 8), Columns: 2, Order: "AB"})
	if err != nil {
		t.Fatal(err)
	}
	return Assets{Images: map[string]*ebiten.Image{"tile": tile}, Fonts: map[string]scrolling.Face{"letters": {Atlas: atlas, Metrics: metrics}}}
}
func TestCompileMixedLayersAndBorrowedAssets(t *testing.T) {
	p := testProject()
	p.Signals = map[string]modulation.Spec{"x": {Base: 4}}
	p.Layers = append(p.Layers, Layer{ID: "sprites", Kind: "sprites", Window: Window{Start: 1, Duration: 2, FadeIn: .5, FadeOut: .5}, LocalTime: true, Sprites: &SpriteGroup{Images: []string{"tile"}, Count: 2, Spacing: Point{X: 10}, Signals: Bindings{"x": "x"}}}, Layer{ID: "scroll", Kind: "scroll", Scroll: &Scroll{Text: "AB\nBA", Fonts: map[string]string{"main": "letters"}, Font: "main", Speed: 10, Page: &Page{Width: 60, LineHeight: 9, Align: "center"}, Repeat: true}})
	assets := testAssets(t)
	compiled, err := Compile(p, assets, Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, seconds := range []float64{0, 1, 1.5, 2.9, 3, 0} {
		if err := compiled.Update(kit.Frame{Time: seconds}); err != nil {
			t.Fatal(err)
		}
	}
	if err := compiled.Close(); err != nil {
		t.Fatal(err)
	}
	if err := compiled.Close(); err != nil {
		t.Fatal(err)
	}
	// Closing one compiled project must leave shared resolver images usable.
	second, err := Compile(p, assets, Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	if second.Canvas() != p.Canvas {
		t.Fatal("canvas changed")
	}
}
func TestCompileReportsMissingAssetsAndCropBounds(t *testing.T) {
	p := testProject()
	if _, err := Compile(p, nil, Options{}); err == nil {
		t.Fatal("nil resolver")
	}
	if _, err := Compile(p, Assets{}, Options{}); err == nil {
		t.Fatal("missing image")
	}
	assets := testAssets(t)
	p.Layers[0].Background.Source = &Rect{Width: 9, Height: 8}
	if _, err := Compile(p, assets, Options{}); err == nil {
		t.Fatal("out of bounds crop")
	}
	p.Layers[0].Background.Source = &Rect{X: int(^uint(0)>>1) - 1, Width: 8, Height: 8}
	if err := p.Validate(); err == nil {
		t.Fatal("accepted rectangle endpoint overflow")
	}
	p.Layers[0].Background.Source = nil
	p.Layers[0].Background.Period = Point{X: 1e-6}
	if _, err := Compile(p, assets, Options{}); err == nil {
		t.Fatal("accepted excessive background density")
	}
	p.Layers[0].Background.CopiesX = 3
	compiled, err := Compile(p, assets, Options{})
	if err != nil {
		t.Fatalf("finite background copies rejected: %v", err)
	}
	compiled.Close()
	p.Layers[0].Background.CopiesX = -1
	if err := p.Validate(); err == nil {
		t.Fatal("negative background copy count accepted")
	}
}

func TestBackgroundFiniteCopiesRoundTrip(t *testing.T) {
	p := testProject()
	p.Layers[0].Background.CopiesX = 3
	p.Layers[0].Background.CopiesY = 2
	p.Layers[0].Background.Velocity = Point{X: -48, Y: 12}
	var encoded bytes.Buffer
	if err := Encode(&encoded, p); err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(&encoded)
	if err != nil {
		t.Fatal(err)
	}
	c := decoded.Layers[0].Background
	if c.CopiesX != 3 || c.CopiesY != 2 || c.Velocity != (Point{X: -48, Y: 12}) {
		t.Fatalf("finite background settings changed: %+v", c)
	}
	compiled, err := Compile(*decoded, testAssets(t), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := compiled.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRotozoomLayerRoundTripAndCompile(t *testing.T) {
	p := testProject()
	p.Layers[0] = Layer{ID: "rotating-tile", Kind: "rotozoom", Rotozoom: &Rotozoom{
		Image: "tile", Center: Point{X: 32, Y: 24}, Phase: Point{X: 80, Y: 40}, Zoom: 1.5, Rotation: .2,
		CenterVelocity: Point{X: -12, Y: 4}, PhaseVelocity: Point{X: 8, Y: -3}, RotationVelocity: .5,
	}}
	var encoded bytes.Buffer
	if err := Encode(&encoded, p); err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(&encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*decoded, p) {
		t.Fatalf("rotozoom project changed during round trip: %+v", decoded.Layers[0])
	}
	compiled, err := Compile(*decoded, testAssets(t), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := compiled.Update(kit.Frame{Time: 2}); err != nil {
		t.Fatal(err)
	}
	if err := compiled.Close(); err != nil {
		t.Fatal(err)
	}
	p.Layers[0].Rotozoom.Zoom = -1
	if err := p.Validate(); err == nil {
		t.Fatal("accepted negative rotozoom zoom")
	}
}

func TestCopperBarsLayerRoundTripAndCompile(t *testing.T) {
	p := testProject()
	p.Layers[0] = Layer{ID: "rasters", Kind: "copper_bars", CopperBars: &CopperBars{
		Image: "tile", Offsets: []int{0, 4, 8, 12}, Height: 8, Count: 4,
		RowStep: 2, SourceStep: 2, SourcePeriod: 8, BaseX: 6, XShift: 1,
		VelocityA: 1, VelocityB: -1, IndexStepA: 2, IndexStepB: 3,
		Clock: "masked", DrawMode: "images",
	}}
	var encoded bytes.Buffer
	if err := Encode(&encoded, p); err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(&encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*decoded, p) {
		t.Fatalf("copper project changed during round trip: %+v", decoded.Layers[0])
	}
	compiled, err := Compile(*decoded, testAssets(t), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := compiled.Update(kit.Frame{Time: 1}); err != nil {
		t.Fatal(err)
	}
	if err := compiled.Close(); err != nil {
		t.Fatal(err)
	}
	p.Layers[0].CopperBars.Clock = "unknown"
	if err := p.Validate(); err == nil {
		t.Fatal("accepted unknown copper clock")
	}
}

func TestAssetIDsRetainCaseAndFixedArraysRejectTruncation(t *testing.T) {
	p := testProject()
	p.Assets["Tile"] = "image"
	var encoded bytes.Buffer
	if err := Encode(&encoded, p); err != nil {
		t.Fatal(err)
	}
	if _, err := Decode(bytes.NewReader(encoded.Bytes())); err != nil {
		t.Fatal(err)
	}
	input := strings.Replace(encoded.String(), "\"background\": [\n      0,\n      0,\n      0,\n      0\n    ]", "\"background\": [255]", 1)
	if input == encoded.String() {
		t.Fatal("test failed to replace background array")
	}
	if _, err := Decode(strings.NewReader(input)); err == nil {
		t.Fatal("accepted incomplete RGBA")
	}
}

func TestNamedSignalsDriveCompiledSpritePoses(t *testing.T) {
	assets := testAssets(t)
	signal, err := modulation.New(modulation.Spec{Base: 2, Inputs: []modulation.Input{{Name: "voice", Gain: 3}}})
	if err != nil {
		t.Fatal(err)
	}
	b := compiler{resolver: assets, images: map[string]*ebiten.Image{}, signals: map[string]*modulation.Signal{"voice-scale": signal}, context: func(f kit.Frame) modulation.Context {
		return modulation.Context{Seconds: f.Time, Inputs: map[string]float64{"voice": .5}}
	}}
	effect, err := b.sprites(SpriteGroup{Images: []string{"tile"}, Count: 2, Origin: Point{X: 10, Y: 20}, Velocity: Point{X: 6, Y: 2}, Spacing: Point{X: 30}, Delay: .5, Scale: Point{X: 2, Y: 3}, Signals: Bindings{"scaleX": "voice-scale"}, PerInstance: []Bindings{nil, {"y": "voice-scale"}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := effect.Update(kit.Frame{Time: 2}); err != nil {
		t.Fatal(err)
	}
	poses := effect.(*sprites.Group).Poses()
	if poses[0].X != 22 || poses[0].Y != 24 || poses[0].ScaleX != 7 || poses[0].ScaleY != 3 {
		t.Fatal(poses[0])
	}
	if poses[1].X != 49 || poses[1].Y != 26.5 || poses[1].ScaleX != 7 {
		t.Fatal(poses[1])
	}
}

func TestCompileRejectsExcessiveRepeatedTextDensity(t *testing.T) {
	p := testProject()
	p.Layers = []Layer{{ID: "dense", Kind: "scroll", Scroll: &Scroll{
		Text: "{scale:1e-9,1}A", Controls: "braces", Fonts: map[string]string{"main": "letters"}, Font: "main", Repeat: true, Speed: 10,
	}}}
	if _, err := Compile(p, testAssets(t), Options{}); err == nil {
		t.Fatal("accepted unrenderable repeated text density")
	}
}
