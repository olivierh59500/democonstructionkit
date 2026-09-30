# Effect templates

These patterns turn source artwork and timing into reusable DCK components.
They are ordinary Go code: images, fonts and music remain caller supplied. The
binary-plane compositor is available in DCK 1.0.1. The other templates
use the existing scrolling, timeline and composition APIs. The integer-color
pass is available in 1.0.2.

## Binary masks with a live palette

`composite.BitplanePalette` turns up to six binary planes into one indexed image.
The first mask contributes bit 0, the next bit 1, and so on. Palette entry 0
may be transparent, allowing a background, logo or scrolling layer to show
through. The mask images can come from sprites, glyph contours, vector polygons
or a retained animation bank. They must all have the configured full-stage size
and origin `(0, 0)`. Each plane defaults to alpha; `Channels` can select red,
green or blue instead for opaque monochrome artwork. Sampling uses the stored
premultiplied source components.

Run `go run ./examples/bitplanes` from the DCK source checkout to see four
independent moving mask banks over a dark base layer with a changing palette.

```go
colors := make([]color.NRGBA, 1<<4)
colors[0] = color.NRGBA{} // Let the layer below remain visible.
colors[1] = color.NRGBA{R: 255, G: 80, A: 255}
// Fill the remaining entries with the desired four-plane palette.

lookup, err := composite.NewBitplanePalette(composite.BitplanePaletteConfig{
    Width: 640, Height: 512, Planes: 4, Palette: colors,
})
if err != nil { return err }
defer lookup.Close()

// Each source is a caller-owned 640×512 image containing a binary alpha mask.
masks := []*ebiten.Image{body, shadow, textureA, textureB}
if err := lookup.Draw(screen, masks); err != nil { return err }
```

Call `SetPalette` when a cue changes colors; reuse the same masks and compositor.
One to four planes use one draw pass. Five or six use one retained, full-stage
packing surface and a second palette pass. No mask pixels are read back to the
CPU. Spaceballs' Ribbons, Trails and Noise show the three source layouts that
motivated this component. Its Noise, Angular, Sliced and Duet materials select
alpha for five retained pose masks and red for the opaque sixth material plane.
The scene migration matches all 309 before/after complete-frame captures over
5,106 rendered frames.

## Integer color passes on a scrolling layer, logo or complete scene

`composite.QuantizedColor` changes the colors of a borrowed image, independently
of its animation. Each RGB channel has its own grid: `{15,15,15}` models RGB12,
while `{31,63,31}` models RGB565. Passthrough preserves the original source;
the other operations can quantize, replace, scale or fade from a chosen target.
The numerator can follow a timeline, a text cue or a music-derived signal.

```go
colors, err := composite.NewQuantizedColor(composite.QuantizedColorConfig{
    Levels: [3]uint16{15, 15, 15},
})
if err != nil { return err }
defer colors.Close()

// The caller draws a live scrolling, logo or composite into its own surface.
state := composite.QuantizedColorState{
    Mode: composite.QuantizedScale, Numerator: 8, Denominator: 16,
}
if err := colors.Draw(screen, liveLayer, state); err != nil { return err }

// Start at white and approach the source color over thirty-two integer steps.
state = composite.QuantizedColorState{
    Mode: composite.QuantizedFromTarget, Target: [3]uint16{15, 15, 15},
    Numerator: 12, Denominator: 32,
}
```

Run `go run ./examples/palettefades` to compare four ordered image passes.
`-capture /path/to/output -frame 32` produces a deterministic PNG. An existing
effect can also provide its source through `composite.NewPass`; close the color
renderer separately from that pass. Each transformed draw takes one shader pass
and no CPU pixel readback. Transparent inputs are handled in straight RGB before
returning to premultiplied colors. `AlphaThreshold` is optional; Mental Hangover
sets it to `0.5` for its binary original artwork. Source timing and the divisors
16, 32, 64 and 128 remain production parameters.

## Projected or deformed bitmap cells

`scrolling.New` owns the text, font metrics, repetition, controls and geometric
cell modes. `font.NewImageCellBank` prepares a decoded CPU image through its
metrics; `font.NewCellBank` accepts another source encoding through a pixel
reader. Sampling happens once at initialization. Different fonts can have
independent dimensions, character sets, aliases and colors in the same painter.

