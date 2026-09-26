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
| `Pseudo3D` | Independent text banks with reverse-order harmonic X/Y/depth poses, clipping and mirrored scale | `PixelsPerUpdate` per bank; shared tick or caller time |
| `Sliced` | Bounded history of glyph strips; a control consumes a transport step; rotation is independent | `SlicesPerUpdate`; `RotationSpeed`: film frames/second |
| `Crawl` | Cached paragraph window projected through configurable rows; bounded surfaces independent of text length | `PixelsPerUpdate`; zero selects manual `SetPosition` on a standalone `Crawl` |
| `Bands` | Independently moving repeated text lanes in a fixed viewport | Lane `Speed`: pixels/second, or pixels/tick with `UseTicks` |
| `Slots` | Recycled bitmap glyphs with waves, optional tangent rotation and a custom pose mapper | `Speed`: pixels/Update |

Choose one transport. `Page` belongs to the regular transport. For the other
transports, configure text, fonts, speed and repetition inside that transport rather
than also setting regular `Text`, `Fonts`, `Speed` or `Repeat`. `Projected.Draw`
and `Sliced.Draw` hold their output placement. A fixed-step setting of 2 at
60 TPS advances 120 pixels/second; changing TPS changes this authored cadence.

Four pseudo-3D text lanes can share one transport without assuming the same
message, character order or viewport. `presets.VivaPseudo3D` supplies editable
bank positions, inverted depth scales, sine/cosine rates, pixel snapping,
clipping and alpha:

```go
pseudo := presets.VivaPseudo3D(face, nil, [4]string{text1, text2, text3, text4}, 768, 540)
pseudo.Banks[2].BaseY = 390
pseudo.Pose.XAmplitude = 40
scroll, err := scrolling.New(scrolling.Config{Pseudo3D: &pseudo})
if err != nil { return err }
controller := scroll.Pseudo3DController()
controller.SetBankSpeed(0, 2) // A cue may change one lane without resetting it.
scroll.Update(frame)
scroll.Draw(screen)
```

Set `Atlas` instead of `Face` for an existing pre-sliced alphabet. `UseFrameTime`
uses the caller's music or timeline position; otherwise `TicksPerSecond` and
`TimeOffset` define a deterministic clock. `SetTransportMultiplier` pauses or
speeds all lanes, while the controller exposes per-bank speed, position and
global phase cues. Viva TCB and its Multiscreen panel render the same four-bank
pose program with different texts and viewport sizes; eight captures per
version through frame 4,800 match their preceding images in all channels.

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

For several simultaneous bitmap scrollers, configure independent `RingConfig`
values in `RingLanesConfig` and still enter through `scrolling.New`. Fullscreen
uses seven fonts/messages lanes with a shared vertical shift every 128 ticks;
`ShiftFirst` fixes the first movement boundary, and a `motion.WrapLimit` recycles
each lane without moving the text itself:

```go
laneConfig := scrolling.RingLanesConfig{
    Rings: ringConfigs, Y: []float64{-28, 52, 132, 212, 292, 372, 452},
    ShiftEvery: 128, ShiftFirst: 130, ShiftVelocity: []float64{2},
    ShiftUpper: &motion.WrapLimit{Boundary: 540, Restart: -28, Inclusive: true},
}
scroll, err := scrolling.New(scrolling.Config{RingLanes: &laneConfig})
scroll.Update(frame)
scroll.Draw(textLayer)
```

Use `scrolling.NewRingLanes` directly when each ring needs its own destination:
Big Sprite draws two synchronized fonts into separate masks with `DrawLaneAt`.
The component owns both glyph transports and lane positions; drawing never
advances either clock, and no message-width image is allocated.

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
- `sprites.Group` with `motion.CuedFormation`: staggered horizontal, vertical,
  arcing and looping motion cues for a row or column of arbitrary images.
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

For a sprite phrase, put the glyph images in text order and use `FrameStride: 1`.
`Origin` and `Spacing` define the resting row. Each cue can choose which item
leads (`LeadIndex`), how long the motion lasts (`Duration`), the delay between
neighbors (`Stagger`), independent X/Y sine harmonics, and a fade at its ends.
Several cues may overlap and `Loop` repeats the complete sequence. For example:

```go
formation, err := motion.NewCuedFormation(motion.CuedFormationConfig{
    Origin: motion.Point{X: 160, Y: 230}, Spacing: motion.Point{X: 32},
    Count: len(letters), Loop: 12,
    Cues: []motion.FormationCue{{
        Start: 0, Duration: 2, LeadIndex: len(letters)-1, Stagger: .08,
        Y: []motion.FormationHarmonic{{FirstAmplitude: 60, LastAmplitude: 60, Cycles: 1}},
    }},
})
group, err := sprites.NewGroup(sprites.GroupConfig{
    Frames: letters, Count: len(letters), FrameStride: 1,
    Formation: formation.At, Speed: 1,
})
group.Update(kit.Frame{Time: seconds})
group.Draw(screen)
```

For a chain of independently phased sprites, use a data-only
`motion.HarmonicFormationConfig`. Each X/Y term selects sine or cosine, either
of two phase clocks, an index phase and an optional shared amplitude envelope.
The group owns both clock advances and a bouncing envelope when configured:

```go
formation, err := motion.NewHarmonicFormation(presets.GrodanSpriteFormationConfig())
group, err := sprites.NewGroup(sprites.GroupConfig{
    Frames: frames, Count: len(frames), FrameStride: 1,
    ScaleX: 2, ScaleY: 2, Harmonic: formation,
    HarmonicClockStep: [2]float64{.02, .03},
    HarmonicEnvelope: &motion.BounceBankConfig{
        Start: []float64{0}, Velocity: []float64{.1}, Min: -50, Max: 50,
    },
})
group.Update(kit.Frame{})
group.Draw(screen)
```

Change `Origin`, `Spacing`, individual term amplitudes/rates/phases or optional
`Bounds` without rewriting the sprite loop. `SetHarmonicState` accepts external
phase or music values instead of the group's fixed clock steps. MegaTwist uses
the same formation component with four time harmonics, two index-only ripples
and clipping; its `sprites.GlowPainter` draws the prepared group with editable
halo layer count, scale, opacity and filter. `DrawGroup` never advances motion
and does not create a full-screen surface.

For nonuniform letter spacing in phase rather than pixels, set `IndexOffsets`
and `UseIndexOffsets` on the affected terms. A term may use `Divisor` instead
of `Rate` to retain exact authored division. Union Beat Dis combines a shared
horizontal wobble with eight differently phased letter orbits. Cuddly LED uses
nine authored phases, a vertical cosine and one bouncing amplitude. Both draw
their prepared group before advancing it, preserving their original first frame.

`sprites.AxisFlip` provides a separate, composable front/back material for a
logo or sprite. It uses one configurable `motion.BounceBank` lane as a signed
vertical scale, switches images at `SwitchAt`, and applies independent face
angles. Draw it on any trajectory or into a mask; `Pose` exposes the selected
image and transform for a custom renderer:

