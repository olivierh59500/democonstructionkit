# Changelog

## 1.0.9 — Developer Edition — 2026-09-30

- Added `palette.PackedRGB`: configurable non-overlapping RGB word channels,
  integer expansion and signed channel-delta interpolation. The zero format is
  RGB12; RGB565 and RGB888 are configurable without a different renderer.
- Added `composite.IndexedPalette`: live color lookup from alpha/R/G/B encoded
  images, index scale/offset/clamping and optional source-alpha multiplication.
  The shader borrows artwork and retains reusable premultiplied color uniforms.
- Added `effects.IndexedImage`: an owned conversion surface, absolute-clock
  palette transitions, copied endpoints, output crop and image placement.
  Backward seeks reproduce the scheduled palette. Clock and fade bounds are
  checked before state changes and are portable across 32/64-bit targets.
- Spaceballs' title, credit, dragon and closing pages now supply artwork,
  palette words, authored timing and layout instead of a local shader/fade loop.
  All 1,218 sampled director frames match over 12,414 drawn updates; all 690
  effect-unit samples also match over 11,209 drawn updates.
- Added the asset-free `examples/indexedpages` composition and independent
  packed-color and GPU checks, including atlas coordinates and resource lifetime.

The migration retains the existing conversion surface and two-pass image path;
no artwork upload or pixel readback occurs during effect update/draw. Native Go
revision comparisons do not establish new Amiga or Android rasterization parity.
The source opening controller's intentional extrapolated fade remains authored
logic; the new bounded interpolation API does not replace that calculation.

## 1.0.8 — Developer Edition — 2026-09-30

- Added `composite.PaletteGrid`: two editable colors per spatial cell, binary
  or continuous control, independent mask channels/offsets and optional body
  replacement. Axes support size, offset, first span and copied cell-index maps.
- The component owns one small palette image and one shader. Updates upload
  only cell colors; drawing uses one pass with borrowed full-stage masks.
  Translucent banks retain premultiplied color and alpha.
- Spaceballs' Blocks/Outline and Mental Hangover's checkerboard finale use the
  common material, retaining source clocks, geometry, RGB12 words and layouts.
  All 690 Spaceballs and 699 Mental Hangover sampled complete frames match
  across 35,210 drawn updates. The independent floor raster oracle passes.
- Spaceballs' palette texture shrinks from 408,320 to 2,040 bytes. Its uploads
  remain 2,040 bytes; Mental Hangover retains its 2,176-byte row-bank image.
- Added `examples/palettecells`, native PNG capture and independent GPU checks
  for mapped spans, mask offsets/channels, thresholds, continuous blending,
  translucency and borrowed-image lifetime.

Complete DCK and both production test/build/vet checks pass. A muted Mental
Hangover traversal measured mean CPU draw submission of 31.66 to 32.22
microseconds; this single-run diagnostic excludes GPU completion/readback.
No stage-sized surface or additional material pass is inserted.

## 1.0.7 — Developer Edition — 2026-09-30

- Added `MeshEffect.SetOutline/DrawOutline`, with supplied polygon boundaries,
  width, color, blend, back-face culling and near-segment clipping. It reuses the
  transformed/deformed mesh points and projects each referenced point once.
  `CubeFaces` provides undivided quad boundaries without triangle diagonals.
- Added independently styled `Warp.SetOutline/DrawOutline` over its existing
  grid, map, depth order and frame time. Filled tint/blend remain independent.
- Added `FieldStyle.Outline` and `HarmonicField.DrawStyle` for batched sprite-box
  skins, retaining image metrics, anchors, rotation, mirrors and depth scaling.
  Translucent corner coverage avoids overlapping strips; aligned boxes skip
  rotation arithmetic. No new working image is added in OldSkool.
- Removed OldSkool's local cube projection, logo-grid and sprite-border loops.
  All 750 complete-frame samples match across 18,002 drawn updates in filled
  and wireframe modes. A longer filled replay retains 730 matching samples
  across 20,001 drawn updates per implementation.
- Added `examples/outlinelayers`, PNG capture, independent clipping/coverage
  checks, borrowed-resource lifetime checks and optional native CPU profiling.

Complete DCK and OldSkool tests, build and vet pass. Paired desktop diagnostics
show variation with display/system state rather than a stable performance
gain. A profiled extended replay measured mean update/draw submission of
31.45/419.15 to 25.98/366.44 microseconds; these timings exclude GPU completion
and readback. The initial slow run and controlled comparisons are retained in
the integration evidence rather than omitted from the measurements.

## 1.0.6 — Developer Edition — 2026-09-30

