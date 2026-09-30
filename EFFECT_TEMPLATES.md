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

## Spatial palettes for shadows, raster rows and scenery

`composite.PaletteGrid` gives each cell two independently editable colors. A
borrowed control image selects a bank at a threshold or blends continuously
between the two. Cell sizes, offsets, a differently sized first span and an
optional index map determine the spatial palette. A second mask can replace
covered pixels with a body color. These masks can contain an animated logo,
scrolling glyphs, sprites or scenery.

```go
bodyColor := color.NRGBA{R: 8, G: 16, B: 28, A: 255}
grid, err := composite.NewPaletteGrid(composite.PaletteGridConfig{
    Width: 640, Height: 360, Columns: 16, Rows: 18,
    X: composite.PaletteGridAxis{CellSize: 40},
    Y: composite.PaletteGridAxis{CellSize: 20},
    ControlThreshold: .5, BodyColor: &bodyColor,
})
if err != nil { return err }
defer grid.Close()

// Both slices contain Columns*Rows colors, in row-major order.
if err := grid.SetColors(backgroundColors, shadowColors); err != nil { return err }
if err := grid.Draw(screen, liveMask, liveMask, composite.PaletteGridState{
    ControlOffset: [2]float32{-12, -8},
}); err != nil { return err }
```

The offset control sample produces the shadow while the unshifted body retains
the foreground shape. Positive X/Y reads farther right/down; outside the mask
is zero. `ControlChannel` and `BodyChannel` independently select alpha/R/G/B.
Colors may be translucent and are premultiplied before upload. `SetBodyColor`
changes an enabled body material without changing the mask or grid.

For color rows, use one column and one-pixel vertical cells:

```go
rows, err := composite.NewPaletteGrid(composite.PaletteGridConfig{
    Width: 640, Height: 360, Columns: 1, Rows: 360,
    ControlChannel: composite.BitplaneRed,
})
if err != nil { return err }
defer rows.Close()
if err := rows.SetColors(firstRowColors, secondRowColors); err != nil { return err }
if err := rows.Draw(screen, columnMaterial, nil, composite.PaletteGridState{}); err != nil { return err }
```

Threshold zero mixes the two colors with the stored control value, so a binary
column mask makes a checkerboard and a grayscale material makes smooth bands.
The source artwork remains static on the GPU while the small row bank animates.
The body image is optional when `BodyColor` is nil.

`FirstSpan` can reserve a taller initial cell; `Indices` can skip, repeat or
reorder palette rows/columns. Mapped span coordinates clamp at each end. Their
tables are copied at construction and limited to 256 entries per axis. Sources
must match the configured stage dimensions with origin `(0,0)`; they remain
caller-owned. Draw uses one triangle/shader pass and no pixel readback.
The only owned image is `2*Columns` by `Rows`, capped at 1,048,576 pixels.

Run `go run ./examples/palettecells` to combine a moving contour/body/shadow with
a live grid and to compare it with an independently moving continuous row
material. `-capture /path/to/output -frame 200` saves a native PNG. Spaceballs'
Blocks and Outline retain their native 24-pixel columns, 36-pixel first row,
16-pixel later rows and skipped palette row. Their palette image shrinks from
408,320 to 2,040 bytes. Mental Hangover retains its two colors per original
floor row. All 1,389 sampled complete production frames retain their prior hashes.

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

## Indexed artwork and packed-color palette transitions

`effects.IndexedImage` places an indexed image while DCK owns color conversion,
its palette clock and the retained output surface. The input stays borrowed.
Artwork uploads once; updates only change a small bank of palette uniforms.
A crop applies after conversion in coordinates starting at `(0,0)`, then normal
`ebiten.DrawImageOptions` scale, rotate, tint, blend or place the output.

```go
layer, err := effects.NewIndexedImage(effects.IndexedImageConfig{
    Image: encodedImage, Palette: []uint32{0x000, 0xf80, 0xfff},
    FPS: 50, Channel: composite.BitplaneRed, Scale: 15,
    Crop: image.Rect(8, 0, 88, 64),
})
if err != nil { return err }
defer layer.Close()
err = layer.Fade([]uint32{0x000, 0x000, 0x000},
    []uint32{0x000, 0xf80, 0xfff}, 100, 17)
if err != nil { return err }
err = layer.Update(kit.Frame{Time: seconds})
if err != nil { return err }
layer.Draw(screen)
```

