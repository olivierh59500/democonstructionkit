# democonstructionkit

A Go/Ebitengine construction kit extracted from the actual productions in `demos/`.
The shared code must preserve their original artwork, text, lookup tables, timing,
pixel rounding, source crops, drawing order and blend operations.

The gallery launches the DCK versions located alongside the preserved originals
in each repository's `dck/` directory. Add `-original` to run the original
implementation. `cmd/studies` contains independent construction examples.

The shared effects support configurable fonts, scroll modes, image deformation,
feedback ribbons, vectorball geometry, particle batches and continuous scene
handoffs. The Cuddly application combines fifteen native screens and its menu in
`demos/go-cuddlymenu/dck`; `demos/go-uniondemo` adds its introduction and eleven
screens. The recipes below apply the same effects to different source images.

## Try the corrected work

From `lib/democonstructionkit`:

```sh
# Original productions using the shared kit; original sound/controls are retained.
go run ./cmd/gallery -demo bilizir-demo
go run ./cmd/gallery -demo viva_tcb
go run ./cmd/gallery -demo megatwist
go run ./cmd/gallery -demo go-multiscreen

# One configurable scroller: mixed fonts, speed/pause/shape/color controls,
# raster bands and independently positioned logos.
go run ./examples/composer
go run ./examples/scrollmodes

# Compare an actual migrated production with its pinned original Git revision.
go run ./cmd/fidelity -demo bilizir-demo
```

`-demos /path/to/demos` selects another checkout. `gallery -list` lists productions.
The launcher runs the migrated source application; it does not replace its scene
script with a kit-generated approximation. Each source repository uses the
published DCK version pinned in its `go.mod`; Go downloads the module automatically.

Bilizir's requested logo variation is enabled by default in `bilizir-demo`: the
logo and scrolling text share the same two-pass deformation. Press **L** to switch
to the original logo. Fidelity comparisons explicitly use original-logo mode.
DMA Is Back similarly enables continuous cube transitions in its DCK version;
baseline comparisons can select the historical transition behavior.

## One scrolling pipeline

```go
scroll, err := scrolling.New(scrolling.Config{
    Fonts: map[string]scrolling.Face{
        "small": {Atlas: smallAtlas, Metrics: smallMetrics},
        "large": {Atlas: largeAtlas, Metrics: largeMetrics},
    },
    Font: "small", Controls: scrolltext.Braces,
    Speed: 120, X: 800, Y: 300, Repeat: true, Gap: 80,
    Text: "HELLO {font:large}WORLD {speed:240}FAST {pause:1}{speed:120}AGAIN ",
})
```

Use `Controls: nil` for literal text. Register any number of independent fonts,
with their own ordering, cell sizes, advances, bearings, aliases and blank glyphs.
The built-in parser supports font, speed, pause, scale, tracking, shape and effect
controls. A custom decoder can retain original syntax or binary control payloads;
`scrolltext.DomSizes` implements the original `^CsN;` syntax.

With `Repeat: true`, `Update` plus `Draw` renders a continuous ribbon: the next
copy follows the previous tail by `Gap` pixels. Short messages repeat enough
times to cover the destination; the first entry still starts at `X`/`Y`, without
an older tail appearing at startup. Speed and pause controls repeat each cycle.
Absolute glyph indices and text offsets keep sine, DNA and 3D phases continuous;
automatic draw samples expose total travelled distance in `Sample.Position`.

Automatic repetition selects whole copies around the destination's pen range,
including glyph overhangs and neighboring copies for deformations. For a path or
projection whose input distances extend beyond that range, set `RepeatBounds`
in **pen coordinates before mapping**, for example
`image.Rect(0, 0, int(math.Ceil(path.Length())), 1)` for a horizontal path.
For closed curves, explicitly clip the mapper to one path length if overlapping
turns are unwanted, or use `Window` to select an exact number of glyphs.
Geometry is rebased near the viewport, so work does not grow with elapsed time.

`Scrolling.DrawAt` also accepts original positions, visible ranges, circular text
windows, reverse drawing order and a glyph mapper. This lets an existing demo keep
its exact tick counters and reset conditions while sharing the renderer. Shader
and animated-strip backends use `DrawState.Paint`, retaining the same iteration
and layout pipeline.
`StateAt` keeps its original single-pass, cycle-local position and control state;
manual `DrawAt`, `Window` and `Ring` retain their existing repetition rules.

## Compose effects without a prescribed layout

- `composite.Sprites`: configurable sprite/logo counts, frames, crops, transforms,
  drawing order and per-instance color/blend/filter options.