```go
flip, err := sprites.NewAxisFlip(sprites.AxisFlipConfig{
    Front: front, Back: back, SwitchAt: .01, BackAngle: 180,
    Motion: motion.BounceBankConfig{
        Start: []float64{1}, Velocity: []float64{-.02},
        Min: -1, Max: 1, Inclusive: true, Directional: true,
    },
    Filter: ebiten.FilterLinear, Blend: ebiten.BlendSourceOver,
})
if err != nil { return err }
position := orbit.At(phase)
flip.DrawAt(canvas, position.X, position.Y)
flip.Step() // Draw-before-step keeps the first fully visible front frame.
```

`Back` may be omitted to flip a single image. `DrawAt` centers whichever face
is selected, so differently sized art stays centered. The effect reuses its
images and creates no intermediate surface. Cuddly Big Sprite uses this
controller for its two-face emblem. `SnapCenter` truncates each selected image's
half dimensions to integer pixels, as Big Sprite requires for its 346×143 and
347×144 faces. Sixteen channel-aware captures around both switches and bounds
match the previous screen exactly.
For a scale that jumps from a strict upper/lower bound and alternates faces on
each jump, set `Saw` to a `motion.SawToggleConfig` instead of `Motion`. Union
Multi-Plane uses `Start:0`, `Velocity:.08`, `Boundary:1`, `Restart:-1`; it calls
`Step` before `DrawAt` to retain the first frame's .08 scale. The front/back
images and their angles remain editable. Eleven native captures around its
face changes match the previous screen pixel for pixel.
`UseAnchor` overrides centering in source pixels; `BackMirrorY` mirrors a single
source image on alternate cycles, with an editable vertical shift.
`DrawAtWith` applies a parent `ebiten.GeoM` after the local pose, for example
the Multi-Plane screen's 2× viewport transform. `OptionsAt` exposes that same
geometry for a custom material. The autonomous TCB screen and its Multiscreen
panel now use these settings; eleven and ten captures respectively match their
previous image exactly.

For a repeated logo or sprite whose X motion multiplies two harmonics and Y
motion adds two, `sprites.RecurrentFormation` shares the image and advances
four sine/cosine seeds across all instances. Count, per-instance phase steps,
time divisors, sine/cosine choice, native origin, output scale and opacity are
editable. `Update` samples once for the chosen tick or music phase; repeated
`Draw` calls reuse poses without trigonometry or an intermediate surface:

```go
recipe := presets.VivaLogoFormation(logo, 768, 540, 51.5)
recipe.Count = 12
recipe.XSecondaryStep = 1.0 / 48.0
logos, err := sprites.NewRecurrentFormation(recipe)
if err != nil { return err }
logos.Update(float64(frame.Tick))
logos.Draw(screen)
```

Standalone Viva and its Multiscreen panel both use this component; the latter
sets its authored vertical amplitude to 37.5. Eight captures per version
through frame 4,800 match the preceding image in every channel.

For repeating image strips, `sprites.NewTrain` owns both the image group and
its X/Y motion. Each axis can be fixed, use a phase-spaced `motion.Wave`
(`Cos: true` selects cosine), or use a `motion.BounceBankConfig`. The bounce
bank accepts one or several velocities and defaults to the one-step overshoot
seen in classic raster effects; `Inclusive` and `Clamp` are optional. For
example, a complete three-raster train needs only its assets and parameters:

```go
bars, err := sprites.NewTrain(sprites.TrainConfig{
    Images: []*ebiten.Image{pink, green, brown}, ScaleX: 390,
    Y: sprites.TrainAxis{Offset: 60, Bounce: &motion.BounceBankConfig{
        Start: []float64{94, 124, 154}, Velocity: []float64{2}, Min: 94, Max: 160,
    }},
})
bars.Update(kit.Frame{Tick: tick}) // Once per simulation update.
bars.Draw(screen)                 // May be drawn in several layers.
```

Image dimensions, image order, scale, blend and layer placement remain
independent of motion. The train reuses its pose and motion slices.
`motion.Wave` also supports `Offset` and `Rectify` for an absolute-sine or
absolute-cosine bounce. Use a `motion.WaveClock` when the phase advances in
simulation ticks and can change speed or reset at a text/timeline cue:

```go
clock, err := motion.NewWaveClock(presets.RectifiedSine(340, -60, .08))
if err != nil { return err }
// In Update, sample before advancing to keep the authored first frame.
scrollY := clock.At(0)
clock.Step()
// Draw the scrolling surface, logo or sprite at scrollY.
// A control event can call clock.SetStep(.04), clock.SetPhase(0) or clock.Reset().
```

Pass a glyph/sprite index to `At(index)` and set `Wave.Spatial` for phase
spacing. The same clock can drive several layers in phase; construct another
clock with a different `Start`, `Step` or `Wave.Phase` for a deliberate offset.
Sampling and stepping allocate no memory per frame. Cuddly Digi, LED,
Megaball and Ehhh share this clock with different amplitudes and cue rates;
Starwars uses the same rectified wave to build one segment of its row profile.
`WaveClockConfig.HoldTicks` delays phase advances without changing the initial
visible pose; `SetHold` can schedule another pause. `presets.VivaTitleMotion`
drives the same horizontal cosine title path with a 970-tick hold in standalone
Viva and zero hold in Multiscreen. The title image is supplied separately to
`RasterTitle.DrawAt`; fourteen captures per version around the hold and raster
wrap boundaries remain identical in every channel.
Set `Directional` when an axis should reverse only while moving outward. With
`AllowOutsideStart`, an image can enter from beyond its normal bounds before
settling into a bounce. Cuddly Mega Scroller uses this pair for its masked text
surface: start at 45, move upward by 2 per tick, then rebound between -70 and
20. The default boundary behavior remains unchanged for existing raster banks.

`sprites.Atlas` caches a row-major image bank once and exposes either borrowed
sub-images or absolute source regions. Negative indices clamp and positive
overflow wraps; cell width, height and atlas dimensions are editable. Pair it
with `sprites.FrameSequence` to choose an arbitrary frame order from scene or
music time:

```go
atlas, _ := sprites.NewAtlas(sprites.AtlasConfig{
    Image: sheet, TileW: 32, TileH: 32,
})
walk, _ := sprites.NewFrameSequence(sprites.FrameSequence{
    Duration: .35, Indices: []int{2, 3, 4, 5, 6, 7, 8, 9}, Loop: true,
})
screen.DrawImage(atlas.Tile(walk.Current(seconds)), nil)
// For a fractional-source renderer, use atlas.Region(walk.Current(seconds)).
```

