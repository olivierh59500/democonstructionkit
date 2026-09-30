# Effect templates

These patterns turn source artwork and timing into reusable DCK components.
They are ordinary Go code: images, fonts and music remain caller supplied. The
examples use the source checkout's development API; DCK 1.0.0 remains the
published dependency for existing demos.

## Binary masks with a live palette

`composite.BitplanePalette` turns up to six alpha masks into one indexed image.
The first mask contributes bit 0, the next bit 1, and so on. Palette entry 0
may be transparent, allowing a background, logo or scrolling layer to show
through. The mask images can come from sprites, glyph contours, vector polygons
or a retained animation bank. They must all have the configured full-stage size
and origin `(0, 0)`.

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
motivated this component; they still use their pinned 1.0.0 renderers until
scene-level parity checks justify a migration.

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