- `composite.Strips`: exact row/column sampling for scrollers, distorted logos and
  image-based rasters; source selection is independent of destination geometry.
- `composite.StripWarp`: a reusable row-then-column deformation, with shared
  parameters and independent surfaces sized for each text/logo/sprite source.
- `composite.NewPass` / `Layer`: reusable surfaces, ordered deformation passes,
  clipping, transforms and blending around any effect.
- `composite.QuadBatch`: bounded triangle batches, explicit diagonal selection and
  optional Kage shaders/custom vertex attributes.
- `composite.Repeat`: repeating/rotozoom textures with explicit origin and color.
- `sprites.Projector`: shared vectorball projection, model matrix, image selection,
  camera conventions, depth ordering and optional perspective sprite scaling.
- `sprites.Cube`, `Pyramid`, `Plane`, `Flag`: parametric point-cloud objects and
  anchored wave deformation, independent of the ball images.
- `scrolling.Mode`: normal, bounce, sine, zoom, perspective, TCB forms and DNA;
  select through text controls or a repeating `ModeSequence`, with arbitrary fonts.
- `geometry.Handoff`: match all transformed vertices at effect boundaries.
- `kit.Group` / `Sequence`: arbitrary layer order and local scene timing.

The full [composer example](examples/composer/main.go) demonstrates these choices.
The original demo applications contain the production-specific schedules and data.
Shared code handles rendering; artistic parameters are not replaced with defaults.

Additional packages provide bitmap metrics (`font`), curves/keyframes (`motion`),
geometry, palettes (`indexed`), vector font outlines (`outline`), asset loading,
and device-independent YM/go-zikmu PCM (`sound`). Use `sound/ebiten` with one
application-owned audio context. Existing production audio was preserved during
these visual migrations. Second Reality still uses its original ST3 synchronization;
its complete scene/audio extraction is not finished.

## Evidence and limits

**18 productions:** **167 complete-frame comparisons with zero differing pixels**
against their original revisions. Seventeen have eight checkpoints at frames
0, 1, 60, 240, 600, 1200, 2400 and 4800. TCB has 31 checkpoints through tick 14000,
including each transition into its eight forms. Rendering methods in these applications
actually call the shared kit. This is a component migration, not a claim that all
application code has moved into the library.

**Second Reality:** indexed palette expansion and its table-driven lens are
shared. The extracted lens matches the production on 725 tested placements,
including path positions and clipped/odd-address cases. Other software-rendered
parts and ST3 synchronization remain in the production.

Native comparisons can be generated locally. Audio is disabled during captures;
clock and random-seed inputs are fixed. Captures and development reports remain
local working files.

## Export portfolio videos

FFmpeg and ffprobe must be available on PATH. From this directory:

```sh
go run ./cmd/record-demos -root ../.. -output ../../videos/portfolio -jobs 2
```

Each demo also has `go run ./dck/cmd/video -output /path/to/demo.mp4`.
Looping productions default to three minutes. Cuddly records the introduction,
one minute per screen, original loading transitions, menu navigation and Reset.
FR-010 and Second Reality stop at the end of the complete production, including
Second Reality's final scroll. `-duration 10s` selects a short preview.

Exports contain only the game canvas and its PCM audio, with no desktop capture
or audio device required. Graphics still need a native display. Video is 60 FPS;
Second Reality retains its 70 Hz simulation and music synchronization. Offline
rendering does not change the playback speed. H.264/AAC MP4 files include fast
start metadata, PNG posters and JSON timing reports; the batch command creates
an HTML gallery and manifest. Existing videos are verified and kept on reruns.

The reusable `video.Run` drives games using `sound/output` contexts. Those
contexts delegate to Ebitengine during normal playback and mix on the simulation
clock during recording. `sound/ebiten.NewOutputPlayer` adapts shared PCM streams;
the existing `NewPlayer` API still accepts raw Ebitengine contexts.

## Build and checks

Go 1.26+ is required. Ebitengine and audio versions remain pinned in `go.mod`.
YM playback uses the published `github.com/olivierh59500/ym-player v1.0.0` module.
Go downloads the library dependencies automatically.
Graphics tests require a native/virtual display.

```sh
go test -race ./...
go vet ./...
go run ./cmd/checkassets -demos ../../demos
go run ./cmd/checkaudio -demos ../../demos
go run ./cmd/checkeffects -demos ../../demos
go run ./cmd/fidelity -demo grodan-kvack-kvack-demo
```