- Added `composite.HarmonicBands`: independently sampled strip edges, per-strip
  colors, thickness, rounding and clamps, filled/outlined materials and borrowed
  cached poses. It needs no intermediate image and can borrow a white pixel.
- Added `sprites.HarmonicField`: a batched sprite population with configurable
  indexed harmonics, absolute clocks, envelope sampling and optional rounding.
  Cached samples can feed another skin; images remain caller-owned.
- Added optional signed `IndexedHarmonic.PhasePeriod` and double-precision
  `render.Batch.StrokePath`, preserving native edge arithmetic until submission.
- OldSkool's sixteen bands and eighty sprites now use these components. All
  750 sampled complete filled/wireframe images match across 18,002 drawn frames;
  every applicable band/sprite pose matches the original oscillator routines
  through tick 9,000. No new working image is added.
- Added `examples/harmoniclayers`, PNG capture, independent geometry/coverage
  checks, seek/rounding tests and borrowed-resource lifetime checks.
- Spaceballs on DCK 1.0.5 and Mental Hangover on 1.0.4 completed their directors
  on Pixel 10a without an observed crash; five-second samples stayed near their
  50 Hz simulation target. Version-specific figures are retained separately.

Complete DCK and OldSkool tests, build and vet pass. One muted native traversal
per mode/implementation measured mean update/draw submission of 17.20/274.27 to
19.79/263.42 microseconds filled and 16.82/453.38 to 19.38/450.25 wireframe.
These diagnostics exclude GPU completion and readback; they are not device
performance claims. Pixel runtime evidence is separate from visual fidelity.

## 1.0.5 — Developer Edition — 2026-09-30

- Added `BitplanePalette.DrawOffsets` with independent finite pixel offsets for
  every binary plane. Several planes can borrow the same source image with
  separate offsets and channel selection; `Draw` resets to zero offsets.
- The existing one/four-plane and five/six-plane rendering paths retain their
  pass counts and working-image budgets. Out-of-bounds pixels sample zero;
  fractional offsets retain nearest-neighbor sampling.
- Spaceballs' Pattern, Wave and Finale now supply their original source offsets,
  palette and display crop instead of a local material shader. All 690 sampled
  full frames match through 11,209 drawn updates across fourteen effect units.
- Added `examples/offsetmaterials`, native PNG capture and independent CPU/GPU
  pixel checks for one through six planes, aliased atlased textures, integer and
  fractional offsets, edge clipping and alternating offset/ordinary draws.

The complete DCK test, build and vet checks pass, as do Spaceballs' source
controllers, rendering checks and full-unit replay. Source images stay borrowed;
the migration adds no working surface or palette pass. These desktop checks
do not constitute a new Pixel performance measurement.

## 1.0.4 — Developer Edition — 2026-09-30

- Added owned `composite.ContourBank` storage for fixed retained slots/layers,
  selective clearing, accumulating writes and borrowed image composition.
- Added bounded `render.Batch` fan, float32 stroke and parity-edge methods.
  Fans map each source vertex once and preserve source UV/color attributes;
  parity spans retain optional endpoint-Y exchange and a configurable ray edge.
- Spaceballs now uses the bank for all six-mask, paired-mask and pattern-mask
  families, plus the shared bitplane palette for opening and Duet's exit.
- Mental Hangover's author objects, BOBs, solids, patterned contours and both
  outline-font passages use the common fan submission.
- All 690 sampled complete Spaceballs unit frames and 699 Mental Hangover
  production frames match the preceding renderers across 35,210 drawn updates.
- Added `examples/contourtrails` and GPU checks for holes, modified-edge parity,
  selective outlines, XOR history and independent slot/layer lifetime.

Full DCK tests, build and vet pass, with Spaceballs' controller/planar oracle
and Mental Hangover's controller and GPU suites. Two muted native Mental
Hangover replays per implementation measured mean CPU draw submission of
90.66–91.36 microseconds before and 89.24–93.80 after. These timings exclude
GPU completion/readback; image counts and native pipeline pass counts stay
the same.

## 1.0.3 — Developer Edition — 2026-09-30

- Added immutable `font.CellBank` preparation from a CPU bitmap atlas or a
  caller-supplied pixel reader, with independent dimensions, Unicode character
  sets, alias sharing and per-pixel colors. No bitmap scanning occurs in Draw.
- Added owned `scrolling.Mode.Cells` to the existing `scrolling.New` facade.
  Flat quads/rectangles and cuboids support row paths, per-cell poses, camera
  projection, vertex colors, face culling/depth ordering and wireframe material.
  Scrolling owns their rendering resources; controllers expose outline toggles
  and row-cache invalidation.
- OldSkool DirectX 8 Go now supplies its font banks and authored row parameters
  instead of local cell drawing/projection loops. All 750 complete-frame samples
  match its previous filled/wireframe renderers over 18,002 drawn frames.