Cuddly Menu shares one atlas component among its map, character and logo
tiles. Disk Copier requests LCD regions from the same component while keeping
its original region-sampling renderer and independent operation clocks.
For independently timed sprite or atlas lanes, `motion.GatedWrapBank` adds one
activation threshold per lane to the existing wrapped transport. A strict or
inclusive gate, velocity, initial phase and boundary are editable; `Frame(i)`
floors the current phase for atlas lookup. Disk Copier starts three LCD tile
clocks after its read, format and write cues, then wraps each strictly after
frame 82. Call `StepAt(sceneTime)` after drawing when the first visible tile
must use the phase from the preceding tick.

For formations that multiply waves rather than simply add them,
`motion.FormulaExpr` provides editable time, index, viewport-radius and count
inputs with ordered arithmetic, sine and cosine operations. DCK validates and
compiles each expression once into a bounded numeric program. The
`sprites.FormationCarousel` then selects modes, places the borrowed atlas
images, runs entry/exit slides and snaps anchors without per-frame surfaces:

```go
config := presets.CuddlyMenuCarousel(carebearAtlas)
config.Hold, config.Slide = 8, 1
carousel, _ := sprites.NewFormationCarousel(config)
carousel.Draw(screen, musicOrSceneSeconds)
```

The preset exposes all seven Cuddly Menu trajectories as data. Each expression
can be replaced or edited independently, including its per-sprite phase and
viewport-relative radius. Repeated Draw at the same time remains deterministic.
On an M4 Max, evaluating all 84 poses from the seven menu modes takes about
4.6 µs with zero allocations, versus 1.58 µs for the seven fixed Go functions;
one menu frame evaluates one mode of 12 poses. This kernel measurement does not
include drawing or establish Pixel 10a performance.

For repeated decor that should cost one draw rather than hundreds of tiles
per frame, `composite.CachedTileParallax` builds a bounded viewport-plus-overscan
surface once. Camera divisors, wrap periods, integer quantization, negative
clamping, blending and tile spacing are independent settings:

```go
config := presets.CuddlyMenuTileParallax(tile, 768, 400, 32)
backdrop, _ := composite.NewCachedTileParallax(config)
backdrop.DrawAt(screen, cameraX, cameraY)
defer backdrop.Close()
```

The Cuddly menu keeps its original 800 × 432 unmanaged surface, half-speed
integer camera motion and one source-copy draw per frame. Another logo or
scrolling layer can use the same renderer with its own tile, size and phase.
`motion.CameraFollow` separately keeps an object at an editable viewport anchor
until the camera reaches a world edge. It returns camera and on-screen object
coordinates together, with independent X/Y world extents. Cuddly Menu rounds
its actor position first, then samples this motion policy for the map camera;
door selection and input remain application decisions.

Union Menu composes smaller independent controllers: `motion.WrapBank` loops
the panorama; `motion.LinearTick` closes the uncover wipe; two
`timeline.PacedIndex` instances choose palette and walking frames with distinct
first-step cadences; `motion.HoldBounce` waits, squeezes and restores its logo.
The image atlas still belongs to the production, and the menu keeps input and
door navigation. Each preset exposes its speed, bounds, hold and frame counts;
all five controllers sample without per-tick allocation.
`motion.WalkParallax` adds independent signed-input speeds and directional
wrap rules for the hall and banner. `Advance(direction)` moves both layers on
one walking action; `Set(layer, position)` keeps the background aligned when
the menu places the character at a selected door. A restart may land exactly
on the opposite bound without triggering a second wrap in the same action.

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

For any two PNG capture trees, run
`go run ./cmd/compare-frames -reference /path/to/before -candidate /path/to/after`.
It checks every RGB and alpha channel and exits unsuccessfully when pixels or
frame sets differ. `-diff /path/to/diffs` writes amplified mismatch images.
Checking only an RGBA difference image's bounding box can miss changes to RGB
when both frames have the same alpha; use this channel-aware comparator.

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

For an automatically scrolling layer, set `VelocityX` and/or `VelocityY` in
destination pixels per second. The velocity is added to `Pose` or to the pose
returned by `Sample`; a custom sampler remains available for bounce, camera
paths and music-linked movement. `Update` receives the logical frame time, so
display refresh does not alter the scroll speed.

```go
renderer, err := composite.NewBackground(composite.BackgroundConfig{
    PeriodX: 640, PeriodY: 400,
    CopiesX: 3, CopiesY: 2,
})
if err != nil { return err }
backdrop := &composite.BackgroundLayer{
    Renderer: renderer, Image: tile,
    VelocityX: -48, VelocityY: 12,
}
// Include backdrop in a kit.Group or kit.Layers and call Update once per tick.
```

`CopiesX/Y` limit the source indices to `[0, count)` on each repeated axis.
Leave either count zero for an infinite repeat. Finite counts retain their
authored origin even after long camera movement; they can preserve intentionally
empty edge regions and gaps between tiles. Grodan uses three columns and two
rows with periods different from its 640 × 398 source size. Cuddly Fullscreen
uses an unbounded 16-pixel strip and `VelocityX` to replace its local scroll
counter. Their migrated outputs match 1,200 and 480 baseline video frames,
respectively.

The same finite effect can be saved for an editor or loaded from JSON:

```json
{"kind":"background","background":{"image":"tile","period":{"x":640,"y":400},"copiesX":3,"copiesY":2,"velocity":{"x":-48,"y":12}}}
```

The `tile` ID is resolved by the host application. The compiler checks copy
counts and worst-case visible submissions before allocating render resources.

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

### Keep mutable backdrop offsets within authored wrap rules

`motion.NewWrapBank` shares the stateful part of repeated scenery. Each layer
has a start offset and velocity; optional lower/upper rules choose their own
boundary, restart position and strict or inclusive comparison. This preserves
historical entrance frames and exact reset ticks while `Background.DrawAt`
continues to own cropping, repetition and blending:

```go
offsets, err := motion.NewWrapBank(motion.WrapBankConfig{
    Start: []float64{-640, -640, -640}, Velocity: []float64{-2, -4, -6},
    Lower: &motion.WrapLimit{Boundary: -640, Restart: 0},
    Upper: &motion.WrapLimit{Boundary: 0, Restart: -640},
})
offsets.Step()
for i, image := range layers {
    background.DrawAt(screen, image, offsets.At(i), 0)
}
```

`SetVelocity`, `AddVelocity` and `ReverseAll` change speeds without moving the
current phase, so controls can affect each parallax layer independently.
Calling `Step` before or after drawing selects the production's original
update boundary. The bank uses no surface and allocates nothing per step.
Set `WrapLimit.Relative` when crossing a boundary should add a period to the
overshooting position instead of jumping to a fixed restart coordinate. The
editable `presets.VivaRasterWrapConfig` and `presets.TCBMountainWrapConfig`
keep the corresponding standalone and Multiscreen bands synchronized without
duplicating their initial offsets, velocities or wrap rules.

`composite.RasterTitle` combines two moving raster phases, a repeated third
strip and a title image under one `Step`/`DrawAt` boundary. The returned preset
is editable: source crops, signed scale, phase offset, clipping, fill color and
blend are independent of artwork and the caller's horizontal trajectory.

