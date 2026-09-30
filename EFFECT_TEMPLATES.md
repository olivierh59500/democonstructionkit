# Effect templates

These patterns turn source artwork and timing into reusable DCK components.
They are ordinary Go code: images, fonts and music remain caller supplied. The
binary-plane compositor is available in DCK 1.0.1. The other templates
use the existing scrolling, timeline and composition APIs. The integer-color
pass is available in 1.0.2.

Independent binary-material offsets are available in 1.0.5. They extend the
existing palette compositor rather than adding another rendering family.

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

### Independently moving materials inside a mask

`DrawOffsets` reads each binary plane at its own `{X, Y}` pixel offset. The same
image can provide several planes: an animated contour, scrolling text or logo
can combine two samples of one texture, with independent trajectories and a
changing palette. Positive X/Y reads farther right/down, making the visible
material move left/up. The samples remain nearest-neighbor; outside the source
they read zero rather than repeating.

```go
lookup, err := composite.NewBitplanePalette(composite.BitplanePaletteConfig{
    Width: width, Height: height, Planes: 3, Palette: colors,
    Channels: []composite.BitplaneChannel{
        composite.BitplaneAlpha, composite.BitplaneRed, composite.BitplaneRed,
    },
})
if err != nil { return err }
defer lookup.Close()

// All three planes have the configured dimensions; texture is borrowed twice.
planes := [3]*ebiten.Image{liveMask, texture, texture}
offsets := [3][2]float32{{0, 0}, {32, 18}, {-12, 40}}
if err := lookup.DrawOffsets(screen, planes[:], offsets[:]); err != nil { return err }
```

Update the offsets from a path, timeline or music signal before drawing. No new
image or pass is needed: up to four planes still take one pass, and five/six
take two. `Draw` always restores zero offsets, so it can alternate with
`DrawOffsets` on the same compositor. `SetPalette` remains independent of the
source images and trajectories. Transparent palette entries let the composition
sit above another scene. Close the compositor separately from its borrowed images.

Run `go run ./examples/offsetmaterials` for an animated contour with a hole and
two independently moving samples of one immutable material. `-capture
/path/to/output -frame 200` saves a native PNG. Spaceballs' Pattern, Wave and
Finale use this path with their original fetch offsets, fine-scroll delay,
RGB12 palette and display crop. All 690 sampled frames through all fourteen
effect units retain their preceding full-image hashes.

## Outline materials for the same moving scene

Meshes, warped images and sprite fields can switch rendering materials while
keeping the same animation and cached positions. This supports an entire
wireframe scene, a temporary contour cue or an outline above a filled object.
The outline path does not reset a transport or evaluate a second animation.

```go
mesh, err := effects.NewMesh(
    effects.Cube(60, geometry.Vec2{X: 1, Y: 1}, color.NRGBA{R: 255, A: 255}),
    nil, geometry.Camera{Center: geometry.Vec2{X: 320, Y: 240}, Focal: 400, Near: 1},
)
if err != nil { return err }
defer mesh.Close()
mesh.Transform = effects.Transform{Position: geometry.Vec3{Z: 200}, Scale: 1}
mesh.CullBackFaces = true
if err := mesh.SetOutline(effects.MeshOutlineConfig{
    Faces: effects.CubeFaces(), Width: 1.2,
    Color: color.NRGBA{R: 255, G: 180, B: 240, A: 255},
}); err != nil { return err }

if err := mesh.Update(frame); err != nil { return err }
mesh.Draw(screen)
mesh.DrawOutline(screen) // Both use the same transformed and deformed points.
```

`Faces` describes polygon boundaries, preserving their order. Omitting it uses
the mesh triangles; `CubeFaces` keeps cube quads free of triangulation diagonals.
Back-face culling uses the mesh's existing setting. Segments crossing the near
plane are clipped; wholly hidden edges do not connect to an artificial origin.
The referenced point set is compiled once and each point is projected once per
outline draw. The outline has its own color, width and blend. A solid mesh
reuses its white pixel; a textured mesh can borrow one through `White`, or owns
a fallback pixel when none is supplied. No full-stage working image is added.

```go
if err := warp.SetOutline(effects.WarpOutlineConfig{
    Width: 1.2, White: sharedWhite,
    Color: color.NRGBA{R: 230, G: 140, B: 250, A: 255},
}); err != nil { return err }
if err := warp.Update(frame); err != nil { return err }
warp.DrawOutline(screen)
```