This encoding stores index `i` as red `i*17`, with opaque input alpha. Select
alpha, green or blue for another image format. For byte-valued indices, set
`Scale:255`; offset and nearest rounding precede palette-bound clamping. Stored
RGB components are sampled as premultiplied values. `SourceAlpha:true` also
multiplies the palette result by the input alpha. Palette entry zero is an
ordinary color; choose a transparent entry with the lower-level
`composite.IndexedPalette` when transparency is part of the palette itself.

The zero packed format is RGB12. Construct RGB565 with
`palette.NewPackedRGB(palette.PackedRGBConfig{Bits:[3]uint8{5,6,5},
Shift:[3]uint8{11,5,0}})` and pass it as `Format`. Packed colors are opaque;
`IndexedPalette` accepts `color.NRGBA` banks when independently editable palette
alpha is required. Different layers can use different formats and image sizes.

`Fade` copies both endpoints. Before its start tick, the layer displays `from`;
the last of `ticks` steps reaches `to`. A one-tick transition retains `from`.
Seeking backward through `Frame.Time` reproduces the scheduled palette instead
of advancing hidden state. Negative time selects tick zero; absolute ticks and
fade arguments are bounded to signed 32-bit values on every platform.
`SetPalette` cancels the transition without changing placement or clock.
Packed interpolation adds a signed, truncated channel delta: a RGB12 fade from
15 to 0 halfway gives 8, rather than the 7 of a weighted unsigned sum. Ratios
outside 0..1 are rejected; intentional source-program extrapolation belongs in
an explicit authored controller.

Update each layer once, then draw it among scrolls, sprite fields, contours or
background layers in the desired order. To distort an indexed illustration,
render it through an existing `effects.Warp`/image-pass source and configure the
same reusable mapping used for a logo or scroll. Close the pass and layer
according to their ownership; closing `IndexedImage` leaves the borrowed input
alive. The high-level layer retains one source-sized surface and uses palette
conversion plus output placement, with no update/draw pixel readback.

`go run ./examples/indexedpages` combines two procedural palette layers with
RGB12/RGB565 fades and independent placement. `-capture /path/to/output -frame
140` saves a deterministic PNG. Spaceballs uses the same API for seven native
page kinds, retaining its original words, cue timing and high-resolution crop.

## Contour fonts, polar text and integer perspective

`font.ContourBank` caches vector glyphs independently of their renderer. Each
character supplies an advance and any number of polygon contours; separate
contours can describe holes or disconnected parts. `Closed:true` removes a
repeated closing point. Banks copy source data once and expose borrowed immutable
points during drawing. Empty glyphs retain their space advance.

Use the same `scrolling.New` constructor and normal text controls. A vector
`Face` pairs ordinary `font.Font` layout metrics with its contour bank, without
requiring a bitmap atlas. `Mode.Contours` collects those banks when its `Fonts`
are omitted. Different faces keep their own dimensions, advances, character
maps, bearings and scales.

```go
scroll, err := scrolling.New(scrolling.Config{
    Text: "VECTOR {font:large}FONTS {shape:wave}IN MOTION ",
    Controls: scrolltext.Braces, Font: "small", Shape: "flat",
    Fonts: map[string]scrolling.Face{
        "small": {Metrics: smallMetrics, Contours: smallContours},
        "large": {Metrics: largeMetrics, Contours: largeContours},
    },
    Speed: 120, X: 640, Y: 80, Repeat: true, Gap: 40,
    Modes: map[string]scrolling.Mode{
        "flat": {Contours: &scrolling.ContourPainterConfig{FillRule: ebiten.FillRuleEvenOdd}},
        "wave": {Contours: &scrolling.ContourPainterConfig{FillRule: ebiten.FillRuleEvenOdd},
            Map: scrolling.Sine(motion.Wave{
                Amplitude: 18, Spatial: .03, Speed: 2,
            }).Map},
    },
})
```