```go
config := presets.VivaRasterTitleCanvas(title, raster)
config.ThirdOffsetY = 96
material, err := composite.NewRasterTitle(config)
if err != nil { return err }
defer material.Close()
material.Step()
material.DrawAt(screen, titleX, 14)
```

`VivaRasterTitleCanvas` owns only a 528×36 surface for the standalone source.
`VivaRasterTitleDirect(title, raster, 800)` draws into Multiscreen's 800×72
clipped region and allocates no intermediate surface. Fourteen capture pairs
per version, including both raster wraps and the title's departure cue, match
the preceding renderers in every color and alpha channel.

### Animate a repeated texture with a rotozoom

`composite.RotozoomBackground` owns the update/draw boundary and renders one
repeat-addressed quad. Set a base `Repetition` and `RotozoomVelocity` for a
simple moving, zooming or rotating tile. A `RotozoomProgram` supplies a complete
pose when the animation has stages or music cues. The source image is borrowed;
each effect instance can use its own center, zoom, rotation and texture phase.

```go
program, err := presets.NewVivaRotozoom(
    presets.DefaultVivaRotozoomConfig(768, 540))
if err != nil { return err }
roto, err := composite.NewRotozoomBackground(
    composite.RotozoomBackgroundConfig{Image: tile, Program: program})
if err != nil { return err }
// Call roto.Update(frame) once per logical tick, then roto.Draw(screen).
```

The Viva preset exposes its entrance speed and travel, stage thresholds,
independent orbit/zoom/rotation rates, amplitudes and phase offsets. Its default
settings match all 1,200 compared frames across the entrance and rotozoom.
`composite.Repeat` remains available when an existing controller already owns
the complete pose.

Simple rotozoom layers can also be saved for a future editor. This layer moves
its center and rotates the texture without any Go callback:

```json
{"id":"tile","kind":"rotozoom","rotozoom":{"image":"tile","center":{"x":384,"y":270},"zoom":1.5,"rotationVelocity":0.3,"centerVelocity":{"x":-12,"y":4}}}
```

Second Reality's rotozoomer uses a different sampling backend: 256 × 256
palette-indexed textures, signed fixed-point steps and 16-bit address wrapping.
`indexed.Rotozoom256` retains that behavior without RGBA conversion or per-frame
allocation. The two backends share the rotozoom effect family and configurable
pose/phase concepts, while preserving their different pixel sampling rules.
Three reference poses, including the rotated-source branch and wrapped
coordinates, match the original indexed pixels exactly. A separate 43-second
capture of the Rotozoomer also matches all 2,580 decoded frames of the
unmodified DCK production.

```go
indexedRoto, err := indexed.NewRotozoom256(
    indexed.Rotozoom256Config{Width: 160, Height: 100})
if err != nil { return err }
if err := indexedRoto.Render(dst, picture, rotatedPicture, x, y, xa, ya); err != nil {
    return err
}
```

### Bend a backdrop row by row

`composite.ScanlineBackground` combines a displacement program, a bounded
horizontally tiled source and a batched row renderer. The foreground scroller
can use a different program and clock. `WaveStep`, `WaveDivisor`, `BaseX/Y` and
the bounce amplitude/rate are independent; `Sample` can replace the source-row
map for a custom effect.

```go
background, err := composite.NewScanlineBackground(
    composite.ScanlineBackgroundConfig{
        Tile: tile, Program: wave,
        Width: 416, Height: 276,
        SurfaceWidth: 672, SurfaceHeight: 64,
        BaseX: 80, WaveDivisor: 2, WaveStep: 5,
        BounceAmplitude: 30, BounceRate: .1,
        AlternateDiagonal: true,
    })
if err != nil { return err }
// Update(frame) samples the rows once; Draw(screen) submits one bounded batch.
```

The MegaTwist migration matches 4,800 decoded frames, covering the intro,
transition and animated main background. It uses one 672 × 64 source surface
instead of an image spanning the full message or viewport history.

### Fill a scene or logo with animated copper bars

`composite.CopperBars` owns two independent phase clocks, the displacement
table, source-strip cache and bounded renderer. Set the number of bars, source
period, row spacing, phase increments, sample spacing and horizontal shift
without copying a demo's draw loop. `CopperQuads` batches large overlapping
rasters; `CopperImages` retains individual DrawImage sampling for short title
fills. `MaskedClock` uses integer power-of-two phase wrapping, while
`SingleWrapClock` keeps fractional speed controls.

```go
bars, err := composite.NewCopperBars(presets.BilizirCopperBars(
    image, 72, composite.CopperImages, composite.SingleWrapClock))
if err != nil { return err }
if err := bars.SetSpeed(1.5); err != nil { return err }
if err := bars.Update(frame); err != nil { return err }
bars.Draw(titleCanvas)
```

Bilizir uses the same table on 300 batched full-screen bars; Coco and its
Multiscreen panel use 36 cached image strips under their independently moving
titles. The migrations match 1,200, 3,600 and 1,200 decoded frames. Draw never
advances the phases or allocates a message-sized texture. A saved project can
store the table and every clock/geometry setting as a `copper_bars` layer;
the host still chooses the image asset and layer order.

### Animate a raster over live text or a logo

`composite.RasterOverlay` draws a borrowed image onto an existing surface with
an explicit blend mode. `BlendSourceAtop` preserves the destination's text or
logo alpha; `BlendSourceIn` applies the incoming raster through it. The effect
owns source cropping, scale, angle, opacity, phase, velocity and exact reset
thresholds. `Draw` never advances the clock, and `Step` can happen before or
after drawing to retain a screen's initial frame. No extra GPU surface is
allocated beyond the composition's existing text or logo canvas.

```go
raster, err := composite.NewRasterOverlay(composite.RasterOverlayConfig{
    Image: colors, ScaleX: 85, ScaleY: 1, Alpha: 1,
    VelocityY: -2,
    WrapY: &composite.RasterWrap{Boundary: -177, Restart: 0, Inclusive: true},
    Filter: ebiten.FilterLinear, Blend: ebiten.BlendSourceAtop,
})
if err != nil { return err }
// Draw text into its existing transparent surface, then color and advance it.
raster.Draw(textSurface)
raster.Step()
```

Cuddly Big Sprite and Starwars use different blend modes and opposite boundary
rules; Union Wow and Replicants use negative and positive phase velocities.
Their migrations match 1,800 decoded Cuddly frames and 13 exact Union captures,
including both wrap boundaries. `effects.Mask` remains useful when color and
alpha are independent effects that need separate working surfaces. In a saved
`authoring` project, a `raster_overlay` layer must draw directly after its
alpha source, with no outer fade; its own `alpha` setting remains editable.