All 19 original repository test suites pass after their migrations.


## Practical guide: live layers and effects

The runnable `examples/effectslab` combines an animated logo, text following
paths, water reflection and a temporary magnifier. Its images and bitmap font
are generated in Go; it needs no demo assets. Its scene package is shared by the
desktop and Android hosts.

```sh
go run ./examples/effectslab
```

An effect reads a `kit.Frame`: `Time` and `Delta` are seconds; `Tick` is the
application update counter. Advance your clock once in `Update`. `Draw` samples
that state and may be called several times without moving anything. Use one
fixed TPS, normally 60, independent of the monitor refresh rate.

### Choose the source and the draw order

An image operation accepts any live `*ebiten.Image`: a bitmap, an animated logo,
a scrolling surface, a group of sprites or the complete scene. Render the source
once into a persistent image, then reuse that same image for every operation.
Do not draw from a texture into itself, including overlapping subimages.

`kit.Group{background, logo, scroll}` draws in that order. `kit.NewPipeline`
adds ordered full-image passes and supplies two reusable alternating buffers:

```go
pipeline, err := kit.NewPipeline(scene, 640, 360,
    kit.ImagePass{Apply: func(dst, src *ebiten.Image, frame kit.Frame) {
        dst.DrawImage(src, nil)
        water.Draw(dst, src, frame)
    }},
    kit.ImagePass{Apply: func(dst, src *ebiten.Image, frame kit.Frame) {
        dst.DrawImage(src, nil)
        lens.Draw(dst, src, lensOptions)
    }},
)
```

An existing `ebiten.Game` can be adapted without changing its drawing code:

```go
scene := kit.Func{
    OnUpdate: func(kit.Frame) error { return demo.Update() },
    OnDraw: demo.Draw,
}
```

Use the original game's logical resolution and update rate, and keep its cleanup
under the application's ownership. The adapter is updated once by the pipeline;
do not also call `demo.Update` elsewhere.

Here the magnifier sees the reflection as well as the scene. Put the magnifier
first to reflect its result instead. Put a pipeline around just the scrolling
layer to affect text without changing the background. A pass must produce its
complete output; overlay effects therefore copy `src` before drawing on top.
If every pass is disabled, the pipeline draws its source directly and skips the
intermediate rendering. It updates its source exactly once per frame.

### Reflect a scrolling text or animated logo

```go
waterConfig := composite.DefaultWaterReflectionConfig()
waterConfig.Source = image.Rect(0, 100, 640, 260)
waterConfig.Horizon = 260
waterConfig.ScaleY = .6
waterConfig.Alpha = .55
waterConfig.Fade = .8
waterConfig.Wave = composite.WaterWave{
    Amplitude: 5, Wavelength: 36, Speed: 2, Phase: .4,
}
waterConfig.RowHeight = 2
water, err := composite.NewWaterReflection(waterConfig)
```

`Source` is an absolute crop of the source image; empty means the whole image.
`X` and `Horizon` position the reflected rectangle's upper-left corner. `ScaleY`
changes its height. `Tint` is an `ebiten.ColorScale`, with identity as its zero
value. `Alpha` is overall opacity; `Fade` is the fraction lost at the bottom.
Wave amplitude and wavelength are destination pixels, speed radians/second,
and phase radians. Negative speed reverses motion. `SetConfig` updates parameters
without restarting the animation.

A flat reflection (`Amplitude=0`, `Fade=0`) is one cached image draw. This is the
exact path used by Vectorballs: source `(0,288)-(640,368)`, horizon 400, scale 1,
opacity 0.5. Waves/fading use reusable bounded triangle batches. `RowHeight=1`
is detailed; 2–4 reduces geometry work. The operation never redraws the source
or reads pixels back to the CPU.

### Put a magnifier over an existing scene

```go
lens, err := composite.NewMagnifier()
options := composite.DefaultMagnifierOptions()
options.CenterX, options.CenterY = 320, 160
options.Radius = 64
options.Zoom = 2.2
options.Falloff = 1
options.Feather = 2
options.Opacity = .9
options.Filter = ebiten.FilterLinear
// Draw the ordinary scene first, using a different source image.
lens.Draw(destination, sceneImage, options)
```

The GPU only shades the lens bounding rectangle. Outside the circle, the
existing destination remains untouched. `Falloff=0` gives uniform magnification;
positive values return gradually to normal size at the edge. `Feather` softens
the edge inward in pixels. `Crop` optionally restricts sampling to an atlas
rectangle. Coordinates are relative to the source's upper-left corner at (0,0).
Several lenses can share a source. Radius or opacity zero disables the draw.

