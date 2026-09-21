# democonstructionkit

A Go/Ebitengine construction kit extracted from the actual productions in `demos/`.
The shared code must preserve their original artwork, text, lookup tables, timing,
pixel rounding, source crops, drawing order and blend operations.

**The first gallery was not a faithful reconstruction.** It has been moved to
`cmd/studies`. The current gallery launches the DCK versions located alongside
the preserved originals in each repository's `dck/` directory. Add `-original`
to run the original implementation.

The shared effects support configurable fonts, scroll modes, image deformation,
feedback ribbons, vectorball geometry, particle batches and continuous scene
handoffs. The Cuddly application combines fifteen native screens and its menu in
`demos/go-cuddlymenu/dck`.

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
script with a kit-generated approximation. Each source repository has a local
`go.mod` replacement pointing back to this module.

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

`Scrolling.DrawAt` also accepts original positions, visible ranges, circular text
windows, reverse drawing order and a glyph mapper. This lets an existing demo keep
its exact tick counters and reset conditions while sharing the renderer. Shader
and animated-strip backends use `DrawState.Paint`, retaining the same iteration
and layout pipeline.

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

**Second Reality:** only indexed palette expansion has been shared and compared
across renderer modes using a VRAM fixture. This does **not** establish complete
scene extraction or musical synchronization through go-zikmu.

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
