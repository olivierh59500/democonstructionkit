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

`scrolling.New` owns the text, font metrics, repetition and controls. A custom
`scrolling.Painter` can draw each lit font cell as a projected cubelet, a
wireframe edge, a colored quad or another material. The painter receives the
resolved glyph and its transform. A screen can change the painter with a
`scrolling.Mode` without duplicating the transport.

```go
cellPainter := func(dst *ebiten.Image, sample scrolling.Sample, op ebiten.DrawImageOptions) {
    // Read lit cells from the caller's font data, project them through a camera,
    // and submit visible triangles through one retained render.Batch.
}
scroll, err := scrolling.New(scrolling.Config{
    Text: message, Fonts: map[string]scrolling.Face{"main": face},
    Font: "main", Speed: 510, X: 720, Repeat: true,
    Shape: "cells", Modes: map[string]scrolling.Mode{
        "cells": {Paint: cellPainter},
    },
})
```

OldSkool DirectX 8 Go uses this boundary for both its large raster cells and
small projected 3D cells. A future shared cell renderer must preserve per-row
motion and whole-stage clipping before replacing those painters. Ordinary atlas
cropping does not preserve cells that travel outside the glyph rectangle.

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