Use `FillRule:ebiten.FillRuleEvenOdd` for polygon holes. A complete glyph run
stays in one bounded batch, so overlapping glyphs retain parity rather than
blending separately. A rejected run is discarded atomically. `BatchTriangles`
defaults to 20,000 and can be smaller for mobile scenes; the maximum is 20,000.
The painter owns geometry storage and a white pixel when `Texture` is nil.
An explicit texture is borrowed; `UV` can sample a row palette or material
independently of the projected position. Colors also follow the glyph's normal
`ColorScale`, including control/mode changes.

`ContourPainterConfig.Map` receives the original font point, glyph sample and
ordinary glyph `GeoM`, then returns final destination coordinates. Nil uses
that `GeoM` directly. A custom map can apply it before or after another
projection, follow any curve, or return false to omit a contour whose points
cannot be projected. Invalid or float32-overflowing coordinates are omitted.
Compose the complete result with the existing image passes for lens, reflection,
warp or CRT effects; font rendering does not add an intermediate framebuffer.

`motion.TablePolar` replaces per-point sine evaluation with a copied signed
wave table. Configure axis phases, table width, arithmetic shift, center and
optional signed word wrapping. `motion.RationalGrid` supplies independent affine
X/Y numerators and a denominator, each with `{base,row,column}` coefficients.
Its signed division truncates toward zero; zero denominators and overflow return
false. These integer maps preserve classic tables and can be used for other
images or geometry through ordinary callbacks. They do not allocate per point.

For an existing authored slot transport, `GlyphWindow` supplies actual current
runes, font names or images through a callback. DCK owns the fixed layout bank;
drawing samples the window without advancing its clock. `scrolltext.ByteWindow`
can own byte positions, positive/negative crossing rules, visible slot count,
head commands, pauses and termination/repeat. Command lookahead skips opaque
payloads without executing handlers. Its raw byte cursor can later visit a
payload as a character; an initial command is not eagerly consumed. This is an
explicit native transport choice; ordinary text should use the regular control
parser. Repeating short messages retains bounded O(slot-count) work.

`go run ./examples/contourscroll` combines mixed vector fonts with sine/zoom,
a moving polar ring and a perspective text plane. All artwork is procedural and
prepared once. `-capture /path/to/output -frame 200 -frames 300` saves six seconds
of 50 Hz consecutive frames for a clip. Mental Hangover uses the same mode and
byte/projection components for its circular and perspective text, retaining 699
matching complete-frame samples and its original control/projection fixtures.

## Sparse point planes and depth queues

Sprite populations can choose individual skins through `FieldRenderer`, or
combine coincident pixels through `sprites.IndexedPointPlane`. The latter owns
its byte masks, touched-index list, independent membership bitset and draw batch.
Choose `PointPlaneOR` or `PointPlaneXOR`; palette entries control the resulting
mask colors. `DrawZero` explicitly selects whether touched cancellation pixels
use palette entry zero. This distinction matters when that entry is opaque.

```go
plane, err := sprites.NewIndexedPointPlane(sprites.IndexedPointPlaneConfig{
    Width: 320, Height: 200, Count: len(points),
    Collision: sprites.PointPlaneXOR, Palette: colors,
    Sample: func(index int) (sprites.PointPlaneSample, bool) {
        p, visible := projection.Project(points[index], offset)
        return sprites.PointPlaneSample{X: p.X, Y: p.Y, Mask: depthColor(p.Depth)}, visible
    },
})
```

`motion.WrappedPointProjection` supplies editable coordinate masks/biases,
reciprocal numerator and depth bias, shift, center and bounds. It preserves
signed word coordinates and long products; ordinary floating camera projections
can instead be supplied by the callback. `WordEulerVelocity` offers a copied
sine bank, angular period, quantization, output shifts and explicit normalized
or guarded lookup policies. The same projected samples can drive pixels or
sprites; scene cues and steering angles remain independent.