`RasterOverlayConfig.Copies` draws several copies from the same phase without
advancing it between draws. `effects.NewMaskWith` composes that material with
any alpha effect using a chosen blend, alpha offset, output placement and an
optional cleared top band. DOM uses the same reusable pieces for its three
moving raster copies and four-size scrolling text:

```go
raster, err := composite.NewRasterOverlay(presets.DOMRasterCopies(colors))
if err != nil { return err }
masked, err := effects.NewMaskWith(presets.DOMScrollMask(raster, scroll))
if err != nil { return err }
if err := masked.Update(frame); err != nil { return err }
masked.Draw(screen)
defer masked.Close()
```

The mask owns two bounded working surfaces and its input effects. A logo,
scrolling surface or whole scene can supply the alpha image; the raster copies,
phase and Porter-Duff blend remain editable independently.

For a background assembled from vertically sampled source strips,
`composite.VerticalStripTrain` caches the source views once and owns the moving
source phase. Configure source stride and crop height separately from the
sample offset and destination spacing. Out-of-range strips are omitted, as in
DOM's background near its wrap boundary:

```go
background, err := composite.NewVerticalStripTrain(
    presets.DOMBackgroundStrips(backgroundImage))
if err != nil { return err }
background.Step()
background.Draw(screen)
```

The same component can sample scenery, raster images or a text surface. It
borrows the image, draws no intermediate full-screen surface and lets the host
place other effects above or below the sampled strips.

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

For sprites whose atlas frame advances at an independent rate per instance,
`sprites.AnimatedField` composes the pure `motion.FrameField` clock with a
borrowed image sequence. Each spawn can set position, fractional starting
frame and frame rate; a completed animation may respawn through the same
callback. The frame selector, count, offsets, speed multiplier, blend and
filter are independent options. DOM's eight animated stars are an editable
recipe:

```go
atlas, err := sprites.NewAtlas(sprites.AtlasConfig{
    Image: starSheet, TileW: 64, TileH: 46,
})
if err != nil { return err }
options := presets.DefaultDOMStarOptions(random.Float64)
options.Count = 12
config, err := presets.DOMAnimatedStars(atlas.Tiles, options)
if err != nil { return err }
stars, err := sprites.NewAnimatedField(config)
if err != nil { return err }
// Update once per simulation tick, then draw at the chosen layer.
if err := stars.Update(frame); err != nil { return err }
stars.Draw(screen)
```

The source recipe's initial frame may exceed its nine-frame lifetime; the
next update respawns that sprite, and out-of-range art is skipped until then.
The DOM spawn callback and per-tick state updates allocate nothing after
construction.
The same field can also move particles independently on X and Y. A signed
`motion.FrameAxisWrap` retains a single strict or inclusive boundary crossing
and may edit the particle in `OnWrap`, such as choosing a new height while
preserving horizontal overshoot. `ImageByParticle` selects a fixed material per
instance instead of sampling a frame clock. Replicants combines three editable
star layers with this mode:

```go
frames, err := sprites.NewSolidFrames(presets.ReplicantsStarMaterials())
if err != nil { return err }
options := presets.DefaultReplicantsStarOptions(random.Intn)
options.Layers[1].Speed = 6.5
config, err := presets.ReplicantsStars(frames, options)
if err != nil { return err }
stars, err := sprites.NewAnimatedField(config)
if err != nil { return err }
if err := stars.Motion().SetSpeedMultiplier(1.4); err != nil { return err }
```

The caller owns the three solid images. Count, speed, color, tile size, wrap
width, height range and speed multiplier can be changed independently; a
music or timeline cue may change the multiplier on the next update.

For logo art that must grow through discrete bitmap sizes, `sprites.NewScaleFrames`
pre-renders an editable size sequence once. `sprites.CoupledLogoPair` combines
two such banks with linked depth and Y motion; it chooses each frame from
depth and paints the farther image first. All phase steps, amplitudes, frame
counts, frame-selection bias/gain and visibility of frame zero are parameters:

```go
config := presets.ReplicantsLogoPair(repLogo, tcbLogo)
config.Motion.SecondaryYSin = 90
config.Secondary.Gain = 28
logos, err := sprites.NewCoupledLogoPair(config)
if err != nil { return err }
if err := logos.SetSpeedMultiplier(1.2); err != nil { return err }
if err := logos.Update(frame); err != nil { return err }
logos.Draw(screen)
defer logos.Close()
```

The image sources are borrowed; `Close` releases only the cached scales.
`motion.CoupledLogoMotion` can also drive another renderer without creating
bitmap frames. Its 5,000-tick test checks both depth orders and changes of
speed against the authored Replicants equations.

`composite.BlockReveal` reveals cached image cells in row-major or caller-
selected order. Its pure `motion.SteppedReveal` clock separates the reveal
cadence from a later handoff tick, so an image may remain fully visible before
the next scene begins. Replicants uses one 640×40 row every three ticks and
hands off at tick 100:

```go
recipe := presets.ReplicantsSplash(splashImage)
recipe.Timing.DoneTick = 120 // Hold the completed image longer.
reveal, err := composite.NewBlockReveal(recipe)
if err != nil { return err }
if !reveal.Done() {
    if err := reveal.Update(frame); err != nil { return err }
}
reveal.Draw(screen)
```

Cell dimensions, block order, initial visibility, cadence, blocks per step,
background color and output position are independent parameters. The source
image is borrowed; views are cached once and `Close` releases the cache.

The foreground pair in Replicants uses the existing `sprites.Train` component
with a rectified sine wave. A quarter-cycle index spacing makes the first
sprite follow cosine and the second follow sine; horizontal spacing, baseline,
amplitude and phase are independent values:

```go
config := presets.ReplicantsBouncingSprites(spriteImage)
config.Spacing.X = 520
pair, err := sprites.NewTrain(config)
if err != nil { return err }
if err := pair.Update(kit.Frame{Time: phase}); err != nil { return err }
pair.Draw(screen)
```

The host may feed an absolute simulation phase, music-driven phase or a
custom trajectory while keeping the same cached sprite images.

`scrolling.Config.SizeBank` composes several differently scaled atlases over
one message and one transport clock. Font controls select the active bank when
they reach an editable right-edge lookahead; the other bank offsets stay
aligned to the same reference position. Each bank has its own scale, cached
glyph strip and vertical repetition. It uses the same `scrolling.New` entry
point as ordinary horizontal, vertical and projected text:

```go
recipe, err := presets.DOMSizeBank(message, fontImage)
if err != nil { return err }
recipe.Layers[3].ScaleY = 10 // Edit the largest font independently.
scroll, err := scrolling.New(scrolling.Config{SizeBank: &recipe})
if err != nil { return err }
bank := scroll.SizeBankController()
if err := bank.SetSpeedMultiplier(2); err != nil { return err }
if err := scroll.Update(frame); err != nil { return err }
scroll.Draw(textLayer)
```

