# Contour-font scrolling

Run from the module root:

```sh
go run ./examples/contourscroll
```

The example draws three text layers through `scrolling.New` and
`scrolling.Mode.Contours`. It generates its two polygon alphabets and sine lookup
table during setup, so it needs no image, font or audio files.

- The top lane repeats continuously with mixed square and beveled font banks.
  `scrolling.Chain` composes an ordinary sine baseline, glyph zoom and font tint.
- The lower-left lane maps font X to angle and font Y to radius through
  `motion.TablePolar`. Change its table, phase offsets, shift, center or the
  mapping callback to produce another ring or radial distortion.
- The lower-right lane projects the same alphabet through `motion.RationalGrid`.
  Its affine numerator and denominator coefficients control the perspective.
  A time-dependent offset then moves the projected text along a custom path.

`font.ContourBank` keeps artwork separate from the renderer. Each glyph supplies
its own polygons and pen advance; disconnected pieces and holes can remain in
separate contours. The example uses filled, non-overlapping cells. For fonts with
holes, set the contour painter's `FillRule` to `ebiten.FillRuleEvenOdd` and ensure
the complete parity run fits `BatchTriangles`.

The continuous lane supplies `scrolling.Glyph` values with independent font names,
scales and advances. The projected lanes use `scrolling.GlyphWindowConfig` for a
fixed visible set. A window callback reads the current character for each slot;
it does not advance a clock. Change characters during `Update`, or connect the
callback to an existing text sequencer, while keeping rendering in DCK.

With no custom point mapper, `ContourPainter` applies the ordinary glyph
`ebiten.GeoM`, including scroll transport and mode transforms. Its custom `Map`
returns final destination coordinates and may use or replace the supplied
`GeoM`. `TablePolar.Point` and `RationalGrid.Point` return `false` for invalid
integer arithmetic; forward that result to the renderer to omit the affected
contour. This example uses full integer precision without a machine-word wrap.

All lanes borrow one white texture. The scrollers own their renderer buffers;
closing them leaves the caller's texture and font banks intact. Font generation
and waveform compilation happen once, with no per-frame image readback.

Animation uses a fixed 50 Hz frame clock. To capture one deterministic native
frame for documentation:

```sh
go run ./examples/contourscroll -capture /tmp/dck-contourscroll -frame 250
```

`-frames` captures consecutive update ticks beginning at `-frame`. Its default
is one, and the maximum is 1,500. For a six-second, 50 fps video, capture 300
frames and encode their numbered PNG files with an existing FFmpeg installation:

```sh
go run ./examples/contourscroll -capture /tmp/dck-contourscroll-video -frame 200 -frames 300
ffmpeg -framerate 50 -start_number 200 -i /tmp/dck-contourscroll-video/%06d.png \
  -frames:v 300 -an -c:v libvpx-vp9 -crf 30 -b:v 0 -pix_fmt yuv420p \
  /tmp/dck-contourscroll.webm
```

Each capture reads the rendered frame only for PNG export. Interactive drawing
does not read back pixels; the video has no audio track.