- Added `examples/cellscroll`, with two simultaneous cell scrollers and native
  PNG capture, plus configuration and GPU resource/mixed-font checks.

The complete DCK and OldSkool test, build and vet checks pass. Native CPU draw
submission in one muted 9,001-frame replay per mode changed from 737.34 to
715.49 microseconds filled and 1,118.58 to 1,029.59 wireframe. These measurements
exclude GPU completion and readback; near-plane-intersecting cuboids are
conservatively discarded by the generic painter.

## 1.0.2 — Developer Edition — 2026-09-30

- Added `composite.QuantizedColor`: independent RGB grids, integer scaling,
  replacement colors, fades from a target, optional alpha cutoff and byte-exact
  passthrough. It borrows any source image and needs one shader without an
  additional working surface. Translucent inputs retain premultiplied alpha.
- Mental Hangover now configures this component for its six copper operations.
  All 699 sampled complete frames match the previous renderer across 24,001
  rendered frames and 41 production units. Its independent oracle verifies
  every RGB12 color at the source transition levels.
- Added the `examples/palettefades` program, deterministic PNG capture and
  reusable color-pass templates. Native replay timing distinguishes CPU draw
  submission from GPU completion and pixel readback.

Complete DCK tests, build and vet pass. Mental Hangover's controller tests and
GPU suite pass against the new renderer. Two whole-production CPU submission
runs per implementation measured mean draws of 37.82–38.63 microseconds before
and 38.27–39.70 after; the pass needs no additional image surface. These are
desktop submission measurements, excluding GPU completion and pixel readback.

## 1.0.1 — Developer Edition — 2026-09-30

- Reviewed the Mental Hangover, Spaceballs and OldSkool DirectX 8 Go
  productions against the earlier twenty-repository effect map.
- Added `composite.BitplanePalette` for one to six binary alpha planes and a
  live, optionally translucent palette. One to four planes render in one pass;
  five or six use one bounded packing surface and a second pass. The caller
  owns masks and their timing. Each plane can read alpha, red, green or blue.
- Migrated Spaceballs' Trails, Noise, Angular, Sliced, Ribbons and Duet palette
  composition to DCK. All 309 sampled complete frames match the preceding
  renderer across 5,106 rendered frames, with unchanged GPU pass counts.
- Added configuration and opt-in GPU checks for bit order, palette updates and
  the one-pass/two-pass boundary. A source-checkout example is documented in
  [the effect templates](EFFECT_TEMPLATES.md).

The existing 1.0.0 API remains compatible. The complete DCK test, build and vet
checks pass, together with GPU palette/plane checks and Spaceballs' independent
planar pixel decoder. The next compatible library updates continue the 1.0.x
patch series; 2.0.0 remains the planned graphical editor.

## 1.0.0 — Developer Edition — 2026-09-28

The first versioned release of Demo Construction Kit is intended for Go
developers. It provides the reusable Go/Ebitengine effect library, executable
examples, command-line tools and a saved-project format. Creating a demo with
this edition requires programming knowledge.

- One scrolling constructor with configurable bitmap metrics, independent
  fonts, horizontal/vertical/path layouts, controls, zoom, perspective,
  recycled transports, cued DNA strips and feedback histories.
- Reusable sprite formations, projected stars/sprites/trails, animation banks,
  music-driven modulation and independently configurable 3D objects.
- Background repetition, rotozoom, image deformation, rasters, plasma,
  reflection, magnifier, CRT and ordered cumulative image passes.
- Owned update clocks, scene timelines, continuous transitions and saved JSON
  compositions rendered by the same desktop and Android hosts.
- One music facade for YM, tracker modules and recorded audio, using
  `ym-player v1.0.0` and `go-zikmu` internally.
- Native capture, comparison and Pixel presentation tools, with the catalog
  migrations and measurement limits documented in
  [the acceptance review](CATALOG_ACCEPTANCE.md) and [effect map](EFFECT_MAP.md).
- Immediate reporting of feedback source errors while retaining the last
  successful history and phase sample.

DCK and all twenty demo modules passed tests and vet against the release code.
The native font check covered 336 mixed-size font/mode combinations; the latest
DNA integrations have 42 exact complete-frame and 22 isolated comparisons.
Device evidence and known visual differences remain explicit in the effect map.

## Planned 2.0.0 — Visual Editor

Version 2.0.0 is planned as a graphical creation tool for everyone, including
users without programming knowledge. The entire creation workflow will use
the interface: assets, effects, parameters, paths, layers, music, timeline,
preview and export. Go code and manual JSON editing will not be required.
See [the roadmap](ROADMAP.md) for the intended workflow.
