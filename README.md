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
[Complete components](#build-a-demo-from-complete-components) include the animated
DMA cube controller and renderer, font construction, perspective text crawl and
independent background bands; their consumers provide assets and configuration.

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

# Three independently configured cubes, including JSON save/load.
go run ./examples/jellycubes

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

### Select layout and transport through the same constructor

`scrolling.New` now selects the following implementations. The separate transport
settings preserve different control timing and history rules; all return the same
`*scrolling.Scrolling` effect and can use `Output` image processing.

| Configuration | Layout and control timing | Advance units |
| --- | --- | --- |
| Regular `Text` / `Fonts` | Horizontal text; `Vertical:true` stacks glyphs in a moving column; cursor-triggered controls | `Speed`: pixels/second |
| `Page` | Newline-separated horizontal lines moving vertically; mixed-font alignment and line-position controls | `Speed`: pixels/second |
| `Recycled` | Reusable visible glyph slots; controls execute as slots recycle | `Ring.Speed`: pixels/Update |
| `Projected` | Depth-sorted visible slots; forms change while visiting those slots | `PixelsPerUpdate`; `Planes.PhaseStep` per Update |
| `Sliced` | Bounded history of glyph strips; a control consumes a transport step; rotation is independent | `SlicesPerUpdate`; `RotationSpeed`: film frames/second |
| `Crawl` | Cached paragraph window projected through configurable rows; bounded surfaces independent of text length | `PixelsPerUpdate`; zero selects manual `SetPosition` on a standalone `Crawl` |
| `Bands` | Independently moving repeated text lanes in a fixed viewport | Lane `Speed`: pixels/second, or pixels/tick with `UseTicks` |
| `Slots` | Recycled bitmap glyphs with waves, optional tangent rotation and a custom pose mapper | `Speed`: pixels/Update |

Choose one transport. `Page` belongs to the regular transport. For the other
transports, configure text, fonts, speed and repetition inside that transport rather
than also setting regular `Text`, `Fonts`, `Speed` or `Repeat`. `Projected.Draw`
and `Sliced.Draw` hold their output placement. A fixed-step setting of 2 at
60 TPS advances 120 pixels/second; changing TPS changes this authored cadence.

The following helpers can be placed in an application's Go package. Package
names match the DCK subpackages; `kit` denotes the module root, and `bitmap`
denotes `github.com/olivierh59500/democonstructionkit/font`. Assets and metrics
are passed in rather than assumed to use any particular alphabet or dimensions.

```go
func NewAtlasFace(atlas *ebiten.Image, cell image.Point, columns int, order string) (scrolling.Face, error) {
    metrics, err := bitmap.NewGrid(bitmap.Grid{
        Bounds: atlas.Bounds(), Cell: cell, Columns: columns, Order: order,
        Advance: float64(cell.X), LineHeight: float64(cell.Y),
    })
    if err != nil { return scrolling.Face{}, err }
    return scrolling.Face{Atlas: atlas, Metrics: metrics}, nil
}
```

For example, pass `image.Pt(16, 26)` with that atlas's actual column count and
character order, or entirely different metrics for another font. `Grid.Stride`
and `Origin` handle gutters/margins; `font.New` accepts individually positioned
rectangles and proportional advances when a regular grid is inappropriate.

```go
func NewCredits(face scrolling.Face) (*scrolling.Scrolling, error) {
    return scrolling.New(scrolling.Config{
        Text: "DESIGN\nOLIVIER\n\nTHANK YOU FOR WATCHING",
        Fonts: map[string]scrolling.Face{"credits": face}, Font: "credits",
        Y: 360, Speed: 35, Repeat: true, Gap: 100,
        Page: &scrolling.PageConfig{
            Width: 640, LineHeight: 28, Align: scrolling.AlignCenter,
        },
    })
}
```

`Page.LineHeight` is a minimum: larger glyphs expand their line. It lays out
explicit newlines; it does not automatically wrap prose to the specified width.
A plain vertical glyph column needs only `Vertical:true` instead of `Page`.

For editable layouts, call `ValidateRenderBounds` with the destination rectangle
before playback. Automatic rendering has a `MaxGlyphsPerDraw` budget (65,536 by
default), preventing a nearly zero text advance from generating millions of
copies. `Err()` exposes a draw-budget failure; the following Update returns it.
Explicit manual `DrawAt` windows remain under the caller's control.

Automatic glyph rendering also rejects offscreen images after all mappers have
run. A path or projection can therefore move an initially distant glyph into
view. Custom painters retain their own clipping policy, and manual DrawAt keeps
its original visit order and range.

```go
func NewRecycledScroll(grid scrolling.BitmapGrid) (*scrolling.Scrolling, error) {
    return scrolling.New(scrolling.Config{
        X: 20, Y: 280,
        Recycled: &scrolling.RecycledConfig{Ring: scrolling.RingConfig{
            Text: "A CONTINUOUS WAVE OF TEXT ", Font: grid,
            Viewport: 600, Speed: 2,
            Waves: []scrolling.RingWave{{
                Amplitude: 20, LetterStep: .3, TickStep: .04,
            }},
        }},
    })
}
```

`BitmapGrid` has independent `Width`, `Height`, `Columns`, `Order` or `First`,
including fractional cells and an optional `ColumnSpan`. Regular `Face` uses
`font.NewGrid` or `font.New` for proportional advances, bearings, gaps, aliases,
blank glyphs and arbitrary cell order. These metrics belong to the asset;
movement, projection and deformation settings belong to the effect.

```go
func NewPlaneScroll(face scrolling.Face, advance float64) (*scrolling.Scrolling, error) {
    return scrolling.New(scrolling.Config{
        Projected: &scrolling.ProjectedConfig{
            Face: face, PixelsPerUpdate: 2,
            Planes: scrolling.PlanesConfig{
                Slots: presets.TCBPlaneSlots("^0NORMAL ^1BOUNCE ^2SINE ^4ZOOM ^5DEPTH ", advance),
                Forms: presets.TCBScrollForms(), Visible: 30, PhaseStep: .02,
                Projection: scrolling.PlaneProjection{
                    Focal: 250, Depth: 150, OriginX: -450,
                    CenterX: 160, CenterY: 100,
                },
            },
            Draw: scrolling.PlaneDraw{ScaleX: 2, ScaleY: 2},
        },
    })
}
```

The preset retains the original visible-slot control convention, including its
control-slot spacing. New productions can instead use regular modes selected by
`{shape:name}` or `ModeSequence`. `Projected.Raster` optionally colors glyphs
using a raster in scene coordinates, independently of atlas positions.

```go
func NewStripScroll(film *scrolling.DNAFrames, tokens []scrolling.SliceToken) (*scrolling.Scrolling, error) {
    return scrolling.New(scrolling.Config{
        Sliced: &scrolling.SlicedConfig{
            Film: film, SlicesPerUpdate: 2, RotationSpeed: 18,
            Stream: scrolling.SliceStreamConfig{
                Tokens: tokens, Capacity: 640, SliceWidth: 1,
                Repeat: true, LoopStart: 0,
                Initial: scrolling.DNASlice{Glyph: -1},
            },
            Draw: scrolling.DNADrawConfig{
                SliceWidth: 1, ScaleX: 1, ScaleY: 1, OriginY: 160,
            },
        },
    })
}
```

Build the borrowed film once with `scrolling.NewDNAFrames` from any glyph images;
front/back/core color images are optional. Tokens contain either `Glyph` and
pixel `Width`, or a named `Control`. `LoopStart` selects a token after a one-time
prefix. `SliceWidth=1` preserves all integer advances; coarser strips round the
advance up to a strip boundary and retain the final partial image strip.
`Draw.SliceWidth` must match the stream.

`OnControl` can trigger scene actions; returning true stops the remaining step
budget. `AdvanceAt(frame)` can return zero to pause insertion while rotation
continues. `RotationAt(frame)` can retain an existing phase accumulator exactly;
`Offsets` adds one phase per visible slot. The low-level `SliceStream` remains
available for direct `Step`, `SetFrames`, `Cursor`, `Reset`, `Slices` and `Head`
access. Phenomena and its Multiscreen variant share that implementation.

### Stack whole-image effects after any scroll

Regular glyph modes run first. A mode has one painter and optional depth sorter;
`scrolling.Chain` combines mappers. `Output.Passes` composes operations on the
complete text image, including slice warps, masks, reflections and temporary
lenses. The same `kit.ImagePass` can process a logo or scene in `kit.NewPipeline`.

```go
func NewLensScroll(face scrolling.Face) (*scrolling.Scrolling, error) {
    lens, err := composite.NewMagnifier()
    if err != nil { return nil, err }
    options := composite.DefaultMagnifierOptions()
    options.CenterX, options.CenterY = 320, 40
    options.Radius, options.Zoom = 35, 2
    scroll, err := scrolling.New(scrolling.Config{
        Text: "MAGNIFY ANY BITMAP FONT ",
        Fonts: map[string]scrolling.Face{"main": face}, Font: "main",
        X: 640, Y: 20, Speed: 100, Repeat: true, Gap: 80,
        Output: &scrolling.OutputConfig{
            Width: 640, Height: 100,
            Passes: []kit.ImagePass{{
                Enabled: func(f kit.Frame) bool { return f.Time >= 3 && f.Time < 9 },
                Apply: func(dst, src *ebiten.Image, f kit.Frame) {
                    dst.DrawImage(src, nil)
                    lens.Draw(dst, src, options)
                },
                Close: lens.Close,
            }},
        },
    })
    if err != nil { lens.Close(); return nil, err }
    return scroll, nil
}
```

`Output.Feedback` optionally inserts the text into one or more persistent DNA
faces before the image passes. Each `FeedbackLayer` has its own profile, direction,
insertion/speed parameters, gradient, position and phase clock. History advances
once per `Update`, including when rendering multiple views of the same effect.
Feedback output consists of those faces; add a separate layer when the ordinary
text should also remain visible. Draw and layer placement never advance history.

Close `Scrolling` when finished to release its output surfaces, feedback and pass
resources. Font atlases, raster images and a supplied `DNAFrames` film remain
caller-owned. Resources used by custom mode painters also remain with their owner.

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

**Earlier baseline migration: 18 productions, 167 complete-frame comparisons with zero differing pixels**
against their original revisions. Seventeen have eight checkpoints at frames
0, 1, 60, 240, 600, 1200, 2400 and 4800. TCB has 31 checkpoints through tick 14000,
including each transition into its eight forms. Rendering methods in these applications
actually call the shared kit. This is a component migration, not a claim that all
application code has moved into the library.

**Second Reality:** indexed palette expansion and its table-driven lens are
shared, together with lookup plasma, sparse image mapping and palette transitions. The extracted lens matches the production on 725 tested placements,
including path positions and clipped/odd-address cases. Other software-rendered
parts and ST3 synchronization remain in the production.

Native comparisons can be generated locally. Audio is disabled during captures;
clock and random-seed inputs are fixed. Captures and development reports remain
local working files.

## Reusable backgrounds, particles and sprite formations

### Scroll a cropped backdrop with independent parallax

```go
func NewBackdrop(tile *ebiten.Image) (*composite.BackgroundLayer, error) {
    config := composite.DefaultBackgroundConfig()
    config.Source = tile.Bounds()
    config.PeriodX = float64(tile.Bounds().Dx())
    config.PeriodY = float64(tile.Bounds().Dy())
    config.ParallaxX = .35
    renderer, err := composite.NewBackground(config)
    if err != nil { return nil, err }
    return &composite.BackgroundLayer{
        Renderer: renderer, Image: tile,
        Sample: func(f kit.Frame) composite.BackgroundPose {
            return composite.BackgroundPose{CameraX: 80 * f.Time}
        },
    }, nil
}
```

`Source` is an absolute image/atlas crop. A zero period disables repetition on
that axis. Periods are source pixels and can differ from the crop dimensions:
larger periods leave gaps; smaller periods overlap copies in row/column order.
`ScaleX/Y`, filter, blend and tint are independent. Zero parallax intentionally
pins an axis; use `DefaultBackgroundConfig` for unit parallax defaults.

Combine several `BackgroundLayer` values with different parallax factors for
scenery. `Background.Draw` accepts a pose directly; `DrawAt` accepts existing
screen offsets. A destination subimage defines a viewport without another render
target. Matching integer-scaled nearest-filtered tiles use a repeat-addressed
quad; fractional, overlapping and linear-filtered cases retain individual crop
boundaries. Only visible copies are submitted, including after long camera travel.

`MaxCopies` bounds the fallback image submissions (16,384 by default); the
single-quad path remains independent of tile count. Excessive-density frames
are skipped as a whole and reported by `Background.Err()` and the layer's next
Update. Saved-project compilation checks this budget before animation starts.

### Use one particle field for stars, incoming sprites and trails

`sprites.ProjectedField` owns movement, projection and the bounded renderer.
Its `FieldConfig` selects depth behavior and spawning; `FieldStyle` selects
pixels, image sprites, atlas frames or vector rectangles with trails. A nil
skin image draws pixels. All materials use the same cached positions and history.

```go
func NewFlight(skin *ebiten.Image) (*sprites.ProjectedField, error) {
    size := 3.0
    if skin != nil { size = 24 }
    return sprites.NewProjectedField(sprites.ProjectedFieldConfig{
        Field: sprites.FieldConfig{
            Count: 96, Depth: sprites.DepthWrap, Near: 20, Far: 600,
            Spawn: func(i int, reset bool) sprites.Point {
                angle := float64(i) * 2.399963229728653
                return sprites.Point{X: 180 * math.Cos(angle),
                    Y: 100 * math.Sin(angle), Z: 20 + float64((i*37)%580)}
            },
        },
        View: sprites.FieldView{Camera: geometry.Camera{
            Center: geometry.Vec2{X: 320, Y: 180}, Focal: 220, Near: 20,
        }},
        Velocity: geometry.Vec3{Z: -1.5}, Delta: 1,
        Style: sprites.FieldStyle{
            Image: skin, ScaleByDepth: true,
            Appearance: sprites.FieldAppearance{
                Width: size, Height: size, AnchorX: .5, AnchorY: .5,
            },
        },
    })
}
```

Use `DepthFree`, `DepthWrap` or `DepthRespawn`; a respawn callback receives
`reset=true`. `FieldView.Offset` also permits absolute-time travel, and `Angle`
rotates the view axis. `SortDepth` selects stable far-to-near ordering.
`ResetCount` changes population while reusing available storage. One `Update`
advances the configured velocity and projects all points. `DrawStyle` redraws
the same samples with another material or blend, so Union Starballs can place
one sprite image behind its logo and another through its mask.

`FieldStyle.Sample` can change size, tint, rotation or visibility from depth,
index or modulation. `Frames` are atlas rectangles selected by each point's
`Image`. `Streak:true` draws between successive sampled positions; recycling
breaks the connection automatically. `VectorRects:true` draws each filled
rectangle followed by its optional `TrailWidth` line in source order, with
independent `FillColor` and `TrailColor`. `VectorLines:true` draws only the
variable-width vector line. `sprites.StreakField(config)` compiles Big Sprite's
streak behavior into the same `ProjectedField`: strict one-step X/Y/depth wraps,
a cold first frame, its original projection arithmetic and antialiased strokes.
The existing `sprites.NewStreaks` API remains available. `DrawImages:true`
retains exact image transform/crop arithmetic; the ordinary path batches
geometry. Close the `ProjectedField`, not its borrowed skin.

For a complete radial star recipe, start with `presets.DefaultNonamenoStarsConfig()`.
Its count, speed, camera, spawn strides, pixel size, brightness, colors and
trail threshold are plain editable values. Compile it with
`presets.NonamenoProjectedStars(config)`, then pass the result to
`sprites.NewProjectedField`. Nonameno, Union Starballs, Cuddly Starwars and
Cuddly Big Sprite now use the same transport and renderer. Their 11, 16, 9
and 12 sampled RGB frames respectively match the previous productions exactly.

### Control count, spacing, delay and music-driven properties

```go
func NewFormation(frames []*ebiten.Image, path *motion.Path) (*sprites.Group, error) {
    pulse, err := modulation.New(modulation.Spec{
        Base: 1, TimeBase: modulation.Beats,
        Oscillators: []modulation.Oscillator{{
            Shape: modulation.Cosine, Amplitude: .12, Frequency: 1,
        }},
    })
    if err != nil { return nil, err }
    return sprites.NewGroup(sprites.GroupConfig{
        Frames: frames, Count: 12, FPS: 8, FrameStride: 1,
        Path: path, Speed: 90, PhaseSpacing: -36, Orient: true,
        AnchorX: .5, AnchorY: .5,
        Signals: sprites.GroupSignals{ScaleX: pulse, ScaleY: pulse},
        Context: func(f kit.Frame) modulation.Context {
            return modulation.MusicContext(f.Time, 120, 0, nil)
        },
    })
}
```

`PhaseSpacing` is distance along a `Path` in pixels. `Delay` instead samples each
instance at an earlier time in seconds; `Spacing` adds a screen-space offset.
These settings can be combined deliberately. `Points` supplies serializable path
coordinates; `SplineSamples` chooses Catmull-Rom sampling. `Orbit` selects a
`motion.NestedOrbit`, and `Weave` selects phased harmonic motion. Choose one
trajectory; without one, `Velocity` supplies linear movement. `SetPhase` retains
an authored phase counter and `SetCount` reuses capacity where possible.

Groups prepare poses during Update, so drawing them in several layers keeps
animation and audio sampling synchronized. `Signals` affects the group;
`PerInstance` adds individual bindings for X/Y, scale, angle and opacity. Position
and angle are additive; scale and opacity multiply. Animation frames use FPS,
frame offset/stride and the same instance delay.

`modulation.Spec` stores a base value, keys, oscillators, named input gains and an
optional clamp as data. Keys support linear/smooth/hold interpolation and optional
looping. Its oscillator frequency and phase use **cycles**, while `motion.Wave`
and geometric rotations use **radians**. A seconds or beats timebase is explicit.
The example supplies a 120 BPM visual clock; it does not analyze audio. For music
synchronization, provide the audible playback position and sampled inputs once
per Update through `modulation.Context`. Decoder read-ahead is not the audible
playhead. `modulation.Change[T]` and `Decay` provide reusable event-triggered
peaks/decays for voice changes, beats or user actions.

## Deform arbitrary meshes and drive authored scanline programs

```go
func NewJellyModel(model effects.Mesh) (*effects.MeshEffect, error) {
    deformation, err := geometry.NewDeformProgram([]geometry.DeformStage{
        {Kind: geometry.DeformWobble, Wobble: geometry.WobbleField{
            Key: geometry.Vec3{X: .01, Y: .02, Z: .03},
            Amount: 20, Radius: 80, BaseInfluence: .5, RadialInfluence: .5,
            X: []geometry.Harmonic{{Gain: .4, Speed: 1, Spatial: 5}},
            Y: []geometry.Harmonic{{Gain: .4, Speed: 1.3, Spatial: 7, Cosine: true}},
        }},
        {Kind: geometry.DeformTwist, Twist: geometry.TwistField{
            Axis: 1, Radius: 80, Amount: .4,
        }},
    })
    if err != nil { return nil, err }
    mesh, err := effects.NewMesh(model, nil, geometry.Camera{
        Center: geometry.Vec2{X: 320, Y: 180}, Focal: 320, Near: 1,
    })
    if err != nil { return nil, err }
    mesh.Deform = deformation.Deform
    mesh.Animate = func(seconds float64) effects.Transform {
        return effects.Transform{
            Position: geometry.Vec3{Z: 500},
            Rotation: geometry.Vec3{X: seconds * .4, Y: seconds * .7}, Scale: 1,
        }
    }
    return mesh, nil
}
```

The same modifier program can process any point array with `Apply(dst, source,
seconds)`, including vectorballs; in-place operation is supported. Stage order
is explicit: scale, capture reference coordinates, harmonic wobble, radial ripple,
axis twist, XYZ rotation and translation can be composed as needed. Original
coordinates remain available for spatial phases; `DeformReference` chooses the
coordinate space used by subsequent radial effects.

Use `SetVector`, `SetWobbleAmount` and `SetTwistAmount` to animate existing stages
without rebuilding storage. Harmonics have independent gains, spatial weights,
frequencies, phases and sine/cosine choices. Rotation preparation is cached for
vertices sampled at the same time. Apply `geometry.Handoff` to the complete final
vertex set when changing programs; stable vertex correspondence preserves the
previous pose exactly at the switch.

DMA-is-back now uses these stages for its jelly cube. Its authored controllers
and face renderer remain local to preserve face grouping, projection and drawing
order. `effects.MeshEffect` is available for new solid/textured/Glenz models;
its triangle renderer is not claimed to replace every legacy quad/material
renderer identically.

```go
func NewRibbonProgram() (*composite.DisplacementProgram, error) {
    curves, err := presets.RibbonCurves(1)
    if err != nil { return nil, err }
    intro, err := composite.JoinDeltaCurves(curves, []int{0, 0})
    if err != nil { return nil, err }
    loop, err := composite.JoinDeltaCurves(curves, []int{1, 4, 2, 3, 7})
    if err != nil { return nil, err }
    return composite.NewDisplacementProgram(intro, loop)
}
```

`DeltaCurve` also compiles your own sine/cosine terms, alternating row signs,
attack/release envelopes and forward drift into integer deltas. `Step`, `Extent`,
`Degrees`, `Drift` and the endpoint policy are explicit. `JoinDeltaCurves` retains
cumulative displacement across curve boundaries. `DisplacementProgram.At` samples
one index; `Fill` writes successive row offsets into caller-owned storage without
per-row division or allocation. The optional intro plays once; the loop continues
its forward drift rather than snapping to zero.

Coco, DMA-is-back, Multiscreen's Coco scene and Megatwist share these programs.
Megatwist's background and text retain separate phase indices and intro sequences.
The resulting offsets can drive any row/column image sampler, sprite position or
text cursor. Source wrapping/clamping, strip thickness and masks remain separate
image-processing choices, rather than being hardwired to a particular font.

## Procedural pixels, indexed mappings and camera tracks

These kernels are pure Go and operate on caller-owned buffers. Their output can
be an ordinary layer, an alpha mask or a live texture on another effect. Palette
choice, geometry and scene placement are separate from pixel generation.

```go
func NewPlasmaUpdater(texture *ebiten.Image) (func(kit.Frame) error, error) {
    width, height := texture.Bounds().Dx(), texture.Bounds().Dy()
    config := plasma.DefaultHarmonicConfig(width, height)
    config.Waves[2].CenterX = float64(width) / 2
    config.Waves[2].CenterY = float64(height) / 2
    config.Waves[2].Speed = -.5
    kernel, err := plasma.NewHarmonic(config)
    if err != nil { return nil, err }
    pixels := make([]byte, width*height*4)
    return func(frame kit.Frame) error {
        if err := kernel.RenderRGBA(pixels, width*4, frame.Time); err != nil {
            return err
        }
        texture.WritePixels(pixels)
        return nil
    }, nil
}
```

Assign the returned callback to a layer's `kit.Func.OnUpdate`, and draw the
texture in its `OnDraw`, or sample it as a mesh texture. `HarmonicWave` supports
horizontal, vertical, diagonal and radial terms with independent frequency,
speed, phase, amplitude and center. `Divisor`, `ColorFrequency`, RGB channel
sine/cosine weights and `MaximumColor` define color presentation. Spatial terms
are compiled once; construct a new kernel outside the frame loop when changing
spatial parameters. `RenderRGBA` reuses internal buffers and writes opaque RGBA;
use separate instances for concurrent rendering. This is a CPU kernel.

For indexed/table-driven plasma, use `plasma.NewLookup(plasma.LookupConfig)`.
Each `Wave` has its own displacement table and color/displacement X/Y increments,
offsets and fractional address bits. Color and displacement tables have
power-of-two lengths. Pass one `plasma.Phase` per wave to `Render`; contributions
sum modulo 256. `RenderRows` selects interleaved or partial rows without modifying
other rows or padding. It performs no trigonometry during rendering. The compiled
lookup can be shared between goroutines writing different destination buffers.
Use `indexed.ExpandRGBA` with a premultiplied packed RGBA palette, then upload
once; retain the original index buffer when only the palette changes.

```go
func NewSparseMap() (*indexed.ScatterMap, error) {
    return indexed.NewScatterMap(indexed.ScatterConfig{
        SourceSize: 3, DestinationSize: 8,
        Offsets: []uint32{0, 2, 2, 3},
        Destinations: []uint32{1, 4, 6},
    })
}
```

Here source pixel 0 maps to destination pixels 1 and 4; source pixel 1 is omitted;
source pixel 2 maps to pixel 6. `ScatterMap.Render` accepts any indexed source,
including text or a logo. `ScatterReplace` copies indices, `ScatterOverBackground`
restores the background for source zero, and `ScatterAddBackground` adds indices
modulo 256. Destination collisions retain source order; unmapped pixels are
untouched. `DecodeScatterMap16` compiles existing count/address tables once.
Source and destination must not overlap; background may alias destination.

`indexed.FadePalette` interpolates packed RGB entries toward a color;
`MixPalette` interpolates two palettes. Both use explicit integer `amount/steps`,
retain six-bit or eight-bit component precision, and support in-place output.
Slice a palette range to keep reserved colors unchanged. These operations are
independent of the scroller, plasma, water or mesh producing the indices.

```go
func NewCameraTrack() (*motion.BSpline32, error) {
    return motion.NewBSpline32([]float32{
        0, 0, 0, 0,
        1, 10, 0, 0,
        2, 20, 5, 0,
        3, 30, 5, 1,
    }, 3)
}
```

Keys interleave time with the requested number of components. Call
`track.Sample(position[:], at)` with a reusable `[3]float32` position buffer.
This uniform cubic B-spline uses float32 arithmetic and normally does not pass
through its control points. The final three keys support the curve rather than
adding segments. Outside its authored range, the first/last segment extrapolates;
clamp the time explicitly when required. Use 1–64 components for camera position,
target or other grouped parameters. It is distinct from the distance-based
Catmull-Rom `motion.Path` used by regular text and sprite trajectories.

## Save a composition and reload it

`authoring` compiles a versioned JSON project into the same DCK effects. It is the
basis for an eventual graphical editor: asset IDs, ordered layers, timing,
parameters, signal bindings and validation already exist independently of a UI.
No graphical editor is included yet.

```sh
# Save the procedural example, then edit its JSON and reload it.
go run ./examples/authoring -save /tmp/dck-project.json -frames 600
go run ./examples/authoring -project /tmp/dck-project.json

# Render a bounded preview; capture happens outside ordinary animation work.
go run ./examples/authoring -frames 600 -capture /tmp/dck-project.png

# Use the shared desktop/mobile laboratory host for profiling this composition.
go run ./examples/effectslab -authoring -frames 900 -profile /tmp/dck-authoring.json
```

The example's resolver supplies `checker`, `orb-cyan`, `orb-pink`, `small` and
`bright`. Other applications supply their own assets through `authoring.Resolver`
or `authoring.Assets`; physical paths and font metrics stay outside the project.
A minimal project using the example's font is:

```json
{
  "version": 1,
  "units": {"distance": "pixels", "time": "seconds", "angle": "radians"},
  "canvas": {"width": 640, "height": 360, "tps": 60, "background": [5, 8, 20, 255]},
  "assets": {"small": "font"},
  "layers": [{
    "id": "message", "kind": "scroll",
    "window": {"start": 1, "fadeIn": 0.5}, "localTime": true,
    "scroll": {
      "text": "A SAVED DCK COMPOSITION ",
      "fonts": {"main": "small"}, "font": "main",
      "origin": {"x": 640, "y": 180},
      "speed": 100, "gap": 80, "repeat": true
    }
  }]
}
```

```go
func LoadComposition(input io.Reader, assets authoring.Resolver) (*authoring.Compiled, error) {
    project, err := authoring.Decode(input)
    if err != nil { return nil, err }
    return authoring.Compile(*project, assets, authoring.Options{})
}
```

The host selects the canvas dimensions and TPS from `Compiled.Canvas()`, supplies
one `kit.Frame` per Update, draws the compiled effect and closes it afterwards.
`authoring.Encode` saves validated data. Version 1 rejects unknown fields,
duplicate keys, unsupported versions, invalid asset/mode references and incompatible
units. Compilation borrows resolver assets and owns only its working resources.

The serialized subset is deliberately explicit:

| Version 1 layer | Supported saved parameters |
| --- | --- |
| `scroll` | Font bank, text/braces controls, horizontal/vertical/page layout, repeat bounds/gap, normal/sine/bounce/zoom/perspective/path modes, timed mode sequence |
| `sprites` | Image animation, count, linear/path/orbit/weave formation, spacing/delay, transform, blend/filter, common and per-instance signal bindings |
| `background` | Source crop, repeat period, scale, parallax, camera/movement velocities, blend/filter |
| `jelly_cube` | Editable default/DMA preset, center, half-edge, camera, six colors, mode order/durations, phase, speed, entrance, transition and five deformation gains |
| All layers | ID, order, start/duration/fades, local clock and final layer blend |

Signals store keys, oscillators and named inputs. `Options.Context` can supply
music time and live values; otherwise project BPM and layer time define the beat
clock. Recycled/projected/sliced compatibility transports, DNA, arbitrary image
passes, particle fields, arbitrary mesh deformation, procedural kernels and Go callbacks
remain Go APIs rather than saved version-1 layer kinds. Unsupported settings are
rejected instead of silently discarded. A future schema can add these named
families without putting effect equations inside an editor.

## Systematic extraction: coverage and current evidence

The source audit covers all 20 demo repositories, including every Cuddly and
Union presentation unit and the parts of Second Reality and FR-010. It examines
animation, control timing, sampling, masking, camera conventions and buffers;
package imports alone are not evidence that an effect has been shared. Detailed
per-screen working reports remain outside the module's Git tree.

Recent independent regression suites include:

| Extraction | Verified comparison |
| --- | --- |
| Backgrounds and choreography | 73 identical PNG pairs across 11 screens; 288 GPU background cases |
| Shared particle field | Five Cuddly starfield and five Union incoming-sprite captures, identical |
| Sprite group / envelopes / vertical transport | Six Digi, five Delta and five Level16 captures, identical |
| DMA cube deformation | 42 identical GPU captures across five modes and their handoffs |
| Sliced DNA transport | 11 identical captures per scene for Phenomena and Multiscreen, through 48,000 updates; original transport/control oracle over three text loops |
| Authored ribbon programs | Every integer sample of eleven curves at five distortion rates |
| Indexed plasma | 60 original phase cases, both fields and interleaved/drop updates |
| Indexed scatter | 96 Forest/Water buffer comparisons |
| Harmonic RGB plasma | Every RGBA byte at four sizes and six times |
| Float32 camera tracks | All nine FR-010 camera/target tracks |

These suites can overlap; the rows are not summed into a unique-frame total.
Existing original packages remain available as references. The pure geometry
comparison allows only 1e-11 float64 variation from compiler operation fusion;
submitted cube images are still checked byte for byte. Comparison render targets
are unmanaged, preventing unrelated texture-atlas placement from moving edges.

Measured M4 Max CPU examples: shared cube deformation takes about 443 ns for its
eight points with zero allocations; indexed plasma about 79 µs versus 98 µs for
the previous kernel; scatter about 3.56 µs versus 8.84 µs. The generic harmonic
plasma is about 455 µs versus 442 µs, a small overhead for configurability. These
are kernel microbenchmarks, not Pixel, whole-frame, GPU or power measurements.
Do not infer zero allocation for the complete application from these kernels.

The audit also records remaining work. Scene scripts, asset selection, local
controls, exact raster/material conventions and some specialized renderers still
live in the productions. Examples include software tunnel/fire/water simulations,
legacy quad/face sorting and materials, particular text-page choreography and
ST3 synchronization. Common callbacks allow authored behavior to compose today;
not every callback has a built-in configurable or serializable equivalent yet.

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

## One music entry point

Native DCK demos open music through DCK. They do not import a decoder or keep
a local YM-to-PCM player implementation. The filename and file signature select
the backend internally; replacing a YM with a MOD/XM/S3M/IT does not require a
different player class. Recorded WAV, MP3 and Ogg/Vorbis tracks use the same API.

```go
func StartMusic(name string, data []byte) (*playback.Player, error) {
    player, err := playback.Open(nil, name, data, sound.Options{
        Loop: true, Interpolation: true,
    })
    if err != nil { return nil, err }
    player.Play()
    return player, nil
}
```

Here `playback` is `github.com/olivierh59500/democonstructionkit/sound/ebiten`;
`sound` is the sibling `sound` package. Close the returned player when the scene
ends. A nil context reuses the current DCK output context or creates one. This
also works with the offline recording mixer. Playback controls and the audible
position remain on the returned player.

For hosts that already own playback, `sound.Open(name, data, options)` returns a
seekable `*sound.Stream`; `sound.OpenFS(files, name, options)` reads an embedded or
filesystem asset. Opening either stream alone opens no audio device. The default
output is 48 kHz stereo float32. `PCMFormat: sound.PCM16` selects stereo int16;
`NewOutputPlayer` and `NewPlayer` choose their PCM device path automatically.

- Recognized content takes precedence over a misleading extension. Packed YM
  files and gzip-wrapped recordings are handled inside DCK.
- `Gain` selects initial attenuation (zero selects unit gain); `SetVolume(0)`
  mutes. Stream gain clamps to 0..1. PCM16 truncation and optional `Quantize16`
  preserve the authored integer gain when old scenes use float32 output.
- `Loop` supports native module endings through go-zikmu. `LoopStartFrame`
  preserves a recorded introduction before repeating the selected PCM region.
- `BlockFrames` controls decode-ahead, independently of caller Read sizes. A
  capture host can choose 800 frames at 48 kHz/60 Hz for exact register cues.
- `Metadata()` returns source format, title, author/comment when available and
  known duration. `Length()` and `Seek` use bytes in the selected PCM format;
  the stream's position counts delivered bytes, not the audible device cursor.
- Packed two-song FC containers use `Track: 0` or `Track: 1`. DCK selects the
  timing-compatible S3M profile in go-zikmu internally, including packed pattern
  handling. `StartOrder` and `TrackerPositionAt(player.Position())` preserve
  music-driven scene cues. The compatibility core retains its original GPL
  license, documented with its source in go-zikmu.

The old explicit `NewYM`, `NewModule` and `NewPCM16` constructors remain available
for compatibility and low-level integrations. Demo clients use `Open`. Backend
requirements can still appear in a demo repository's go.mod because the preserved
original implementation shares that module; its DCK variant has no direct
decoder imports. `cmd/checkboundaries` verifies that rule across all 20 adapters.

Regression coverage includes 16 independent one-second PCM fingerprints from the
previous adapters, the longer FR-010 filtered/attenuated stream, both Second
Reality soundtracks and their tracker markers, recorded loop seams, partial
reads, gain changes, seeks, decoder selection and callback allocation checks.

## Build and checks

Go 1.26+ is required. Ebitengine and audio versions remain pinned in `go.mod`.
YM playback uses the published `github.com/olivierh59500/ym-player v1.0.0` module.
Go downloads the library dependencies automatically.
Graphics tests require a native/virtual display.

```sh
go test -race ./...
go vet ./...
go run ./cmd/checkassets -demos ../../demos
go run ./cmd/checkfontpixels -demos ../../demos
go run ./cmd/checkaudio -demos ../../demos
go run ./cmd/checkboundaries -demos ../../demos
go run ./cmd/checkeffects -demos ../../demos
go run ./cmd/fidelity -demo grodan-kvack-kvack-demo
```

Run production suites from their own repositories as well as the module tests;
the preserved original packages and DCK consumers have separate entry points.


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

The earlier production migrations preserve Vectorballs' reflection and Second
Reality's indexed lens. Twelve Union intro/pointer-trail captures match exactly
before and after their extraction. Generic GPU magnifier tests compare against
an independent pixel model; they do not imply the analytic effect is identical
to every historical indexed lens.

FR-010 now shares its float32 spline camera tracks while keeping specialized
part renderers. Second Reality shares plasma, indexed scatter and palette
operations while retaining other software-rendered parts and ST3 synchronization.
Cuddly/Union keep scene choreography, assets and artistic constants. The
systematic extraction section above records what is shared and what remains.


Earlier measurements on the Pixel 10a with the original water/path/lens effects
laboratory composition (900
updates, first 60 excluded; all effects in that composition enabled). These
figures do not measure the newly added saved-project scene or software kernels:

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


### Saved-project scene measured on Pixel 10a

The new authored scene was also measured on the Pixel 10a on 2026-09-23:
1,800 updates, first 60 excluded, approximately 29 seconds of measurement.
It combines a repeated background, 24 image sprites with beat-grid modulation,
mixed-font scrolling, six timed geometry modes and a title. No live audio analysis
or procedural plasma is included in this particular measurement.

| Logical resolution | Observed draws/s | Observed updates/s | Mean Update CPU | Mean Draw submission CPU | Logical RGBA storage |
| --- | --- | --- | --- | --- | --- |
| 640×360 | 59.94 | 60.01 | 46.8 µs | 1.69 ms | 1.86 MiB |
| 320×180 | 59.94 | 59.98 | 45.7 µs | 1.70 ms | 0.54 MiB |

Both runs held approximately 60 updates/s. Lower resolution reduced logical image
storage; CPU submission time was similar in this composition. The complete app
allocated about 60 MB over each 29-second run, including engine/runtime activity;
this is not a zero-allocation application or a GPU/power/thermal measurement.
Automatic glyph culling runs after geometry mapping and preserves the checked
rendered pixels; custom painters and manual compatibility transports keep their
own clipping policy.

Android accepts `--ez authoring true`, `--ez eco true` and `--el frames 1800`
when starting the effects-lab activity. Bounded runs save their report and resume
continuous playback automatically. During a bounded run the activity may display
over the keyguard without dismissing it. `--el frames 0` starts ordinary continuous
playback. Desktop uses `go run ./examples/effectslab -authoring`; add `-eco`,
`-frames` and `-profile` to choose a bounded measurement.

## Build a demo from complete components

A complete component includes its state, controller, geometry or text layout,
and rendering resources. For example, DMA Is Back now constructs
`effects.NewJellyCube(effects.DMAJellyCubeConfig())` and updates/draws that instance;
its DCK source no longer implements cube modes, deformation, face sorting or
triangle submission. Lower-level `geometry`, `motion` and `composite` functions
remain available for effects with different behavior.

| Component | Complete shared behavior | Production consumers |
| --- | --- | --- |
| `scrolling.Atlas` + `presets.FontAtlas` | Font metrics, cached glyph images, case/alias/fallback rules and ribbon layout | DMA Is Back, TeamG1, Megatwist, Coco, Multiscreen, Nonameno, Grodan |
| `BitmapRecipe`, `BitmapText`, `BitmapParagraph` | Fractional atlas sampling, cached character lookup and paragraph alignment | Cuddly screens, Union screens/menu/loaders |
| `effects.JellyCube` | Five-mode controller, entrance, deformation, continuous handoffs, projection and rendering | DMA Is Back; `examples/jellycubes` |
| `effects.SolidCube` / `SolidCubeBatch` | Material, culling, face ordering, outlines and bounded batch submission | Bilizir, Coco, Multiscreen Coco |
| `effects.TexturedCube` | Live texture mapping, camera, rotation, face ordering and culling | TeamG1 |
| `effects.PerspectiveCheckerboard` | Perspective stripe geometry, two-axis motion, XOR composition and bounded surfaces | 3D DOC and Cuddly 3D DOC |
| `effects.ProjectedBallTrain` | Blended movement programs, projected sprites/shadows, phase and depth/palette ordering | 3D DOC and Cuddly 3D DOC |
| `sprites.ProjectedField` | Bounded spawn/respawn, strict/wide wrap, projection, history and pixel/sprite/vector materials | Nonameno stars, Union Starballs, Cuddly Starwars and Big Sprite |
| `scrolling.Config.Crawl` | Paragraph window, vertical transport and perspective projection | Cuddly Starwars |
| `scrolling.Config.Feed` | Finite glyph insertion into a cached scrolling trail | DMA Is Back, Coco, TeamG1, MegaTwist intros |
| `scrolling.Config.Scanline` | Proportional text transport, cumulative wave, cyclic bounce and bounded strip rendering | DMA Is Back, Coco, MegaTwist main screens |
| `scrolling.Config.Profiled` | Independent proportional-text and floating-profile clocks with clipped strip sampling | TeamG1 main scrolling |
| `scrolling.Config.RowColumn` | Fixed-cell bitmap text, sampled source rows and independent destination columns | DMA 3D and Replicants |
| `scrolling.Config.RowBands` | Circular bitmap text, ordered row displacement passes and a final crop | 3D DOC intro and main screen |
| `composite.ProfileImage` | Cached source rows, editable displacement table, motion phase and finite wrap copies | TeamG1 logo |
| `plasma.HarmonicImage` | Harmonic kernel, reusable CPU pixels, live GPU surface and dirty-frame upload | TeamG1 plasma |
| `sprites.Group` with `CircleFormation` | Indexed circular poses, secondary harmonic motion and independent sprite scales | TeamG1 twelve-logo formation |
| `effects.TimedCRTOverlay` | Adjustable time-varying scanlines, glow, color fringe and flicker | TeamG1 intro |
| `scrolling.Config.Bands` | Cached repeated text, independent lanes and bounded viewport rendering | Cuddly Spreadpoint |
| `scrolling.Config.Slots` | Glyph recycling, wave motion, tangent orientation and custom poses | Cuddly Reset |
| `scrolling.Reveal` | Cached text layout and ordered per-character entrance | Union loader |
| `composite.Bands` | Independently moving/repeated background strips and batched drawing | Union Multiplane |

`FeedConfig.ProgressiveEntry` keeps the active glyph at the viewport edge until
its full advance has entered, revealing its bitmap over successive updates.
The default retains the earlier one-frame glyph insertion used by other intros.
TeamG1 selects progressive entry at the 768-pixel right edge and a flat CRT
projection, keeping the shader's scanlines, chromatic fringe, glow and flicker
without clipping the top or bottom strokes of its 72-pixel-high font.

For concrete integration, see the constructors in `demos/dma-is-back/dck/game.go`,
`demos/teamg1-demo/dck/game.go`, `demos/go-cocoisthebest/dck/main.go`,
`demos/go-multiscreen/dck/main.go`, `demos/go-cuddlymenu/dck/screens/starwars.go`,
`demos/go-uniondemo/internal/loader/loader.go` and
`demos/go-uniondemo/internal/screens/b_multiplane.go` in the surrounding workspace.
Original source versions remain alongside them. This consolidation excludes
FR-010 and Second Reality; their previously shared components remain available,
and their source was not modified by this change.

The [effect map](EFFECT_MAP.md) lists every screen in the other native DCK
catalogs, the complete components already used there and the mechanisms still
worth extracting. It distinguishes production data from an effect's controller
and rendering code so shorter adapters keep the intended screen behavior.

### Reuse the finite intro and sampled text ribbon

Both variants use `scrolling.New`, with font dimensions, character order and
fallback behavior supplied by `scrolling.Atlas`. A finite intro can expose its
current image to a logo warp or CRT pass:

```go
feed := presets.DMAIntroFeed(atlas, introMessage)
intro, err := scrolling.New(scrolling.Config{Feed: &feed})
if err != nil { return err }
// Advance once per simulation tick. intro.Finished() signals the last glyph.
err = intro.Update(frame)
crt.DrawAt(screen, intro.Image(), 0, 164)
```

The proportional scanline renderer owns its glyph window, cumulative wave
sampling, character cursor, bounce and triangle buffers:

```go
sampled, err := presets.DMAScanlineScroll(atlas, message)
if err != nil { return err }
sampled.WaveStep = 12
sampled.BounceAmplitude = 24
scroll, err := scrolling.New(scrolling.Config{Scanline: &sampled})
if err != nil { return err }
err = scroll.Update(kit.Frame{Tick: tick})
scroll.Draw(screen)
```

Choose source height, strip height, visible cursor rows, clock step, bounce,
background color, destination offset, missing-glyph policy and one of three X
sampling rules independently. `ScanlineSplit` draws explicit wrapped quads;
`ScanlineAddressRepeat` delegates repetition to the texture sampler;
`ScanlineReject` discards partially out-of-range source windows. An optional
`DisplacementProgram` gives the scroll a one-time introduction before its loop.
`UseTime` accepts a variable-speed scene clock; default playback uses `Tick`.
All modes reuse bounded surfaces and triangle arrays. DMA and Coco use the same
wave recipe but different strip sizes, image formats, clocks and wrap rules;
MegaTwist adds the introductory wave and strict source-window rejection.

The shared `effects.CRTOverlay` accepts curvature, scanline, chromatic and
vignette parameters. `composite.ImageGrid` draws finite, independently spaced
copies of any borrowed image, including overlapping logo tiles. DMA positions
that grid with `motion.NestedOrbit`; changing the atlas or orbit does not
change the scroll or cube component.

TeamG1 uses four more complete components: `Config.Profiled` for its text and
floating row table, `composite.ProfileImage` for its wrapped logo lines,
`plasma.HarmonicImage` for a live plasma surface, and `sprites.Group` with a
`motion.CircleFormation` for twelve independent logos. Their recipes expose
font metrics, source window, profile samples, row and text speeds, logo phase,
sprite count, scale harmonics and placement. The plasma image renders CPU pixels
and uploads them only after its clock changes. Reusing one image in several
layers therefore does not repeat its pixel work. The TeamG1 adapter now supplies
the actual images, messages, composition order and soundtrack timing.
The intro's `effects.TimedCRTOverlay` also owns its shader and simulation
clock. The animated material accepts independent scanline, glow, color-fringe,
curvature and flicker settings, so another intro can use the same pass with a
different bitmap image or selected source surface.

`sprites.Group` also supports a centered grid plus a shared multi-harmonic
translation. Coco uses these parameters for sixteen logos and selects
`AlphaOnly` to preserve its bright RGB values while fading alpha. The reusable
group owns the phase and prepared poses; a user speed control calls
`group.Advance(delta)` once per update. The number of images, grid steps,
harmonic terms, image bank, scale and material can be varied independently.

On mobile, `plasma.HarmonicConfig.ColorLookupSize` may replace the standard
four-wave kernel's per-pixel color trigonometry with a bounded RGB table. Zero
keeps exact coloring. At 16,384 entries the source plasma differs from exact
output by at most one channel level in the regression samples, without frame
allocations. TeamG1 also enables `composite.ProfileImageConfig.Batch` on mobile
to submit its sampled logo rows together. The desktop game keeps its exact
backends. Both switches are explicit options; they do not change the effect's
clock or layer order.

Short unlocked Pixel 10a SurfaceView present-timestamp samples of the actual
DCK APKs (63 frames each) measured 59.92 FPS for DMA Is Back, 59.91 for Coco
and 59.91 for MegaTwist, with no interval above 25 ms. TeamG1's exact DCK
backend measured 53.07 FPS with eight intervals above 25 ms; its combined
mobile options measured 59.92 FPS in two later samples, with no such interval.
The 320×200 CPU color kernel measured 453 µs/frame exact versus 116 µs/frame
with the lookup on Apple M4 Max; this CPU-only comparison does not identify
which mobile change produced the end-to-end improvement. The phone samples
cover visible main screens and do not measure long-run thermal or battery use.

The row/column family shares another complete scrolling transport:

```go
atlas, err := presets.FontAtlas("dma-3d", fontImage)
if err != nil { return err }
config := presets.DMA3DRowColumn(atlas, message)
config.ColumnAmplitude = 48
scroll, err := scrolling.New(scrolling.Config{RowColumn: &config})
if err != nil { return err }
err = scroll.Update(frame)
scroll.Draw(screen)
```

Change the font, message, advance, row table, row height, column width, source
bias, text speed, independent column speed, amplitude and output placement.
`RowColumnQuads` batches both passes; `RowColumnImages` keeps separate source
crops for the other authored sampling topology. `SetTransportMultiplier` lets
input or a music signal change speed without resetting the text or waves.
The two work images remain bounded by the configured viewport rather than
message length. DMA 3D and Replicants use different crop and clock presets,
but the same controller, glyph compilation and two-pass renderer.

3D DOC uses the same `scrolling.New` entry with `Config.RowBands`. Its intro
has no image passes and exposes `CursorRune()` for the scene cue. Its main
message applies two ordered row-displacement passes and a final crop. Each
pass has an independent lookup phase, strip thickness, output size and
vertical motion. RGB complete-frame checks cover 13 frames of 3D DOC,
including its intro-to-main handoff. Replicants matches 21 captures,
including two user-speed changes. DMA 3D matches 13 of 15 frames exactly;
the other two differ by one channel level in a single pixel at their respective
sample times, including the old text reset boundary.

### Put a perspective floor under another scene

`PerspectiveCheckerboard` owns the floor surfaces, projected stripe geometry and
two-axis movement. The default 320 by 80 recipe is editable: change counts,
stripe coordinates, horizon, focal length, motion periods, palette, opacity and
destination position. `Position` can replace the built-in oscillator with any
deterministic Go trajectory. Separate mask XOR and direct per-band XOR are
selectable because they produce different antialiased edges. Clock order and
wrap policy are explicit when a scene needs exact historical phase boundaries.

```go
floorConfig := presets.DOCCheckerboard()
floorConfig.X, floorConfig.Y = 80, 220
floorConfig.Color = color.RGBA{R: 40, G: 120, B: 210, A: 255}
floorConfig.Opacity = .75
floorConfig.Position = func(tick uint64) (x, y float64) {
    return 8 * math.Sin(float64(tick)/30), float64(tick) * .4
}
floor, err := effects.NewPerspectiveCheckerboard(floorConfig)
if err != nil { return err }
defer floor.Close()
// Call floor.Update(frame) once per logical tick, then floor.Draw(layer).
```

`ProjectedBallTrain` borrows one ball image and a shadow palette. Its count,
camera, center, scale, anchors, shadow plane, palette quantization, painter
order and spin rate are editable. `Programs` contains movement functions and
`Sequence` selects two program indices plus a blend factor for each slot.
Call `AdvanceAt(sceneSeconds)` when the scene owns a clock, or `Update(frame)`
for the configured fixed step. `PoseAt` samples a transition frame without
advancing rotation. Both effects have independent state and can be layered or
instantiated repeatedly. The DOC and Cuddly presets keep their distinct
movement loops and shadow order. Full RGB captures match 13/13 standalone
frames and 12/12 Cuddly frames through late playback.

### Load a font without initializing characters in the demo

The atlas image remains an application asset. Its reusable metrics and lookup
rules come from DCK. This function produces a regular scrolling effect directly:

```go
func NewMessage(image *ebiten.Image, text string) (*scrolling.Scrolling, error) {
    atlas, err := presets.FontAtlas("dma-is-back", image)
    if err != nil { return nil, err }
    return scrolling.New(scrolling.Config{
        Text: text,
        Fonts: map[string]scrolling.Face{"main": atlas.Face()}, Font: "main",
        X: 640, Y: 180, Speed: 120, Gap: 96, Repeat: true,
    })
}
```

The proportional alphabet, glyph rectangles, spacing and blank entries live in
one recipe. Other presets include `teamg1-demo` (with its special logo glyph),
`grodan-up`, `nonameno-small`, `bilizir-demo` and
`tcb-multi-plane-3d-scroller`. `presets.Fonts()` lists the integer atlas families.
For a new font, supply `font.NewGrid` or explicit `font.Config` metrics to
`scrolling.NewAtlas(image, metrics)`. Those metrics can use any character order,
per-glyph advance/bearing, aliases, fallback and supported blank characters.

`atlas.Glyph(r)` resolves normal case/alias/fallback rules;
`atlas.ExactGlyph(r)` preserves a literal atlas lookup. Both reuse cached image
views. `atlas.Layout(text, scrolling.AtlasText{Vertical: true})` advances by line
height; `Literal` and `SkipMissing` retain authored transport conventions when
needed. Supply the returned glyphs to `scrolling.Config.Glyphs`. Regular mixed
fonts, control commands and modes still use `Text` / `Fonts` on the same constructor.

Fractional cells use a separate image recipe to avoid rounding their source
sampling. The recipe is editable before it is bound to an image:

```go
func NewFractionalFont(image *ebiten.Image) (scrolling.BitmapGrid, error) {
    recipe, err := presets.BitmapRecipe("cuddly-bigsprite")
    if err != nil { return scrolling.BitmapGrid{}, err }
    // The preset retains 83.25-by-41 cells and a fractional column span.
    // Set Width, Height, Columns, First or Order here for another atlas.
    return recipe.Grid(image, ebiten.FilterNearest)
}
```

`presets.BitmapFont(id, image, filter)` combines those two calls. Use
`scrolling.NewBitmapText` for cached static/individually animated glyphs,
`NewBitmapParagraph` for aligned lines, or a recycled scrolling configuration for
fractional scrolling cells. `grid.Scrolling(text)` also creates the common scroll
renderer when its cells have integer dimensions. `presets.CuddlyChromeAlphabet()`
provides an editable variable-width tile alphabet; `CuddlyChromeTiles(text)`
compiles its established menu recipe.

Font regressions cover all entries of ten original metric maps and the complete
byte-range lookups of seven original character mappers. `cmd/checkfontpixels`
checks cached GPU glyph crops against actual assets: 1,147 drawable glyphs across
24 integer-atlas catalog entries. These checks concern glyph geometry/pixels; scene-level
comparisons separately exercise animation and composition.

### Place several complete animated cubes

Each `JellyCube` owns its pose, mode controller and rendering buffers. Configuration
selects behavior; the caller supplies the clock and draw order:

```go
func NewCubePair() (kit.Group, error) {
    config := effects.DefaultJellyCubeConfig()
    config.X, config.Y, config.Size = 160, 180, 52
    left, err := effects.NewJellyCube(config)
    if err != nil { return nil, err }

    config.X = 480
    config.Phase, config.Speed = 3, 0.8
    config.Deformation.Twist = 1.8
    config.Steps = []effects.JellyCubeStep{
        {Mode: effects.JellyBounce, Duration: 3},
        {Mode: effects.JellySwing, Duration: 4},
        {Mode: effects.JellyNormal, Duration: 3},
    }
    right, err := effects.NewJellyCube(config)
    if err != nil { left.Close(); return nil, err }
    return kit.Group{left, right}, nil
}
```

Call the group's `Update(frame)` once per simulation update, `Draw(dst)` in layer
order and `Close()` when it is no longer needed. `kit.NewLayers` adds start times,
durations and fades to either cube, with `LocalTime: true` for an entrance relative
to its layer start. `cube.SetPosition(x, y)` can follow a path or music-driven
callback without restarting the animation. Use `kit.Func` when connecting such a
callback: set the position in `OnUpdate`, then call `cube.Update(frame)`; draw the
same instance in `OnDraw`, and close it explicitly when the wrapper is discarded.

The five modes are normal, tumble, pulsate, swing and bounce. Their durations and
order are independent of the wobble, ripple, squash, twist and translation gains;
a zero gain disables its deformation family. `Transition` controls pose handoff,
`Delay` and `ZoomDuration` control the entrance, and six colors select the palette.
`DMAJellyCubeConfig` adds the authored entrance to the default configuration.
**JellyCube `Size` is half an edge; SolidCube and TexturedCube `Size` are full edges.**

The clock uses absolute `Frame.Time` in seconds. `Speed` scales that time and
`Phase` offsets the complete animation, including its sequence. A deterministic
60 Hz internal controller keeps mode boundaries consistent across display rates.
Backward seeks replay the controller; `MaxReplaySteps` bounds work per update
(default 100,000 steps). Handle an Update error for a seek beyond that budget.
Ordinary updates and mode changes reuse their Go buffers; this does not assert
zero allocation for Ebitengine or the complete application.

Try the self-contained three-cube example, which supplies configuration only:

```sh
go run ./examples/jellycubes
go run ./examples/jellycubes -save /tmp/cubes.json -frames 600
go run ./examples/jellycubes -project /tmp/cubes.json
go run ./examples/jellycubes -eco -frames 600 -capture /tmp/cubes.png
```

Its JSON uses the same authoring schema as scrolls, sprites and backgrounds. For
example, this layer can be added to a project's `layers` array without assets:

```json
{
  "id": "cube", "kind": "jelly_cube",
  "window": {"start": 2, "fadeIn": 0.5}, "localTime": true,
  "jellyCube": {
    "preset": "dma-is-back", "center": {"x": 320, "y": 180},
    "halfEdge": 52, "phase": 0, "transition": 0.75,
    "steps": [{"mode": "normal", "duration": 3}, {"mode": "bounce", "duration": 4}],
    "deformation": {"twist": 1.5, "translation": 0}
  }
}
```

Omitted cube fields inherit the selected preset; explicit zero values are retained.
The JSON compiler constructs `effects.JellyCube`; it contains no duplicate cube
controller. A GUI can edit these data fields and use the same validation/compiler.
The editor itself is not included.

### Reuse solid materials, live cube textures and text entrances

`presets.BilizirCube(size)`, `CocoCube(size)` and `MultiscreenCocoCube(size)` return
independent `SolidCubeConfig` values. They retain the palettes, face order, culling
and outline conventions of their consumers; every field remains editable. Keep
an instance's `Rotation` continuous, or call `Rotate(dx, dy, dz)` during updates.
For many cubes, construct `effects.NewSolidCubeBatch(capacity)` once, then
`Reset()`, `Add(cube, x, y)` in drawing order and `Draw(dst)` per frame. `Add`
returns false when capacity is exceeded. The batch preserves object insertion
order; it does not depth-sort faces across separate cubes. It owns its buffers
and white texture, while the added cubes remain caller-owned.

A textured cube accepts a live image, such as a plasma surface, scroller or logo:

```go
func NewLiveCube(surface *ebiten.Image) (*effects.TexturedCube, error) {
    config := presets.TeamG1TexturedCube()
    config.X, config.Y = 320, 180
    config.AngularVelocity = geometry.Vec3{X: 0.6, Y: 0.9, Z: 0.3}
    return effects.NewTexturedCube(surface, config)
}
```

Render the texture's contents before drawing the cube. `Update(frame)` samples
absolute-time rotation; `Rotate` is the alternative for caller-controlled fixed
steps. Select one clock policy. UVs, filtering, depth order and culling are
configurable; mapping is affine per triangle. `Close()` releases the reference
but leaves the borrowed source image alive.

Cuddly's perspective text is also a complete scrolling configuration:

```go
func NewPerspectiveCredits(atlas *ebiten.Image, lines []string) (*scrolling.Scrolling, error) {
    grid, err := presets.BitmapFont("cuddly-starwars-crawl", atlas, ebiten.FilterNearest)
    if err != nil { return nil, err }
    config, err := presets.CuddlyStarwarsCrawl(grid, lines)
    if err != nil { return nil, err }
    config.PixelsPerUpdate = 0.5
    return scrolling.New(scrolling.Config{Crawl: &config})
}
```

The preset expects at least 30 lines, including blank padding. Font, paragraph
alignment, visible-line count, speed, output crop and placement are editable.
Replace `Projection` with `composite.NewRowProjection(rows)` for a different
perspective; the row projector also accepts arbitrary live images. The crawl's
two working surfaces are bounded by its viewport, not by the total message length.
Its speed is **pixels per Update**; advance it once at the selected simulation TPS.
A standalone `scrolling.Crawl` also offers `SetPosition(line, offset)` for seeking.

For the Union loader entrance, `scrolling.NewReveal(presets.UnionCreditsReveal(
font, lines))` compiles all glyph positions. Customize `Order`, `Delay`, `Duration`,
`FromX`/`FromY`, `UniformStartY` and an optional `Ease(progress)` function before
construction. Call `DrawAt(dst, time)` in any consistent time unit. The historical
preset's delay/duration use its original counter units; convert them when driving
it with seconds. Changing the supplied time can seek the entrance without changing
its final layout or creating textures.

For independently scrolling scenery, `composite.NewBands(presets.UnionMountainBands())`
constructs the complete mountain-strip renderer. Change each crop, velocity,
phase, wrap period, placement and motion scale in the returned `BandsConfig`;
`CopyOffsets` controls repetitions. Call `Step()` once per simulation update and
`DrawAt(dst, atlas, x, y)` during drawing. Velocities are phase units per Step.
The renderer reuses bounded geometry without copying the source image.

Repeated text lanes use the same scrolling constructor:

```go
config := presets.CuddlySpreadpointBands(grid, message)
config.Lanes[0].Speed = 4
scroll, err := scrolling.New(scrolling.Config{Bands: &config})
```

Each lane has its own position, speed and phase. Configure the viewport,
entrance, repetition and final image filter before construction. This preset
uses `Frame.Tick`; a plain `BitmapBandsConfig` uses seconds unless `UseTicks`
is enabled. The renderer owns one fixed viewport image and submits batches of
at most 256 glyph quads. A long message increases cached text data, never texture
width. This bounds GPU memory by the viewport; it does not imply that a viewport
image is smaller than every short text strip.

For moving and rotating individual glyphs, use another transport selection:

```go
config := presets.CuddlyResetSlots(grid, message)
config.Count = 12
config.Period = float64(config.Count) * config.Advance
config.Waves[0].Amplitude = 25
scroll, err := scrolling.New(scrolling.Config{Slots: &config})
```

`BitmapSlotsConfig.Map` can change position, angle, scale and opacity per glyph;
`PreviousTangent` optionally aligns each glyph with the previous visible point.
Each `Update` advances the fixed-step transport once. `Draw` never advances or
recycles characters, so repeated draws keep the same pose. Both `Bands` and
`Slots` support the common `Output` passes for whole-image deformation,
reflections or magnification. A standalone `BitmapSlots.SetSpeed` changes speed
without resetting its positions or wave phases.

`BitmapText.DrawWindow(dst, x, y, width)` provides cached visible-range rendering
when a scene owns the scroll clock. A destination subimage selects an interior
viewport; text coordinates remain absolute. Union TNT2 uses this instead of
creating substrings during drawing.

Keep assets alive for every borrowing atlas/effect. Close JellyCube, SolidCube,
SolidCubeBatch, TexturedCube, Crawl and BitmapBands (or their owning group/layers)
when finished. Atlas, BitmapText, BitmapSlots, Reveal, RowProjection and
`composite.Bands` allocate no owned GPU image
requiring Close. Construct fonts, paragraphs, recipes and geometry buffers outside
Update/Draw. Callback code should reuse its storage and follow the same rule.
Existing Pixel measurements above concern their named scenes; new combinations
need their own measurement, especially when adding full-resolution surfaces.

## Reusable construction presets

`presets` contains editable effect recipes and shared asset metrics. Messages,
asset bytes, screen layout and scene order stay in the demo. A preset returns
independent data; modifying it never changes another effect instance.

The Bilizir and TCB deformation tables now use `motion.CompileWaveTable`:

```go
sections := presets.BilizirWaveSections()
sections[0].Terms[0].Amplitude *= 1.5
sections[1].Samples = 240
samples, err := motion.CompileWaveTable(sections...)
if err != nil {
    return err
}
wave := composite.ProfileStrips{
    Offsets: samples, Axis: composite.Rows, Thickness: 2, Speed: 1,
}
```

Call `wave.Advance()` once per simulation update and `wave.DrawAt(dst, image, x,
y)` during drawing. The image may be a text surface, logo or complete background.
Each section contains a sample count, offset and any number of sine/cosine terms.
Frequency and phase are radians per sample and radians respectively. Empty terms
produce a hold; section order determines the program. `TCBLogoWaveSections`
supplies the other authored recipe. Tables are compiled once, outside drawing.
`BilizirCopperOffsets` also supplies the exact quantized copper table shared by
Bilizir, Coco and Multiscreen, with independent storage for each caller.

The multi-plane scrolling recipe uses the common constructor:

```go
config := presets.TCBProjectedScroll(message, 32, face, raster)
config.Projected.PixelsPerUpdate = 4
config.Projected.Draw = scrolling.PlaneDraw{ScaleX: 2, ScaleY: 2}
scroll, err := scrolling.New(config)
```

Its returned projection, visible-slot count, eight forms, font, raster, placement
and output passes can all be changed. This preset retains the authored `^0` to
`^7` command timing and two control slots. For ordinary text with no historical
slots, use `Controls` and `Modes` on the regular scrolling configuration.

`effects.SolidCube` provides a reusable flat-colored, outlined cube. It caches its
geometry buffers, depth-sorts whole faces, and renders their outlines in the same
order. Palette, edge colors, edge width, size, camera distance and depth order are
configured independently:

```go
config := presets.BilizirCube(20)
config.EdgeWidth = 2
cube, err := effects.NewSolidCube(config)
```

Keep `cube.Rotation` in XYZ radians or call `cube.Rotate(dx, dy, dz)` each update;
render with `cube.DrawAt(dst, centerX, centerY)`, then release its owned texture
with `cube.Close()`. `effects.DefaultSolidCubeConfig` supplies a neutral material.
Use `effects.NewTexturedCube` for live cube textures and `effects.NewMesh` for
arbitrary textured/deformable objects. The fixed cube
geometry path uses zero Go allocations after construction; this is not a claim
about allocations inside the graphics engine or the whole demo.

`scrolltext.FontProgram` compiles font selection once when several synchronized
surfaces show different scales or materials of a single message:

```go
program, err := scrolltext.NewFontProgram(message, scrolltext.DomSizes, "0")
smallText := program.MaskedText("0", ' ')
largeText := program.MaskedText("3", ' ')
activeFont := program.FontAt(visibleGlyphIndex)
```

Masked texts preserve all visible glyph positions, substitute blanks in inactive
banks and remove control bytes. `FontAt` uses a zero-allocation binary search;
its index counts Unicode glyphs, not bytes. Pass another decoder, such as `Braces`,
for a different syntax. Timing/effect controls are rejected by this specialized
program; the regular scrolling constructor handles mixed controls and mixed fonts.


### Three independent cubes on Pixel 10a

The complete three-cube example was measured for 1,800 updates at each resolution,
with the first 60 excluded (about 29 seconds). All five mode families and their
transitions are exercised, with independent phase, speed and palette settings.

| Logical resolution | Draws/s | Updates/s | Mean Update CPU | Mean Draw submission CPU | Logical RGBA images |
| --- | --- | --- | --- | --- | --- |
| 640×360 | 59.94 | 59.98 | 77.7 µs | 626.2 µs | 1.76 MiB |
| 320×180 | 59.95 | 59.99 | 76.5 µs | 276.3 µs | 0.44 MiB |

These are observed application frame rates and CPU submission timings, not GPU
execution, battery or thermal measurements. Both runs held about 60 updates/s.
The whole application allocated about 3.1 MB over each 29-second measurement,
including Ebitengine/runtime and the debug labels. The cube core's zero-allocation
benchmark is narrower than this application-level measurement.

The shared Android host selects this scene with `--ez cubes true`. Bounded runs
use `--el frames 1800`, save a report and automatically resume animation.
`--el frames 0` runs continuously; the normal lifecycle pauses it while the
phone is locked/backgrounded and resumes it when the activity becomes visible.