`motion.ScaledTextClock` is the image-free controller behind the bank. It
accepts any number of scales, per-mode reference speeds, a control program,
strict or loose wrapping, and an initial offset. Its 20,000-tick unit test and
the DOM text's 100,000-tick parity test cover bank changes, speed controls and
wraps without a graphics device. Surfaces and glyph views are constructed
once; only the currently visible bank is redrawn each update.

### Animate staggered text pages

`motion.GlyphPageCycle` owns the per-character entrance, exit, page rotation,
delay-pattern rotation and stable depth order. `sprites.GlyphPages` draws that
state with any `scrolling.Atlas`; it can use literal characters or the font's
aliases and fallback. Text, grid size, spacing, center, target depth, duration,
easing and delay values are independent parameters. The three reusable delay
generators produce serpentine rows, mirrored columns and an inward spiral.
For Nonameno's exact dimensions and timing, use its editable preset:

```go
config, err := presets.NonamenoGlyphPages(pages, 0)
if err != nil { return err }
config.EnterDurationMS = 2400
config.DelayPatterns[0], err = motion.SpiralGlyphDelays(20, 8)
if err != nil { return err }
cycle, err := motion.NewGlyphPageCycle(config)
if err != nil { return err }
letters, err := sprites.NewGlyphPages(sprites.GlyphPagesConfig{
    Cycle: cycle, Font: fontAtlas, OffsetX: 12, OffsetY: 0,
})
if err != nil { return err }

// In Update, use the same elapsed simulation clock as the other effects.
if err := cycle.UpdateAt(elapsedMilliseconds); err != nil { return err }
// In Draw, this layer may be placed before or after any other DCK layer.
letters.Draw(screen)
```

`SetPageAt(page, pattern, nowMS)` is a timeline cue. Completion barriers start
each outgoing or incoming wave only after all glyphs finish. Construction
copies text and delay data; updates and stable depth sorting allocate nothing.
The Nonameno preset retains its source spiral's repeated delay value and the
original elastic equations. Its two page texts remain production data.

For a scrolling baseline with several independent waves, use
`scrolling.HarmonicSine` or `HarmonicSineWith`. The latter can restart the
spatial phase at each repeated text copy while leaving the time phase running.
`WaveVertical` and `WaveHorizontal` select the displacement axis; the result
is an ordinary `Mode` and can be combined with other mappers through `Chain`.
Nonameno's bottom ribbon is one editable recipe for that common renderer:

```go
config, err := presets.NonamenoBottomScroll(message, smallFont.Face(),
    presets.NonamenoScrollOptions{
        Width: 640, BaselineY: 442, TicksPerSecond: 60,
        PixelsPerTick: 1, Gap: 641,
    })
if err != nil { return err }
scroll, err := scrolling.New(config)
if err != nil { return err }
// One update per 60 Hz simulation tick, then scroll.Draw(screen).
if err := scroll.Update(kit.Frame{Tick: tick, Time: float64(tick)/60}); err != nil { return err }
```

The gap of 641 pixels matches Nonameno's strict original restart. Set `Gap:0`
for a continuous ribbon where the next copy follows the tail directly. Font
metrics determine the pen advance; `presets.NonamenoBottomWaves` returns the
two sine terms for editing their amplitude, frequency or direction.

`scrolling.Config.Ribbon` is a fixed-tick alternative for authored horizontal
or vertical scrolls with strict return thresholds. It uses the normal
`scrolling.New` constructor, cached atlas glyphs and a pure
`motion.RibbonClock`. Font advance determines the message length, while
`CullAdvance` may preserve a different historical width used only to choose
the first visible glyph. Grodan uses four independent recipes:

```go
recipes := presets.GrodanRibbons(messages, bigAtlas, upAtlas, smallAtlas)
recipes[1].Clock.Velocity = 4 // Change only the vertical lane.
scroll, err := scrolling.New(scrolling.Config{Ribbon: &recipes[1]})
if err != nil { return err }
if err := scroll.Update(frame); err != nil { return err }
scroll.Draw(verticalLayer)
```

Each ribbon independently configures direction, velocity, start and restart,
wrap threshold, font, scale, baseline and culling. `RibbonController` exposes
the live offset and speed for editor cues. Choose ordinary `Config.Repeat`
for a seamless loop instead of a historical exit-and-restart interval.

`composite.SurfaceLayer` composes one or more scrolling/image effects into a
single persistent surface, applies ordered raster passes and draws output
copies at editable positions and scales. It uses one surface per configured
layer, so Grodan's two small ribbons still share the same 320×32 target:

```go
recipes := presets.GrodanRasterLayers(scrolls, bigRaster, upRaster,
    smallTopRaster, smallBottomRaster)
layers := make([]*composite.SurfaceLayer, len(recipes))
for i, recipe := range recipes {
    var err error
    layers[i], err = composite.NewSurfaceLayer(recipe)
    if err != nil { return err }
}
if err := layers[2].Update(frame); err != nil { return err }
layers[2].Draw(screen)
```

Source effects, raster crops/transforms/blends, destination copies, clear or
feedback mode and ownership are independent choices. The host chooses when
each layer updates and where it sits among backgrounds, logos and sprites.

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
| Backgrounds and choreography | 73 identical PNG pairs across 11 screens; 10,260 identical Grodan/Cuddly/Viva/MegaTwist/Second Reality decoded frames; 292 GPU background cases; three exact indexed rotozoom reference poses |
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

`cmd/fidelity -reference <git-revision> -frames 0,1,60,240` compares a new
production commit to a chosen earlier commit instead of the pinned original.
This is useful for checking an effect extraction against its immediately
preceding renderer at update and wrap boundaries.

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

For finite intros that hand off to a main scene, `timeline.IntroHandoff` owns
the entry tick, a clamped per-update fade and one one-shot cue. The production
still supplies its text, music stream and scene layers:

```go
handoff, _ := timeline.NewIntroHandoff(presets.FadedIntroHandoff(.03, .1))
if !handoff.Main() {
    intro.Update(frame)
    handoff.Step(intro.Finished())
} else {
    handoff.Step(false)
    if handoff.CueReady() && player != nil {
        player.Play()
        handoff.MarkCue()
    }
    updateMain()
}
// Draw the main scene with handoff.Fade(); Draw never advances the cue.
```

The fade starts at zero on the entry tick and advances on the next main tick.
DMA Is Back and TeamG1 use the strict `fade > .1` cue; Coco selects
`presets.ImmediateIntroHandoff()` for opaque entry and same-tick music. The
configuration is data-only, so an editor can change the fade and cue threshold
without changing either renderer.
`IntroCueOnFirstMainTick` keeps a final intro frame visible after the state
changes and releases music on the following main update; Cuddly 3D DOC uses
this mode. The standalone 3D DOC screen selects an opaque handoff without a
music cue because its soundtrack already plays during the intro.