Call `plane.Sample` once after advancing the motion state, or use its `Update`
when the callback reads the current frame. `Draw` reuses prepared masks. Colors
can change through `SetPalette` without resampling positions. Only previously
touched pixels are cleared; no full-frame CPU clear, GPU upload or readback is
required. A white pixel is borrowed or initialized lazily on first draw.

`sprites.DepthQueue` owns an ordered point ring and cached visible samples.
Configure count through its copied point bank, leading depth, spacing, bounds
and crossing policies. `QueueUpperInclusiveFirst` includes equality on entry,
then uses strict correction, retaining native transitions at the upper edge.
Large jumps normalize with bounded arithmetic. An ordinary `geometry.Camera`
or custom `Project` chooses placement, scale, visibility and atlas frames.
`Style.Frames` supports differently sized frames; `DrawStyle` can reuse the same
sampled population with another material. Drawing never advances the queue.

## Peak signals and YM-driven visual effects

`modulation.PeakBank` accepts several variable charges, retains their maxima,
then applies configured decay. Duplicate columns combine before same-tick
release. This works for music, user input or arbitrary scene signals.
`sound.YMPeriodMeter` maps `Stream.YMRegisters()` snapshots through editable
volume curves, period ranges, column scales, rounding, mixer gating and envelope
rules. The adapter never decodes or advances music; playback remains DCK's job.

```go
registers, available := music.YMRegisters()
if available {
    if err := meter.Step(registers); err != nil { return err }
}
bars.Draw(screen) // Or bars.DrawOutline(screen).
```

`composite.GradientBars` owns its bounded batch and borrows a level callback.
Configure baseline, spacing, width, vertex colors, texture/UVs and outline style.
The same levels can drive sprite size, palette intensity or deformation instead
of bars. When a seek replaces the meter, use a closure that reads the current
meter rather than retaining a method bound to the earlier instance.

## Indexed artwork banks with retained slots

`effects.IndexedImageBank` separates borrowed index images from named palettes
and copied display slots. A slot selects image/palette indices, crop, transform,
tint, blend and visibility. `SetSlots` validates the complete window before
replacing it; an optional `Select` fills reusable slots once per `Update`.
`Draw` submits one direct palette lookup per visible slot, without an RGBA
conversion surface or a precolored copy for every palette.

```go
bank, err := effects.NewIndexedImageBank(effects.IndexedImageBankConfig{
    Images: indexedArt, Palettes: palettes, Slots: slots,
    Channel: composite.BitplaneRed, Scale: 15, MaxSlots: 8,
})
```

Palette entry counts remain equal across the bank. Crop coordinates start at the
source's local origin, including atlased images. Slot options keep source images
and uniforms reserved for the component. Palette updates change colors without
reuploading artwork; selectors and drawing do not own the caller's scene clock.
Spaceballs uses three slots over 21 images and two palettes for both short tile
passages. Its original cues and working/display groups select the slots.

`go run ./examples/sharedlayers` combines one indexed artwork bank, two projected
point planes and gradient bars driven by synthetic YM snapshots. It needs no
external image or music files. Its `-capture` and `-frames` options generate a
native PNG or consecutive frames for a clip.


## Cached row recipes with explicit precision

`motion.HarmonicRowProfile` composes ordered stages of the existing harmonic
formation engine. Each stage sums its own waves before adding them to a row;
a fixed `Index` shares one global sample across every row. Configure row count,
clock scales, origins, spacing and waves. `Apply` caches by time, so a logo
grid and its outline can share the same row motion without repeated sine calls.

`IndexedHarmonic.RoundProduct` rounds an amplitude product before accumulation.
`FusedIndexPhase` explicitly uses a fused index/clock phase. Both default to
false, preserving existing behavior; their native presets select the precision
needed by the source expression. They are independent of font size or artwork.
Several row stages preserve grouping between global waves, row spacing and
local waves without inventing another oscillator engine.

OldSkool uses these recipes for its logo, large flat letters and small cubelet
letters. All row/vertex poses match strictly across 9,001 ticks, while its 750
complete filled/wireframe samples remain identical.