```go
bank, err := font.NewImageCellBank(cpuAtlas, metrics, characters, 0)
if err != nil { return err }
cells := scrolling.CellPainterConfig{
    Fonts: map[string]*font.CellBank{"main": bank},
    Flat: scrolling.FlatCellConfig{Size: geometry.Vec2{X: .85, Y: .85}},
    Rows: func(_ string, row int, seconds float64) geometry.Vec3 {
        return geometry.Vec3{Y: 12 * math.Sin(seconds*3 + float64(row)*.2)}
    },
}
scroll, err := scrolling.New(scrolling.Config{
    Text: message, Fonts: map[string]scrolling.Face{"main": face},
    Font: "main", Speed: 510, X: 720, Repeat: true,
    Shape: "cells", Modes: map[string]scrolling.Mode{
        "cells": {Cells: &cells},
    },
})
```

Set `Shape: scrolling.CellCuboid` and supply `Cuboid.Size`, `Camera`, `Origin`
and optional eight vertex colors for voxel letters. `Rows` offsets can also
move depth; `Pose` can rotate, scale, hide or reposition each cell. Row offsets
are cached per font and time; `InvalidateRows` refreshes captured live parameters.
`CellPainterController("cells").SetOutlined(true)` selects wireframe material.
A configured `Wireframe` callback takes precedence over that controller value.

The default flat mode follows all four transformed corners, including tangent
rotation and mirroring. `Flat.Rectangles` preserves axis-aligned rectangle
sampling; `ScreenGap` adds constant pixel gaps. `CullBounds` controls cuboid
visibility tests, while the destination supplies actual raster clipping. Cuboid
cells crossing the near plane are discarded as a whole; keep the requested
row/pose depth beyond that plane. `Mode.Cells` uses regular text or supplied
glyphs rather than the separate projected/recycled transport backends.

OldSkool DirectX 8 Go uses the same two cell modes with its recovered font data
and source oscillator parameters. Its 750 sampled filled/wireframe complete
frames match the preceding renderer across two 9,001-frame traversals. The
native row paths stay authored data; face generation, culling, projection and
batch submission belong to DCK. `go run ./examples/cellscroll` demonstrates two
simultaneous lanes; hold Space for wireframe, or use `-capture /path/to/output
-frame 360` for one native PNG. `Scrolling.Close` closes its cell painters.

## Retained contours, silhouettes and trails

`composite.ContourBank` keeps a fixed number of pose slots with independent
layers. A write can clear its target or accumulate on the previous image; other
slots keep their data. `Image` returns a borrowed handle for a palette, raster,
reflection or another pass. A source choreography can choose the displayed
indices without rebuilding geometry resources or allocating new canvases.

```go
bank, err := composite.NewContourBank(composite.ContourBankConfig{
    Width: 352, Height: 290, Slots: 6, Layers: 2,
    FillRule: ebiten.FillRuleEvenOdd,
})
if err != nil { return err }
defer bank.Close()

err = bank.Paint(workingSlot, 0, true, func(batch *render.Batch) {
    for _, contour := range contours {
        batch.Fan(len(contour), func(index int) ebiten.Vertex {
            p := contour[index]
            return render.Vertex(p.X, p.Y, 0, 0, color.White)
        })
    }
})
// A second write with clear=false retains the earlier silhouette.
currentMask := bank.Image(workingSlot, 0)
```

`Fan` shares an even-odd batch across concavity and holes; omit a duplicated
closing vertex from its count. Source vertices can also supply colors and UVs
for a palette atlas or repeated material. Each vertex is mapped once, so the
mapper must be stable within that call. `StrokeContour` draws open/closed paths
with a configurable width and optional horizontal-edge exclusion.
`ParityContour` uses ray spans for independently deformed edges, including
endpoint-Y exchange. These methods work on any `render.Batch`, without a bank.

Spaceballs retains its own buffer-pointer/cue program and integer coordinate
maps while DCK owns mask storage and geometry submission. Mental Hangover keeps
its exact fixed-point projectors while sharing fan geometry and materials.
`go run ./examples/contourtrails` combines three delayed masks with a live
palette. `-capture /path/to/output -frame 200` saves a deterministic native PNG.

## Compose independent layers and timed changes

Place each effect on its own retained surface when it needs a mask or an image
pass. Update its clock once, then draw layers in visible order. A cue can change
a palette, scrolling mode, sprite bank or image phase without rebuilding the
other layers. `timeline.CueClock`, `CueRanges` and `StageSequence` can own these
changes; `kit.Layers` and `composite.SurfaceLayer` own the drawing order and
intermediate surfaces. DCK's `authoring.Project` records several common layer
types as JSON for the planned graphical editor.

The original production's operation counts, decoded tables and signal edges
are data. A shared component should own its resource lifetime, bounds,
transport and rendering path, while exposing those authored values as explicit
parameters. See the [effect map](EFFECT_MAP.md) for the new production-by-
production extraction boundary.