`timeline.CueClock` handles overlapping tick windows for loaders and other
finite transitions. A window starts accumulating before the end-of-frame tick
increment, preserving countdown and fade boundaries. `SetRate` changes 50/60 Hz
playback without resetting elapsed fades; `Reached(duration)` compares a fixed
recording duration by exact rational ticks. `timeline.Countdown` adds editable
first/second counts, a final hold and a fade lead without moving those formulas
back into an individual screen. Cuddly's sector/blipp loader uses both
components; Union's credits loader uses the same clock with its recorded
duration and a different bitmap reveal.

For staged visuals, `timeline.CueRanges` owns ordered time windows with explicit
open or closed endpoints. `timeline.SteppedEnvelope` selects an image/color bank
level during a stage's entry and exit. Disk Copier supplies its messages and
LED/LCD images while a DCK preset supplies the editable ranges and eight-step
palette program:

```go
ranges, _ := timeline.NewCueRanges(presets.UnionDiskCopierCueRanges())
fade, _ := timeline.NewSteppedEnvelope(presets.UnionDiskCopierFade())
if index, window, active := ranges.At(sceneTime); active {
    shade := fade.At(sceneTime, window.Start, window.End)
    drawStage(index, palette[shade])
}
```

Sampling performs no per-tick allocation. A gap returns `active=false`, so
the scene may retain or clear previous layers explicitly.
`timeline.EventStages` handles text, input or music events that change the
active part immediately. Later stage checks in the same update see that change;
there is no forced one-frame wait. Cuddly Reset combines its five named stages
with three `CueRanges` windows and two independent `HoldRamp` fades through
one `timeline.StageSequence`. Its `Window` is sampled before `StepWindow`,
retaining the first zero-opacity frame. Fonts, rasters, draw order and event
sources remain in the screen; stage names and fade durations are editable DCK
preset data.
`timeline.HoldRamp` covers a finite splash or interstitial: its configured
number of ticks reaches full progress, then the following Step reports exit.
An optional interior offset changes partial-frame opacity without moving the
zero or full endpoints. MegaTwist uses the editable 90-tick preset; the source
screen still chooses its images and which scene follows the hold.

```go
countdown, _ := timeline.NewCountdown(timeline.CountdownConfig{
    First: 49, Second: 162, Hold: 24, FadeLead: 42,
})
clock, _ := timeline.NewCueClock(timeline.CueClockConfig{
    Rate: 60, Windows: []timeline.CueWindow{
        {StartTick: countdown.BlankTick(), Duration: .25, Tolerance: 1e-9},
        {StartTick: countdown.FadeStartTick(), Duration: 1.5},
    },
})
clock.Step()
shown := countdown.At(clock.Tick())
volume := max(0, 1-clock.Elapsed(1)/1.5)
```

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
mesh. For separable sine/cosine motion, `composite.HarmonicCellWarp` accepts four
independent `motion.Waves` banks: X/Y displacement from row and column. They may
be combined and changed at a timeline cue with `SetWaves`:

```go
cells, err := composite.NewHarmonicCellWarp(composite.HarmonicCellWarpConfig{
    Cell: image.Pt(32, 16), Filter: ebiten.FilterLinear,
    Waves: composite.CellWaveBank{
        XRows: motion.Waves{{Amplitude: 32, Spatial: .3, Speed: .08}},
        YColumns: motion.Waves{{Amplitude: 16, Spatial: .3, Speed: .08}},
    },
})
if err != nil { return err }
defer cells.Close()
cells.DrawAt(dst, letters, kit.Frame{Tick: tick}, 70, 136)
```

`UseTime` selects seconds instead of ticks. The source image and destination
position are independent of the wave program. The Union intro's 512×224 image
contains 224 cells: its two original per-cell waves required 448 sine samples
per frame, while the shared renderer caches the same values for 14 rows and
16 columns, using 30 samples. Eight native captures remain pixel-identical.
Use the lower-level `CellWarp.Sample` for nonlinear or per-cell transforms.

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
| `motion.FrameField` + `sprites.AnimatedField` | Independent fractional atlas clocks or planar motion, single-edge wrap, ordered respawn and per-instance image selection | DOM animated stars, Replicants layered stars |
| `motion.CoupledLogoMotion` + `sprites.CoupledLogoPair` | Two linked logo paths, cached quantized scale banks, depth-based frame selection and draw order | Replicants paired logos |
| `motion.SteppedReveal` + `composite.BlockReveal` | Grid-cell reveal order, fixed tick cadence and independent completion hold | Replicants splash screen |
| `scrolling.Config.SizeBank` | Shared transport, controlled font-size cues, synchronized scaled offsets and repeated cached text layers | DOM four-size scroll |
| `motion.GlyphPageCycle` + `sprites.GlyphPages` | Font-independent staggered pages, editable delay grids, elastic depth motion, completion barriers and stable atlas rendering | Nonameno text pages |
| `scrolling.HarmonicSine` / `HarmonicSineWith` | Independent sine banks over the common text pipeline, optionally resetting spatial phase per repeated copy | Nonameno bottom scroll |
| `scrolling.Config.Ribbon` | Fixed-tick horizontal/vertical atlas transport, editable strict wraps, independent cull width and scale | Grodan four scroll lanes |
| `composite.SurfaceLayer` | One bounded canvas with ordered source effects, raster passes and repeated output placements | Grodan big, vertical and paired small scrolls |
| `scrolling.Config.Crawl` | Paragraph window, vertical transport and perspective projection | Cuddly Starwars |
| `scrolling.Config.Feed` | Finite glyph insertion into a cached scrolling trail | DMA Is Back, Coco, TeamG1, MegaTwist intros |
| `scrolling.Config.Scanline` | Proportional text transport, cumulative wave, cyclic bounce and bounded strip rendering | DMA Is Back, Coco, MegaTwist main screens |
| `scrolling.Config.Profiled` | Independent proportional-text and floating-profile clocks with clipped strip sampling | TeamG1 main scrolling |
| `scrolling.Config.RowColumn` | Fixed-cell bitmap text, sampled source rows and independent destination columns | DMA 3D and Replicants |
| `scrolling.Config.RowBands` | Circular bitmap text, ordered row displacement passes and a final crop | 3D DOC intro and main screen |
| `scrolling.Config.RingLanes` | Independent fonts and texts, synchronized slot updates, paced vertical motion and exact wrap | Cuddly Fullscreen and Big Sprite |
| `composite.ProfileImage` | Cached source rows, editable displacement table, strict phase wrap, parent viewport scale and finite wrap copies | TeamG1 and TCB/Union Multi-Plane logos |
| `plasma.HarmonicImage` | Harmonic kernel, reusable CPU pixels, live GPU surface and dirty-frame upload | TeamG1 plasma |
| `sprites.Group` with `CircleFormation` | Indexed circular poses, secondary harmonic motion and independent sprite scales | TeamG1 twelve-logo formation |
| `sprites.Group` with `HarmonicFormation` | Independent X/Y wave banks, two phase clocks, authored index phases, bounce envelope and optional bounds | Grodan sprite train, MegaTwist glowing logos, Union Beat Dis, Cuddly LED |
| `sprites.GlowPainter` | Configurable outer-to-inner halo layers and final image over prepared group poses | MegaTwist glowing logos |
| `effects.TimedCRTOverlay` | Adjustable time-varying scanlines, glow, color fringe and flicker | TeamG1 intro |
| `scrolling.Config.Bands` | Cached repeated text, independent lanes and bounded viewport rendering | Cuddly Spreadpoint |
| `scrolling.Config.Slots` | Glyph recycling, wave motion, tangent orientation and custom poses | Cuddly Reset |
| `scrolling.Reveal` | Cached text layout and ordered per-character entrance | Union loader |
| `composite.Bands` | Independently moving/repeated background strips and batched drawing | Union Multiplane |
| `composite.RotozoomBackground` | One tiled GPU quad with independent pose, phase, velocity or a staged motion program | Viva TCB |
| `indexed.Rotozoom256` | Allocation-free fixed-point rotozoom over indexed 256 × 256 textures | Second Reality Rotozoomer |
| `composite.ScanlineBackground` | Bounded horizontal tile source, independent wave/bounce clocks and batched source rows | MegaTwist |
| `composite.CopperBars` | Two-phase raster bank with editable table, clocks, source strips and quad/image materials | Bilizir, Coco, Multiscreen Coco |
| `composite.RasterOverlay` | Moving raster material with source crop, blend, scale, independent copies and exact wrap policy | Cuddly Big Sprite/Starwars, Union Wow/Replicants, DOM |
| `effects.Mask` / `NewMaskWith` | Two owned working surfaces, configurable alpha blend/offset, output placement and top crop | DOM raster-filled scrolling; reusable for logos and scene layers |
| `composite.VerticalStripTrain` | Cached source bands, independent sample/destination steps, moving phase and exact edge skipping | DOM scrolling scenery |
| `composite.RasterTitle` | Two-phase moving raster behind a title, optional small canvas or direct clipped draw | Viva TCB and Multiscreen Viva |

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