Warp outlines use the configured grid, `Map`, depth order and current time.
They draw cell borders without drawing the source image. Filled `Tint` and
`Blend` remain independent from the outline material. Borrowed white images
must be one pixel with origin `(0,0)` and survive until the component closes.

```go
style := field.Style
style.Outline = &sprites.FieldOutline{Width: 1, Color: color.White}
field.DrawStyle(screen, style)
```

`FieldStyle.Outline` is an inset box skin for any `FieldRenderer`, including
projected particles. Image/frame dimensions, anchors, scale, rotation and
depth modulation still define the box. The image's pixels are not sampled;
the existing fallback white pixel draws its border. Width is in destination
units before rotation. Bounds at or below twice that width are hidden. Corners
are emitted without overlapping border strips, so translucent colors do not
double their opacity. This skin takes precedence over vector/image choices.
`HarmonicField.DrawStyle` draws cached samples with another style and leaves
the default `Style` intact.

Run `go run ./examples/outlinelayers` to compare one combined moving logo,
cube and sprite formation through filled and outlined materials side by side.
`-capture /path/to/output -frame 200` saves a native PNG. OldSkool uses these
three paths and retains all 750 complete filled/wireframe frame samples through
two 9,001-frame replays, without a new working image or animation clock.

## Harmonic bands and sprite formations

`composite.HarmonicBands` treats the two coordinates of a harmonic formation as
the heights of a strip's left and right edges. Each edge can combine its own
sine/cosine terms, rates, instance phases and envelopes. Width, thickness,
colors, pixel rounding and post-rounding bounds are independent parameters.

```go
bands, err := composite.NewHarmonicBands(composite.HarmonicBandsConfig{
    LeftX: 0, RightX: 640, Thickness: 8, Colors: stripColors,
    Motion: motion.HarmonicFormationConfig{
        Origin: motion.Point{X: 180, Y: 180},
        X: []motion.IndexedHarmonic{{Amplitude: 120, Rate: 1.1, IndexRate: .32}},
        Y: []motion.IndexedHarmonic{{Amplitude: 120, Rate: .9, IndexRate: .43}},
    },
})
if err != nil { return err }
defer bands.Close()

if err := bands.Update(frame); err != nil { return err }
bands.Draw(screen) // DrawOutline(screen, 1.2) uses the same cached poses.
```

`sprites.HarmonicField` uses the same formation for positions and owns its
bounded batch renderer. `Style.Image` can be a borrowed sprite; a nil image
uses the fallback white pixel. Count, origin, spacing, image metrics, anchor,
blend and appearance remain configurable.

```go
field, err := sprites.NewHarmonicField(sprites.HarmonicFieldConfig{
    Count: 80, PixelSnap: true,
    Motion: motion.HarmonicFormationConfig{
        Origin: motion.Point{X: 320, Y: 180},
        X: []motion.IndexedHarmonic{{Amplitude: 180, Rate: 1.2, IndexRate: .15}},
        Y: []motion.IndexedHarmonic{{Amplitude: 120, Rate: .9, IndexRate: .19}},
    },
    Style: sprites.FieldStyle{Image: ball, Appearance: sprites.FieldAppearance{
        Width: 24, Height: 24, AnchorX: .5, AnchorY: .5,
    }},
})
if err != nil { return err }
defer field.Close()

if err := field.Update(frame); err != nil { return err }
field.Draw(screen)
```

Both components accept explicit two-clock/envelope values through `Sample` for
music signals, controls or authored cues. `ClockScale` lets `Update` use another
time unit. `PhasePeriod` optionally applies signed phase remainder before each
wave; zero preserves the existing harmonic behavior. `Poses`/`Samples` expose
borrowed cached positions for another material, without evaluating the waves
again. Steady sampling allocates no memory and drawing does not advance time.
Harmonic bands support up to 4,096 strips; the sprite field supports up to
65,536 samples and flushes bounded geometry as needed. Close components before
releasing their borrowed images.

Run `go run ./examples/harmoniclayers`, or add `-outline`; **W** switches band
materials. `-capture /path/to/output -frame 200` saves a native PNG. OldSkool
uses sixteen strips and eighty sprites with its original 85-unit clock and
four oscillator values per formation. All original positions match through
9,001 ticks and all 750 complete-frame samples retain their previous hashes.

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
