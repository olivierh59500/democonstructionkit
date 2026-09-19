# democonstructionkit

A Go construction kit for Ebitengine demoscene productions, developed from the
19 demos in the sibling `demos/` directory. Effects accept explicit assets, font
metrics and motion parameters; production scripts remain ordinary Go.

The library includes 23 source-atlas descriptions, reusable effect families,
YM/module audio adapters and 19 runnable composition recipes. Recipes are effect
studies, **not faithful ports** of the full productions. Original timelines, scene
data, artwork and musical synchronization remain part of each production.

## Run

Requires Go 1.26+ and Ebitengine's platform dependencies. Versions match the local
source demos: Ebitengine 2.9.11, ym-player `3f73bdca82e5`, go-zikmu `b245427b8556`.

From this repository:

```sh
# Self-contained example; all graphics are created in Go.
go run ./examples/minimal

# Reuse the original local assets, without copying them into this repository.
go run ./cmd/gallery -demos ../../demos -demo bilizir-demo -audio
go run ./cmd/gallery -demos ../../demos -demo go-multiscreen
go run ./cmd/gallery -demos ../../demos -demo go-secondreality -audio
go run ./cmd/gallery -list

# Optional external module/YM music.
go run ./cmd/gallery -demo teamg1-demo -music /path/to/music.xm

# Render a contact sheet at a fixed animation time.
go run ./cmd/gallery -demo all -time 3 -capture captures/gallery.png
```

Space pauses the gallery, Escape exits. Music is optional; `-audio` uses the
selected recipe's original track. `-music` accepts a standalone YM/MOD/XM/S3M/IT.
The Second Reality example extracts and normalizes its original Purple Motion
S3M from `Reality.FC` before passing it to go-zikmu.

## Compose a demo

```go
metrics, err := font.NewGrid(font.Grid{
    Bounds: atlas.Bounds(), Cell: image.Pt(32, 33), Columns: 10,
    Order: " ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!?", Uppercase: true,
})
if err != nil { return err }

scroll, err := effects.NewScroller(atlas, metrics, effects.ScrollConfig{
    Width: 640, Height: 400, Message: "HELLO DEMOSCENE! ",
    Speed: 120, Gap: 64, Y: 300, Scale: 1,
    Wave: motion.Waves{{Amplitude: 24, Spatial: .02, Speed: 2}},
})
if err != nil { return err }

game, err := democonstructionkit.NewGame(
    democonstructionkit.Group{background, logo, scroll},
    democonstructionkit.Config{Width: 640, Height: 400, TPS: 60},
)
if err != nil { return err }
defer game.Close()
ebiten.SetTPS(60)
return ebiten.RunGame(game)
```

For irregular/proportional atlases, use `font.New` with a `map[rune]font.Glyph`.
`Rect`, `Advance`, `OffsetX`, `OffsetY` and `LineHeight` are independent. Grid
descriptions accept NUL holes, gutters, margins, aliases and a fallback character.
See [font presets](presets/fonts.go) and the [complete minimal example](examples/minimal/main.go).

## Building blocks

| Package | Responsibility |
| --- | --- |
| `font` | Immutable bitmap metrics and Unicode text layout; no Ebitengine dependency |
| `motion`, `geometry`, `timeline` | Waves/tables, easing/keyframes, 3D math/clipping/morphing, clocks and scene timing |
| `outline` | TrueType/OpenType curves flattened into reusable vector text |
| root package | `Effect`, `Group`, `Sequence`, `Viewport`, `Game` and custom callbacks |
| `effects` | Text, scrolling, strip/grid warp, masks, rasters, sprites, tiles/rotozoom, tilemaps, stars, meshes, wireframes, vectorballs, reflection, plasma, tunnel, ripple, lens, palette pixels and CRT |
| `render`, `assets` | Bounded triangle batches, persistent surfaces and `fs.FS` asset caching |
| `sound` | Device-independent stereo PCM for ym-player and go-zikmu |
| `sound/ebiten` | Playback using the application's single Ebitengine audio context |
| `presets`, `recipes` | Source font/music metadata and compositions for all 19 repositories |

All animation uses seconds, radians and pixels. Convert original frame-based
increments explicitly (e.g. 2 pixels/tick at 50 Hz becomes 100 pixels/second).
`Update` receives an absolute local time; `Draw` renders the current state. Do not
share one mutable effect instance between two independently timed scenes.

`Warp` is the common mechanism for scanline scrollers, column sine waves, DNA
twists and text planes. `MeshEffect.Deform`, `PointCloud.Shape`, sprite paths,
keyframe tracks and pixel callbacks keep artistic choices outside shared code.

## Audio contract

Create one `audio.Context`, then pass it to `sound/ebiten.NewPlayer`. Sample rates
must match. Construct device playback during the first Update for mobile startup.
Streams emit stereo float32 little-endian PCM and accept arbitrary read lengths.
Fixed decoder blocks preserve YM output across different callback sizes.

PCM seeking uses byte offsets and replays from the beginning when moving backward;
it is accurate but expensive. Prefer `Player.Seek(time.Duration)` when a device is
attached, because it also resets buffered playback. Device `Position()` measures
audible progress; stream `Position()` measures bytes consumed by the device.

YM loops use the decoder's loop mode. YM EOF has decoder-block granularity.
go-zikmu currently exposes neither end-of-song nor order/row markers publicly:
module `Duration` is optional, but **required with `Loop: true`**. Without it, the
stream follows the upstream replay engine until closed. The kit does not guess
song endings from silence or claim exact Second Reality musical synchronization.

## Verify

```sh
go test -race ./...
go vet ./...
go run ./cmd/checkassets -demos ../../demos
go run ./cmd/checkaudio -demos ../../demos
go run ./cmd/gallery -demos ../../demos -check
GOOS=js GOARCH=wasm go build -o /tmp/dck-minimal.wasm ./examples/minimal
```

Ebitengine tests and `gallery -check` require a working graphics session (or an
appropriate virtual display on Linux). Pure packages and source asset/audio checks
can run without a display:

```sh
go test -race ./font ./motion ./geometry ./timeline ./outline ./presets ./sound
```

The gallery check samples every recipe at seven times, verifies nonuniform and
animated output, and compares repeated Draw calls. It is not an image comparison
against the original demos. Source audio checks decode both Second Reality tracks
and the 18 YM assets; they do not assert replay fidelity for an entire song.

See the source inventory, migration guide,
validation notes and commit progress.

Original assets stay in their repositories and retain their authorship and
licenses. This repository does not bundle the source artwork or music.