`composite.RowWarp` is the reusable strip engine behind those DOC passes.
`RowWarpDestinationX` moves whole source rows, retaining their bitmap edges;
`RowWarpSourceX` reads fractional horizontal positions from a borrowed image.
The latter powers Cuddly 3D DOC's paired inner and outer fonts. Each instance
has an editable lookup table, strip thickness, independent horizontal/vertical
clocks and filter, while the scene chooses its fonts, raster mask and layer
order. Call `DrawInto` before or after `Step` to preserve the source screen's
first-frame timing. The component allocates no image and can process a logo or
another live layer in the same way.

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

For profiles with holds, gaps or later writes that overwrite earlier samples,
use `motion.CompileWaveProgram`. `WaveAppend` appends a section; any nonnegative
`At` writes at an absolute index. `SampleStart` keeps a continuous sine phase
when only part of its range is written:

```go
program := presets.CuddlyDigiWaveProgram()
program[4].Section.Offset = -30 // Change one overlapping hold.
samples, err := motion.CompileWaveProgram(program...)
if err != nil { return err }
profile := composite.ProfileStrips{Offsets: samples, Speed: 1, Thickness: 1}
```

`presets.CuddlyIntroWaveProgram` and `CuddlyEhhhProfile` use the same compiled
wave sections. Ehhh reuses Digi's first 763 samples at half amplitude before
its own tail; `CuddlyEhhhTailWaveProgram` exposes that tail for variations.
All compilation happens once at setup, and profile drawing samples a finite
table without trigonometry or per-frame allocation. Thirty capture pairs
through late profile wraps in the three Cuddly screens match their previous
renderers pixel for pixel.
Union Multi-Plane also reuses `TCBLogoWaveSections`: setting `SampleStart` to
40 and 844 on its two sine sections preserves Union's original global-index
phases. Eight captures at the section joins and wrap match its previous image.
All three Multi-Plane screens now draw those rows through `composite.ProfileImage`:

```go
config := presets.TCBLogoRowProfile(samples, 303)
config.ScaleX, config.ScaleY = 2, 2
config.OutputX, config.OutputY = 64, 60
rows, err := composite.NewProfileImage(logo.SubImage(image.Rect(0, 16, 303, 48)).(*ebiten.Image), config)
if err != nil { return err }
rows.Advance() // The strict phase wrap is part of the effect.
rows.Draw(screen)
```

Union leaves scale at one and draws into its native stage. The source crop,
profile table, phase limit, native placement, output scale and viewport offset
are independent parameters. The preset's `RowStep: 1` advances the sine phase
for each scanline; without it the entire logo slides sideways instead of
undulating. A reusable quad batch draws the rows without an intermediate
full-screen image. Nine Union, ten standalone TCB and nine Multiscreen captures
now match the earlier undulating renderers in all color channels.

`effects.MultiPlaneScene` composes the mountain bands, row-profile logo,
front/back emblem and projected scrolling in one configurable effect. Supply
the images, text/font-backed `scrolling.Config`, source crops, wave table and
viewport. `NativeStage:true` retains Union's 320×200 stage and its two-pass
scaling; direct mode retains the standalone and Multiscreen clipped draws with
no intermediate full-screen surface. The host still chooses music, controls
and any additional layers. Call `Update` once per tick, then `Draw` as needed;
the component exposes its constituent effects for live editor cues.

```go
part, err := effects.NewMultiPlaneScene(effects.MultiPlaneSceneConfig{
    Mountains: mountains, Logo: logo,
    LogoSource: image.Rect(0, 16, 303, 48),
    CenterSource: image.Rect(114, 0, 193, 15),
    Bands: presets.TCBMountainBands(), Rows: config,
    Center: flipConfig, Scroll: scrollConfig,
    Viewport: image.Rect(64, 60, 704, 460), StageSize: image.Pt(320, 200),
    CenterX: 160, CenterY: 88,
})
if err != nil { return err }
defer part.Close()
if err := part.Update(frame); err != nil { return err }
part.Draw(screen)
```

For the standalone and Multiscreen TCB mountain backgrounds, use
`presets.TCBMountainBands()` with `composite.NewBands`. Its 32 moving crops
keep the original upper/lower strip order and copy offsets; `TruncatePhaseX`
quantizes each signed phase before its 2× screen displacement. Band geometry
is reused and the wrap remainder is calculated only when a boundary is crossed.
Thirteen standalone and twelve Multiscreen captures match the preceding
renderers pixel for pixel, including fractional speeds and late wraps.

Colorshock II provides an example of separating a trajectory from its source
samples. `presets.CuddlyColorshockOrbit()` returns a serializable X/Y formula
for the backdrop; edit its expression tree to change either frequency, radius
or center. `presets.CuddlyColorshockTableClock()` returns a strict `WrapBank`
that reads two position samples per tick and resets after index 1872. The
position table stays in the demo's data. Ten captures across its table wrap
match the previous screen pixel for pixel.

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