The exact Second Reality effect uses a separate `effects.IndexedLens`: its
integer displacement tables and palette-bank masks are compiled once, then
applied to indexed byte buffers. Its DCK version calls this shared mapper.
The analytic GPU lens is intended for arbitrary live images; the table mapper
preserves the historical production's particular distortion and colors.
The previous `effects.Lens(image.Image,...)` CPU utility remains available.

### Supply coordinates, a predefined curve, or your own function

`motion.Path` samples **distance in pixels**, preserving steady travel speed
along a polyline or a sampled curve. Its constructor copies coordinates; `At`
returns the position and a unit tangent without allocating. Open paths clamp
to their endpoints; closed paths wrap in either direction.

```go
path, err := motion.NewSpline([]motion.Point{
    {X: 40, Y: 160}, {X: 180, Y: 70},
    {X: 360, Y: 220}, {X: 600, Y: 140},
}, false, 32)
position, tangent := path.At(seconds * 100) // 100 pixels per second.
```

Use `NewPolyline` for exact straight segments and corners. `NewSpline` uses a
Catmull-Rom curve through the points; it may overshoot sharp corners. For named
curves, sample once when constructing the scene:

```go
path, err := motion.SamplePath(
    motion.SineCurve(motion.Point{X: 20, Y: 160}, 600, 40, 2, 0),
    256, false,
)
```

Other curve constructors are `EllipseCurve`, `FigureEightCurve`,
`LissajousCurve` and `BezierCurve`. `SamplePath` also accepts any
`func(u float64) motion.Point`, where `u` ranges from 0 to 1. More samples improve
shape and distance accuracy but consume construction memory; the runtime lookup
is logarithmic in the number of segments. Build paths outside Update/Draw.

Apply the path to **any font**, including mixed proportional fonts:

```go
mode, err := scrolling.AlongPath(scrolling.PathConfig{
    Path: path, Orient: true, Clip: true, NormalOffset: 0,
})
scroll, err := scrolling.New(scrolling.Config{
    Text: "HELLO FROM THE CONSTRUCTION KIT ",
    Fonts: faces, Font: "small",
    Speed: 100, X: path.Length(),
    Modes: map[string]scrolling.Mode{"curve": mode}, Shape: "curve",
})
```

The current pen X is distance along the path, so font advances and scale remain
meaningful. `Vertical:true` in `PathConfig` uses the pen Y instead. `Orient` aligns
glyphs with the tangent; `Rotation` adds radians and `NormalOffset` moves the
baseline perpendicular to it. `Clip` hides glyph origins outside an open path.
For time-dependent geometry, supply `PathConfig.Sample` instead of `Path`:

```go
Sample: func(distance, seconds float64) (motion.Point, motion.Point) {
    angle := distance / 100
    return motion.Point{X: 320+160*math.Cos(angle), Y: 160+60*math.Sin(angle+seconds)},
        motion.Point{X: -160*math.Sin(angle), Y: 60*math.Cos(angle+seconds)}
},
```

Register multiple modes and select them with `{shape:curve}` controls or
`scrolling.NewModeSequence`. Use `scrolling.Chain` to combine mappers. Operation
order matters: a displacement before a rotation is not the same as one after it.

### Trigger, fade and overlap effects

A `timeline.Window` is an absolute-time interval. Duration 0 means no end;
otherwise its end is exclusive. Fades are seconds and sampling is stateless,
so jumps, seeks and pauses do not accumulate timing errors.

```go
layers, err := kit.NewLayers(640, 360,
    kit.TimedLayer{Effect: background},
    kit.TimedLayer{Effect: scroll, Window: timeline.Window{
        Start: 2, Duration: 12, FadeIn: .5, FadeOut: 1,
    }},
    kit.TimedLayer{Effect: logo, Window: timeline.Window{Start: 5, Duration: 7}},
)
```

Layers are drawn in order and overlap freely. Inactive children are neither
updated nor drawn. Fade composition shares one scratch image. `LocalTime:true`
subtracts the activation time from `Frame.Time`; `Frame.Tick` remains global.
`SetWindow(index, window)` can retrigger a layer from a key, touch or music event.
Close the layer tree once; do not share one mutable child between owning trees.

For a temporary magnifier in a pipeline:

```go
window := timeline.Window{Start: 8, Duration: 4, FadeIn: .2, FadeOut: .4}
pass := kit.ImagePass{
    Enabled: func(f kit.Frame) bool { _, _, on := window.At(f.Time); return on },
    Apply: func(dst, src *ebiten.Image, f kit.Frame) {
        dst.DrawImage(src, nil)
        _, alpha, _ := window.At(f.Time)
        options.Opacity = alpha
        lens.Draw(dst, src, options)
    },
}
```

Changing `window.Start` on the game goroutine schedules another appearance.
Place this pass after all other composition to magnify the complete demo.

### Tiled waves, trails and colored masks

`composite.CellWarp` moves rigid fragments of any image. It supports independent
cell width/height, source cropping, destination gaps, filtering and a
`Sample(row,column,frame)` callback returning `CellTransform`. Row/column identity
is preserved when fragments cross; this differs from stretching a continuous
mesh. Union's introduction uses it for crossed waves on 32×16 fragments.

`motion.NewPointHistory(capacity, initial)` stores recent positions (or any Go
value). `Push` belongs in Update, `At(delay)` in sampling/rendering. A 61-entry
history provides delays 0, 20, 40, 60 for four sprites without frame allocations.
Union's hidden screen uses this API. An analytic trajectory can instead be
sampled at earlier times without a history buffer.

To color text with animated rasters, draw the text into a transparent surface,
then draw the raster using `ebiten.BlendSourceAtop`; existing text alpha is
preserved. `BlendSourceIn` also applies the mask to the incoming content but
removes destination content outside it. Clear scratch targets before reuse.
`effects.Mask` packages an independently animated color source and alpha source.

### Pixel and mobile resource budgets

- Create fonts, shaders, paths and working images once. Source/crop changes may
  rebuild small caches; stable frames reuse them.
- Render a shared scene once. Use Pipeline's two buffers rather than repeated
  full-scene renders for each overlay. A 640×360 RGBA surface is about 0.88 MiB;
  two buffers are about 1.76 MiB, before driver/texture overhead.
- Start at 640×360 and test 320×180 for a low-resolution style. Halving both
  dimensions reduces pixel work and render-target memory to one quarter.
- Water RowHeight 2–4 reduces geometry work. Lens radius bounds fragment work;
  avoid large overlapping lenses unless measured. Disable inactive passes.
- Keep `ReadPixels`, `ebiten.Image.At`, PNG export and GPU readback out of the frame loop.
  CPU indexed effects should keep their persistent index buffer and upload once.
- CPU microbenchmarks do not measure GPU time, battery draw or thermal behavior.
  Measure a release/debug-equivalent scene on the actual device, warmed up,
  with stable resolution, refresh rate and effects. The example's timings report
  CPU submission time and observed FPS/TPS; they are not power measurements.

### Shared effects and remaining production-specific code

All 20 current demo repositories use DCK, but that does not mean every algorithm
has been extracted. Reusable families include bitmap scrolls and control modes,
DNA/feedback, row/column/cell warps, masks, rasters, sparkles, sprite projection,
meshes, paths, histories, reflections, magnifiers and scheduled composition.

The latest production migrations preserve Vectorballs' reflection and Second
Reality's indexed lens. Twelve Union intro/pointer-trail captures match exactly
before and after their extraction. Generic GPU magnifier tests compare against
an independent pixel model; they do not imply the analytic effect is identical
to every historical indexed lens.

FR-010 still has specialized part renderers. Second Reality retains software
plasma, tunnel, fire/water and other fixed-point/palette routines as well as its
ST3 synchronization. Cuddly/Union retain scene choreography, assets and artistic
constants. Further extraction should preserve those integer tables and compare
frames, rather than replacing a distinctive effect with a vaguely similar one.


Measured on the connected Pixel 10a with the Android effects laboratory (900
updates, first 60 excluded; all demonstrated effects enabled):

| Logical resolution | Water rows | Observed updates/s | Mean Update CPU | Mean Draw submission CPU | Logical image storage |
| --- | --- | --- | --- | --- | --- |
| 640×360 | 1 | 59.64 | 12.1 µs | 992.9 µs | 4.55 MiB |
| 320×180 | 4 | 60.00 | 12.6 µs | 695.4 µs | 1.22 MiB |

These short runs include Ebitengine/runtime activity and are not isolated GPU,
thermal or battery measurements. Both maintained about 60 updates/s. The complete
application still allocated about 7.2 MB over 14 seconds (including engine work),
so the zero-allocation primitive benchmarks must not be read as a claim of a
zero-allocation application. The lower-resolution mode reduces surface memory
and CPU draw submission, with intentionally coarser output.
