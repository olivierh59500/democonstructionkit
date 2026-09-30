# Effect map for the native demo catalog

This map covers every DCK production under `demos/` except FR-010 and Second
Reality. The screen and workspace audits began on 2026-09-23; the wrap/train
inventory was checked across all 20 DCK repositories on 2026-09-25. The DCK
module-wide build, vet and test checks were refreshed on 2026-09-28. A complete
component owns its transport, animation state, geometry and rendering resources.
The production supplies assets, messages, presets, input and scene order. A
shared draw helper alone is not counted as a complete effect.

`Shared` names a complete component already used by that screen. The third
column distinguishes intentional production composition from reusable logic
still to extract. Authored artwork, input and layer order do not become DCK
effects simply to shorten a source file. The entries are acceptance work, not
claims of pixel-perfect or Pixel-device verification. The final
[catalog acceptance review](CATALOG_ACCEPTANCE.md) connects the requirements to
the tested revisions, source inventory and device evidence.

## New native productions reviewed on 2026-09-30

The original twenty-repository inventory predates Mental Hangover, Spaceballs
and OldSkool DirectX 8 Go. They initially used DCK 1.0.0; Spaceballs now uses
the binary-plane palette component added in 1.0.1, and Mental Hangover uses the
quantized-color component added in 1.0.2. OldSkool uses owned flat/cuboid
cell modes added in 1.0.3. Shared retained contours arrived in 1.0.4 and binary
material offsets in 1.0.5. Sampled harmonic bands/fields arrived in 1.0.6.
The following classification
distinguishes completed integration from remaining extraction candidates.

| Production | Reuse already in place | Further shared component or template | Source-specific boundary |
| --- | --- | --- | --- |
| Mental Hangover | `scrolling.New`, bitmap metrics, bounded `render.Batch.Fan` for all author/BOB/solid/patterned/outline-font contours, `RasterOverlay`, `CueClock`, cue ranges and `QuantizedColor`. | Polar transport/projection, row-palette floor sampling and original fixed-point geometry programs remain candidates for complete components. | Decoded 68000 tables, integer projection, level divisors, palette order and cue timing remain production data. Immediate contours use shared fan geometry without an additional working surface. |
| Spaceballs | `kit.Effect`, `timeline.Sequence`, `ContourBank` for six-mask/paired-mask/pattern-mask storage, shared fan/stroke/parity edges and `BitplanePalette` for retained planes, opening, Duet exit and offset materials. | Palette-grid and indexed-page components remain candidates; Pattern, Wave and Finale now share the DCK palette material with independent sampling offsets. | Decompressed point banks, source coordinate maps, 25 Hz pose interrupts, working/display indices, bitplane offsets and cue order remain production data. |
| OldSkool DirectX 8 Go | `font.CellBank`, owned `scrolling.Mode.Cells`, `effects.Warp`, `effects.Mesh`, `composite.HarmonicBands`, `sprites.HarmonicField` and DCK audio. | Cube wireframe and the YM-to-meter adapter remain extraction candidates; logo row oscillators can become a named preset. | The executable's font bit order, oscillator constants, cube face mapping, clipping margins and meter mapping remain authored parameters. Cell masks and harmonic poses are cached; shared components own their geometry and materials. |

### Screen-level extraction boundaries

| Spaceballs unit | Shared path now | Next reusable boundary |
| --- | --- | --- |
| Opening and the State/Of/The/Art, credits, dragon and closing pages | A common `PageEffect` renderer and `timeline.Sequence` already handle the authored pages and fades locally. | Palette-indexed page fades could move to DCK after their RGB12 integer endpoints, transparency and high-resolution crop are compared. |
| Pattern, Wave and Finale | DCK's three-plane palette material, with independent sampling offsets and two reads from one borrowed monochrome texture. | Source mask offsets, fine-scroll delay, bank selection, integer motion and display crop remain authored data; no local material shader remains. |
| Trails | DCK's six-slot `ContourBank` and five-plane `BitplanePalette`. | Working/display order, zoom coordinate maps and source cue order remain authored data. |
| Noise, Angular, Sliced and Duet | DCK retained bank, six-plane palette lookup and shared parity edges; Duet exit also uses the one-plane palette backend. | Source dither bit, endpoint exchange, coordinate maps and white-exit cue values remain authored configuration. |
| Ribbons | DCK's four-slot/two-layer `ContourBank` and four-plane `BitplanePalette`. | Paired draw/clear timing and mirror coordinates remain source parameters. |
| Blocks and Outline | Both use one local gradient-grid renderer with body/shadow masks. | A palette-grid material component with programmable cell size, shadow offset and color-bank uploads. |
| Tiles and Vote | Predecoded tile/palette images are selected by source clocks. | Existing image-bank and timed-layer components can express the display path; extraction should wait until tile selection, not just drawing, is configurable. |

| Mental Hangover unit | Shared path now | Next reusable boundary |
| --- | --- | --- |
| Eagle, title, cards and sign raster | DCK text pages, cue ranges, `RasterOverlay`, retained image layers and `QuantizedColor` for the six source copper modes. | Source artwork, operation order, thresholds and ratios remain authored parameters; no local color shader remains. |
| Author vectors, filled BOBs, filled solids and eight patterned objects | DCK's common fan submission, scrolling transport and original word-sized projections, preserving UVs, palette atlases and repeat materials. | Fixed-point projector and cue-program abstractions remain candidates; original point and cue tables stay authored data. |
| Star greeting pages | Cached DCK text pages and one batched point field. | A configurable page/star composition whose star mask and text palette each have their own cue level. |
| Circular text | The shared scrolling facade feeds a local polar outline painter. | A reusable polar glyph mapper/painter template with original control bytes as parameters. |
| Contact spheres and perspective text | Existing projected batches and scrolling samples, with local reciprocal/projective arithmetic. | A projected glyph/point template needs supplied integer projection and clipping rules before it can replace these scenes. |
| Checkerboard finale | Retained row-palette floor, projected balls and logo material. | The floor's palette-row shader can become a configurable image material; ball motion is already expressible through the DCK projected-sprite family. |

| OldSkool DirectX 8 Go phase | Shared path now | Next reusable boundary |
| --- | --- | --- |
| Cube, present from the start | `effects.Mesh` owns the camera mesh and filled faces. | Source face order and realtime interpolation remain parameters; a shared wireframe material could remove the local edge loop. |
| Colored bands, from 4 s | Complete `composite.HarmonicBands` with independent left/right waves, owned cached poses, filled/outlined materials and `render.Batch.StrokePath`. | Native four-oscillator values, 85-unit clock, thickness, palette, rounding and integer clamps remain parameters. |
| Ball field, from 12 s | Complete `sprites.HarmonicField`, with shared indexed waves, absolute clock, cached rounded samples and batched borrowed sprites. | Native constants, count, image, anchor and cue start remain authored configuration; the local pose loop has been removed. |
| Deformed logo, from 18 s | `effects.Warp` owns its source surface and ten-row mapping. | The map is already a parameter; expose a named multi-oscillator row preset rather than a new renderer. |
| Small 3D and large raster scrolls, from 26/36 s | One `scrolling.New` transport per font with owned `Mode.Cells`; cached native font masks, row paths, cuboid face generation, culling and filled/outline materials. | Source binary readers, row oscillator values, insertion edges and face colors remain authored configuration. The local cell geometry loops have been removed. |
| YM-driven frequency columns | DCK opens and replays the music; the production converts AY register periods to its authored columns. | A register-to-signal adapter could feed `modulation` without embedding this demo's 80-column strength and decay curve. |

The first new primitive is `composite.BitplanePalette`: 1–4 mask planes take one
GPU pass; 5–6 take a bounded RGBA packing pass and a palette pass. It accepts
caller-owned full-size planes with per-plane alpha/R/G/B selection and a replaceable palette with transparent
entries. The production still decides which retained mask frame appears in each
slot. Its unit and opt-in GPU checks cover plane order, palette replacement and
the source-count boundary, including opaque monochrome art in Ebitengine's atlas.
Spaceballs now uses it in six units. All 309 complete-frame fingerprints match
the previous renderer over 5,106 rendered frames; this includes the first twelve
and last three frames of each unit and samples every 23 ticks. The ordered
fingerprints and revisions are retained in
[the integration evidence](fidelity/spaceballs-bitplanes-20260930.json).

On 2026-09-30, `go test ./...` passed in all three new demo repositories on
their published DCK 1.0.0 pins. The DCK source checkout passed its complete
unit suite, build and vet, plus the new compositor's GPU pixel checks. These
checks, the independent six-plane pixel decoder and the complete-frame samples
establish renderer-migration parity for the six Spaceballs units. These captures
compare native Go revisions; they do not establish new Amiga hardware or Pixel
rasterization parity.

Mental Hangover's color extraction uses `composite.QuantizedColor` with an RGB12
grid and the source's alpha threshold of 0.5. Its operation table supplies
replacement gray, a fade from white and scaling divisors 16, 32, 64 and 128.
All 699 complete-frame fingerprints across 41 production units match the
preceding renderer over 24,001 drawn ticks, with no extra image pass or surface.
The independent source `PaletteWord` oracle checks all 4,096 colors at all 293
mode/level combinations. DCK also verifies RGB565 and eight-bit grids, mixed
targets, translucent colors, cropped sources and byte-exact passthrough.

Two timed desktop replays per implementation recorded mean CPU draw submissions
of 37.82–38.63 microseconds before and 38.27–39.70 after. The p95 ranges were
87.21–91.08 and 88.46–94.42 microseconds; peak Go heap ranges were 70.12–70.18
and 64.51–69.65 MiB. These timings exclude GPU completion and pixel readback;
Go heap excludes native and GPU resources. Raw timings, their limits and frame
fingerprints are retained in
[the color integration evidence](fidelity/mental-color-20260930.json).

OldSkool's flat/cuboid cell extraction preserves both original rendering modes.
Two 9,001-frame traversals at the authored 50 Hz simulation cadence yielded
375 matching complete-frame fingerprints each, covering the first twelve
frames, every 31 ticks, all five entrance cue windows and the final tick.
This samples 750 complete filled/wireframe frames; every intervening update
was drawn. Fonts retain their original binary readers and lookup order. The
generic cell bank caches each unique glyph's row-major lit pixels, while
`Mode.Cells` owns the painter and releases it through `Scrolling.Close`.

One muted native replay per mode/implementation measured mean CPU draw
submission of 737.34 to 715.49 microseconds filled, and 1,118.58 to 1,029.59
microseconds wireframe. These are desktop submission measurements, excluding
GPU completion and readback. Row paths are sampled once per font/time and no
bitmap pixels are scanned during Draw. Generic cuboids conservatively discard
cells intersecting the near plane; the native depth bank stays well beyond it.
The ordered fingerprints and timing samples are retained in
[the cell integration evidence](fidelity/oldskool-cells-20260930.json).

The contour integration covers all 14 Spaceballs units and the complete Mental
Hangover director. All 690 Spaceballs unit-frame fingerprints match through
11,209 rendered updates, including held opening planes, accumulating edges,
sliced endpoint exchange and delayed paired masks. All 699 Mental Hangover
fingerprints match through 24,001 drawn updates, including concave author
objects, textured stencil holes, projected BOBs and polar/perspective font UVs.
The source chooses update cadence and working/display indices; `ContourBank`
owns fixed image storage, selective clearing and geometry-batch lifetime.
`render.Batch` also exposes the same geometry for immediate destination drawing.

Fan mapping now computes each vertex once. Two controlled muted native replays
measured Mental Hangover draw means of 90.66–91.36 microseconds before and
89.24–93.80 after, with p95 219.42–223.17 and 224.13–226.83 microseconds.
Go heap peaks ranged 65.92–71.77 and 69.73–71.77 MiB. These are desktop CPU
submission/Go heap measurements, excluding GPU completion, readback and native
memory. No new surfaces or GPU passes are inserted into the native pipelines.
Source revision limits and fingerprints are in
[the contour integration evidence](fidelity/contours-20260930.json).

The binary-material extension shares independently offset planes through the
existing `BitplanePalette.DrawOffsets`. All 690 Spaceballs full-frame samples
match across 11,209 rendered updates, including 175 samples in Pattern, Wave
and Finale. Independent CPU/GPU checks cover one through six planes, integer
and fractional offsets, aliasing one atlased texture, zero-offset restoration
and transparent source boundaries. No image or rendering pass is added.
The cross-catalog shader search found no other equivalent binary offset
material. Spaceballs' palette grid/indexed pages and Mental Hangover's
row-palette floor are distinct extraction candidates. Revisions, ordered hashes
and the resource limits are in
[the offset-material evidence](fidelity/offset-materials-20260930.json).

The harmonic integration preserves OldSkool's 750 sampled complete filled and
outlined frames over 18,002 drawn updates. Its independent original oscillator
oracle also compares every strip and sprite pose at every applicable tick
through 9,000. Shared sampling allocates no memory; poses are reused by repeated
draws and the strips borrow the existing white pixel. Immutable strip colors
are boxed once at construction. No intermediate image is added.
One muted native traversal per implementation/mode measured update/draw means
of 17.20/274.27 to 19.79/263.42 microseconds filled and 16.82/453.38 to
19.38/450.25 wireframe. These are CPU submission diagnostics, excluding GPU
completion and readback, rather than a device performance claim.
Revisions and hashes are in
[the harmonic integration evidence](fidelity/oldskool-harmonics-20260930.json).

Spaceballs' DCK 1.0.5 and Mental Hangover's DCK 1.0.4 builds were installed on
the Pixel 10a on 2026-09-30 and completed their directors without an observed
crash. Five-second samples recorded 49.8–50.9/57.0–60.1 TPS/FPS for Spaceballs
and 49.2–51.0/58.1–60.1 for Mental Hangover. Peak Go heap was 67.0 and 53.3 MiB,
excluding native/GPU memory. This runtime traversal is separate from visual
reference fidelity; version-specific transitions and limits are retained in
[the Pixel runtime evidence](fidelity/pixel-latest-20260930.json).

OldSkool's published DCK 1.0.6 build was subsequently installed and exercised
on the same Pixel. Its 37 ten-second samples after the 36-second final entrance
recorded 49.2–50.9 TPS and 59.7–60.2 FPS without an observed crash. The filled
composition used normal audio and realtime cube interpolation. Native ELF
segments and APK placement passed 16 KiB alignment. This runtime check does
not establish new Android wireframe/visual-reference or memory evidence.
[Version-specific OldSkool device measurements](fidelity/pixel-oldskool-harmonics-20260930.json)

Spaceballs' palette grids and indexed-page fades remain candidates.
Mental Hangover's fixed-point projection/transport and floor
sampling remain separate extraction families. OldSkool's band, ball-field and
cube-outline path is the next rendering candidate.

## Cross-catalog check: wrapped motion and image trains

The DCK implementations in all 20 demo repositories, including every Cuddly
and Union screen, were searched for local threshold resets, per-image phase
loops and moving raster/background positions. This classification concerns the
`WrapBank`, `BounceBank` and `sprites.Train` families, not every effect in a
production. Preserved original implementations remain unchanged.

| Repository | Result for these effect families |
| --- | --- |
| `3d_doc` | Projected balls and checkerboard use their own shared components; no equivalent local wrap/train controller found. |
| `bilizir-demo` | Water reflection, warped logo, raster and cube already use their shared components; text transport is a different scrolling mode. |
| `dma-3d` | The masked three-speed starfield uses `motion.FrameField` and `sprites.BatchedSolidField`; its cyclic three-form mesh uses `geometry.MorphingMesh` and `effects.MorphingMesh`. |
| `dma-is-back` | Cube, intro feed, CRT, scanline scroll and the only intro/music handoff use dedicated shared components; no further local timed effect remains. |
| `go-cocoisthebest` | Sprite grid/translation already uses `sprites.Group`; scanline scroll and cube are separate families. |
| `go-cuddlymenu` | Colorshock's orbit/table clock, Ehhh raster train, LED backdrop/gradient wraps and harmonic letters, Big Sprite front/back `sprites.AxisFlip`, Big Sprite/Fullscreen/Digi Weave groups, and Mega Scroller's directional entrance bounce now use shared controllers. |
| `go-dom-intro` | Its sampled background uses `VerticalStripTrain`, its font-size program uses `scrolling.Config.SizeBank`, and animated stars use `sprites.AnimatedField`. |
| `go-fr010` | Software-rendered parts were searched; no direct GPU image-train or simple threshold-wrap equivalent was migrated. |
| `go-multiscreen` | Embedded Viva title rasters and TCB mountain strips use the same DCK presets as their standalone versions. The camera tour is a separate director effect. |
| `go-secondreality` | Indexed software effects were searched separately; their palette/VRAM clocks are not interchangeable with these Ebitengine image controllers. |
| `go-uniondemo` | Intro harmonic cell waves, Replicants raster trains and per-letter `motion.KeyframedFormation`, Beat Dis/Wow/TNT2/Level 16 wraps and Disk Copier's six-strip raster are shared. Beat Dis letters use the harmonic formation with a common wobble and individual phases. |
| `go-vectorballs` | Ball projection, morphing and reflection are different shared families; no direct image train/wrap candidate found. |
| `grodan-kvack-kvack-demo` | Four scroll lanes use `Config.Ribbon`, three `SurfaceLayer` compositions, `GatedBackgroundPair` owns the scenery, and the sprite chain uses `motion.HarmonicFormation`. |
| `megatwist` | Its multi-frequency, clamped sprite motion now uses the same formation family; `sprites.GlowPainter` owns the configurable halo passes. |
| `nonameno-demo` | Its text pages use `motion.GlyphPageCycle` and `sprites.GlyphPages`; the bottom ribbon uses the common scrolling transport and harmonic mode. |
| `phenomena-dna-scroll-intro` | Raster-bar thresholds trigger scene-state changes; treating them as a periodic wrap would change the sequence. |
| `tcb-multi-plane-3d-scroller` | All 32 mountain strips now use the relative `WrapBank` preset also used by Multiscreen. |
| `tcb-replicants-demo` | Layered stars, coupled logos, stepped splash and the foreground sprite pair use complete DCK components. |
| `teamg1-demo` | Circular sprite formation, intro handoff, timed CRT, plasma, cube and scroll use shared components. Remaining checks only compose their order and start music. |
| `viva_tcb` | Paired title rasters now use the shared `WrapBank` preset; harmonic logos remain a separate formation. |

## Individual demos and intro screens

| Production / screen | Shared today | Remaining responsibility or extraction |
| --- | --- | --- |
| 3D DOC | Atlas, `Config.RowBands` using `composite.RowWarp`, `PerspectiveCheckerboard`, `ProjectedBallTrain`, `timeline.IntroHandoff` | Whole-scene transform remains scene composition data. |
| Bilizir | Atlas, scrolling with a relative `motion.WrapBank` and `scrolling.CyclicWindow` over two bounded virtual text copies or an opt-in strict single-copy reset, two `StripWarp` instances on one `motion.WarpTableClock` with independent phase/gain variations, `motion.HarmonicTransform` logo paths on one `motion.WaveClock`, batched `SolidCubeTrain`, WaterReflection and `CopperBars` | Authored text and artwork stay production data; the default DCK loop bridges its former blank reset while the original source remains intact. |
| DMA 3D | Atlas, `Config.RowColumn`, `sprites.BatchedSolidField` with timed wrap, and complete `effects.MorphingMesh` with ordered per-face materials | Artwork, text, soundtrack and scene layer order stay production data. |
| DMA Is Back, intro | `Config.Feed`, configurable CRTOverlay, `timeline.IntroHandoff` | Message and scene materials stay production data. |
| DMA Is Back, main | `Config.Scanline`, ImageGrid, `motion.TrajectoryClock` with a `NestedOrbit` logo path, JellyCube, shared fade/music cue | Authored text, artwork, resource lifetime and layer order remain production composition. |
| Coco, intro | `Config.Feed`, configurable CRTOverlay, immediate `timeline.IntroHandoff` music cue | Scene materials stay production data. |
| Coco, main | `Config.Scanline`, `SolidCubeTrain`, sprites.Group grid/translation, shared font metrics, `RotozoomBackground` source quad and complete `CopperTitleBand` with a retained surface | Authored art and scene layer order. |
| DOM intro | Atlas, `scrolling.Config.SizeBank`, `motion.ScaledTextClock`, `VerticalStripTrain` background, `RasterOverlay` copies, `effects.Mask` and `sprites.AnimatedField` stars | Authored text and scene layer order. |
| Multiscreen, four embedded productions | `Config.Scanline`, recurrent `SolidCubeTrain`, `RotozoomBackground`, `sprites.Group` recurrent translation and direct `CopperTitleBand` in Coco, `RotozoomBackground` in Viva, Phenomena's main-only `scrolling.Config.CuedSlices`, `motion.RecurrentRowWave` with its owned scene clock, `composite.ScalarStagePainter` HSL materials and reusable gradient/mask assets, projected-plane engines, atlas recipes, Viva `RasterTitle`, `scrolling.Config.Pseudo3D` text banks and `sprites.RecurrentFormation` logos, complete TCB `effects.MultiPlaneScene` | Authored images, messages, panel sizes, placement and scene order remain in the host. No local trigonometric path, mesh submission or scrolling renderer remains in its DCK panels; the standalone Phenomena intro is preserved separately. |
| Multiscreen, camera tour | Complete `composite.SceneTour` over `motion.CameraTour`: ten held/eased poses, direct fixed views, continuously updated sources, masked retained transition canvases, one-pass shader and fallback | Authored screen sources, world placement and music stay production configuration. |
| Vectorballs | `sprites.PaletteAtlas` CPU recoloring and irregular-frame packing, `sprites.Projector.DrawIndexed` for direct projection of the authored point bank, `sprites.ProjectedObject` for independent cube, pyramid, plane, flag, sphere or custom clouds with owned rotation, deformation and projection; optional batched sprites, rear-hemisphere culling and perspective sizing; WaterReflection; and complete `geometry.PointSequence` for the authored point morphs, sine grid, rotors, Y orbit and bounce | Palette values, artwork, authored shape/action data and layer placement stay production parameters. |
| Grodan | Atlas, four `scrolling.Config.Ribbon` lanes, three `SurfaceLayer` raster compositions, `GatedBackgroundPair` and harmonic `sprites.Group` | Authored art and scene layer order. |
| MegaTwist, intro | `Config.Feed`, atlas, `timeline.HoldRamp` 90-tick splash, shared `CRTOverlay` with copy blending | Authored intro image and CRT parameters remain scene data. |
| MegaTwist, main | `Config.Scanline`, independent `ScanlineBackground` and DisplacementPrograms, harmonic `sprites.Group` and `sprites.GlowPainter` | Authored artwork and layer order remain scene composition. The former black transition overlay was visually inert after the splash reset and has been removed from the DCK version. |
| Nonameno, stars | `sprites.ProjectedField`, editable radial pattern and vector pixel/trail material | The star field is complete. |
| Nonameno, text pages and bottom scroll | Atlas, `motion.GlyphPageCycle`, three delay generators, `sprites.GlyphPages`, `scrolling.Config` with `HarmonicSineWith` and editable presets | Page text and bottom message remain production data. |
| Phenomena DNA intro | Atlas, `scrolling.BitmapPage` retained intro stills, DNAFrames, `scrolling.Config.CuedSlices` with two-pixel insertion, loop-start and configurable control cues, `motion.RecurrentRowWave` baseline and owned scene clock, `motion.GravityBounce` photon path, `motion.WrapBank` hue clock, `timeline.ScalarStages` intro/outro thresholds, `composite.ScalarStagePainter` ordered reveal/fade and HSL main materials, and reusable gradient/silhouette/inverted-font assets | Authored pre-roll, input/music and final black mask remain production composition. |
| TCB multiplane | Complete `effects.MultiPlaneScene`: projected scrolling, font-independent forms, snapped `composite.Bands` mountains, per-row `composite.ProfileImage` logo, `sprites.AxisFlip` emblem | Text, artwork and music remain production parameters. |
| Replicants | Atlas, `Config.RowColumn` with variable-speed controls, `sprites.AnimatedField` stars, `CoupledLogoPair`, `BlockReveal` and `sprites.Train` foreground pair with owned speed-scaled phase | Authored text, input and scene order. |
| TeamG1, intro | `Config.Feed` with the authored 640-pixel insertion edge inside the 768-pixel output, atlas, flat TimedCRTOverlay, `timeline.IntroHandoff` | Message and materials remain production data; the flat CRT intentionally avoids clipping the original curved pass's outer glyph rows. |
| TeamG1, main | TexturedCube, HarmonicImage, ProfileImage, `Config.Profiled`, sprites.Group circular formation | Timed scene/audio cues and whole-scene presentation remain composition data. |
| Viva TCB | Atlas, four editable `scrolling.Config.Pseudo3D` glyph banks, ten-logo `sprites.RecurrentFormation`, staged `RotozoomBackground`, two-mode `composite.RasterTitle` and held `motion.WaveClock` title path | Whole-scene layer schedule and host music cue. |

Vectorballs' shared sequence distinguishes absent fields from explicit zero
values, and an omitted animation list from an explicit empty list. Rotation and
translation advance before ordered point animations; bounce replaces Y while
Y rotation replaces the full position. Stage-boundary ticks select the next
action without also advancing motion. Shape replacement remains an explicit
reset, while `PointMorph` starts from the current coordinates and keeps each
ball's image index through a handoff.

Second Reality remains outside this full-screen inventory, but its Rotozoomer
now uses the `indexed.Rotozoom256` backend. The live RGBA tile used by Viva and
the palette-indexed fixed-point sampler in Second Reality expose different
renderers so both retain their source pixels. Its 2,580-frame isolated capture
matches the previous implementation exactly.

## Cuddly presentation units

| Screen | Shared today | Remaining responsibility or extraction |
| --- | --- | --- |
| Menu | TileAlphabet, `sprites.Atlas` image banks, `sprites.FrameSequence` character animation, `sprites.FormationCarousel` with seven compiled formula modes, `composite.CachedTileParallax`, `motion.CameraFollow`, scrolling, shared `CRTOverlay` with copy blending | Door/input semantics and map content stay local. |
| Loader | Bitmap font recipes, scrolling, `timeline.Countdown`, `timeline.CueClock` overlapping fade/hold windows and `timeline.CueRamp` gain applied by the playback host | Initial pre-render, text and layer placement remain authored. |
| Introduction | `composite.WaveChain` with an editable row-then-column logo preset, `ProfileStrips` with compiled `motion.WaveWrite` table, configurable sparkles and `timeline.StillThenMain` for the still/blank/music handoff | Authored artwork, star positions, music filename and layer order remain scene data. |
| Big Sprite | `sprites.ProjectedField` with strict-wrap vector lines, `scrolling.Config.RingLanes` dual-font transport, `sprites.Group` Weave formation, `RasterOverlay` fill, `sprites.AxisFlip` front/back material on a `motion.TrajectoryClock` orbit | Authored layer placement and image order stay scene data. |
| Colorshock II | Background sampler, scrolling, editable `FormulaFormation` two-frequency orbit on `motion.TrajectoryClock` and `WrapBank` position-table clock | Source position samples and artwork remain screen data. |
| Ehhh | Scrolling, `composite.TableWarpLogo` over the compiled row profile shared with Digi, `sprites.Train` raster bars, `motion.TrajectoryClock` backdrop orbit, `motion.CuedWaveClock` lookahead/landing roller, `motion.WrapBank` inner strip and `RasterOverlay` middle-text fill | Authored art, text and layer order remain screen composition. |
| Mega Scroller | `composite.TiledWaveBackdrop` for the cached two-wave tile field, finite overlapping `composite.Background` mask copies, scrolling, directional `motion.BounceBank` text-surface transport and reusable `RasterOverlay` source-atop fill | Authored text, bar artwork and scene layer order remain production data. |
| Spreadpoint | `Config.Bands`, `Config.Output.Feedback` DNA over a recycled text lane, atlas, a 20-ball `sprites.Group` with pixel-snapped formula, a dynamic `SurfaceLayer` for the raster-filled logo, `motion.PhaseSequence` with `motion.HarmonicTransform` for its orbit/zoom and `timeline.TintedCards` for the four stills and handoff cues | Authored card images, music filenames and layer order remain scene data. |
| Digi | Scrolling, complete `composite.TableWarpLogo` over compiled `motion.WaveWrite` rows and one rectified bounce shared with the Union logo, `sprites.Group` Weave formation, and independent `motion.WaveClock` scroll bounce | Authored artwork, text and layer order stay screen data. |
| LED Scroller | `composite.TiledWaveBackdrop` with retained scrolling source and explicit preload, cached bubble matrix, scrolling, `motion.WrapBank` gradient offset, harmonic `sprites.Group`, rectified `motion.WaveClock` and `palette.UniformGradient` 11-color raster material | Authored artwork and layer order remain scene composition. |
| 3D DOC | Scrolling, paired `composite.RowWarp` source-sampling programs, `PerspectiveCheckerboard`, `ProjectedBallTrain`, first-main-tick `timeline.IntroHandoff` cue and two ordered `RasterOverlay` source-atop passes for the inner text | Authored artwork, text and scene placement remain production data. |
| Fullscreen | `BackgroundLayer` velocity, `sprites.Group` Weave formation, `scrolling.Config.RingLanes` with paced vertical wrap and a bounded `SurfaceLayer` logo bar | Authored layer order and screen text. |
| Starwars | `Config.Crawl`, `scrolling.Config.DualProfiled` with two independent fonts, segmented wave profile, source-in raster and exact first-frame filtering, `sprites.ProjectedField` with camera velocity/angle and depth shade, `sprites.SampledSpriteTrain` over one authored XY path, and `modulation.PeriodicDecay` backdrop flash | Authored text, image assets and layer placement remain scene data. |
| Knucklebuster | Scrolling and complete `sprites.LatchedOverlay` over `motion.LatchedTriggers`: seeded five-tick hit sampling, three held channels and four ordered sprite images | Authored artwork and message remain screen data; a separate signal preset supports music-driven variations. |
| DNA | Two `Config.Output.Feedback` faces with independent fonts and gradients, wave strips, seeded `geometry.SphereCloud` distribution and complete `sprites.RotatingDiscCloud`, `composite.TwistingRibbon` with exact occlusion and two `composite.SampledRows` logo programs | Authored intro/music cue and layer order. |
| Megaball | Scrolling, rectified `motion.WaveClock` text bounce and `sprites.Group` with an ordered `motion.CoupledOrbitFormation` for two interleaved ball trains | The nine-value input mapping and authored labels remain screen data. |
| Reset | `Config.Slots`, scrolling, atlas, `timeline.StageSequence` with immediate events and cue windows, `composite.PairedRasterOrbit` for ordered pairs, `motion.WrapBank` for vertical raster fill and `motion.FormulaTrajectory` for the two-clock backdrop orbit | Source-atop compositing and layer order remain screen composition. |

## Union presentation units

| Screen | Shared today | Remaining responsibility or extraction |
| --- | --- | --- |
| Introduction | Image repetition on `motion.WrapBank`, `HarmonicCellWarp` with two active editable row/column wave banks, `motion.HarmonicTransform` logo path and bitmap recipes | Authored artwork, text placement and layer order remain screen composition. |
| Menu | Scrolling, background sampler, `sprites.Atlas` character frames, `motion.WrapBank` panorama, `motion.LinearTick` uncover wipe, `timeline.PacedIndex` palette/walk cycles, `motion.HoldBounce` logo and `motion.WalkParallax` hall/banner | Door navigation and authored layer order stay local. |
| Loader | `Config.Reveal`, bitmap recipes and `timeline.CueClock` exact-duration transition | Playback cue and layer placement remain host composition data; retain column-major reverse-row glyph order. |
| Beat Dis | Background sampler with independent single-copy entrance, scrolling, `motion.WrapBank` wallpaper and pattern offsets, harmonic `sprites.Group` letters with global wobble | Layer order stays authored scene data. |
| Delta Force | YM register snapshots and complete `sprites.SignalFrameBank` for the three voice-triggered ball frames, scrolling, WaveStrips, `motion.BounceToggle` logo, `motion.WrapBank` gold fill, `motion.EnterHoldExit` handoff and one reusable `RasterOverlay` source-atop material for both texts | Authored text, images and layer order remain screen composition. |
| TNT Crew 3 | Complete `effects.SolidMeshCarousel` over `motion.ModelCarousel` and `geometry.OrderedEuler`, reusable `SolidSphereModel` with editable checker and pole materials, plus `scrolling.Config.Caption` for ordered slide/hold/exit text | Other authored vertices, face groups, messages and input mapping stay production data. |
| Wow Scroller | Scrolling, `RasterOverlay` source-atop fill and `motion.WrapBank` paired panel offsets | The two 640×1235 images are single oversized draws naturally clipped by the 640×400 stage; no additional repeated-image transport is present. |
| Hidden | `sprites.DelayedTrail` over `PointHistory` and `timeline.PacedIndex` for the one-tick palette offset | Authored palette colors, crosshair and border clipping remain scene composition. |
| Starballs | Camera, scrolling, `sprites.MaskedProjectedField` with one projected population, two borrowed materials, editable mask paint, live count and depth opacity | Authored logo, scroll message and input mapping remain screen data. |
| Replicants | Atlas, two synchronized bitmap text windows on a strict `motion.WrapBank` clock, `RasterOverlay` fill, `motion.KeyframedFormation` with independent per-letter paths and a seamless 25-second loop, `sprites.Train` rasters driven by `motion.BounceBank` | The authored position table, text, key mapping and layer order remain scene data. |
| TNT Crew 2 | Background sampler, BitmapText.DrawWindow on a strict `motion.WrapBank` text clock, `motion.WrapBank` three-layer parallax and mutable speeds | Per-key control mapping remains scene data. |
| Level 16 | Vertical scrolling, background sampler, `motion.TrajectoryClock` with a `NestedOrbit` ball path and two independent `composite.RasterOverlay` materials with exact water/raster wraps | Authored artwork and layer occlusion remain scene data. |
| Multi-Plane | Complete `effects.MultiPlaneScene` in native-stage mode, with Union's source-index phases, `composite.ProfileImage` row renderer and strict `sprites.AxisFlip` cycle | Artwork, text, soundtrack and door routing remain production data. |
| Disk Copier | Bitmap recipes, `sprites.Atlas` LCD regions, `motion.GatedWrapBank` three LCD clocks, `composite.WindowedImageBank` six-strip raster with its shared phase, `timeline.CueRanges` stages and `timeline.SteppedEnvelope` palette | Input/state program and LED layer placement remain scene composition data. |

Union's recording tour additionally uses `motion.RampedWavePath` for its
hidden-screen pointer route. Its independent X/Y frequencies and two-second
entrance ramp are configuration; door order and screen durations remain tour
data.

## Verification ledger (2026-09-27)

- All 20 demo modules pass `go build ./...`, `go vet ./...` and `go test ./...`
  against DCK `e2819d1` through temporary Go workspaces; their committed
  dependency pins were unchanged. The earlier build/vet pass with published
  pins also passed. FR-010 and Second Reality are included in compatibility
  checks, not in the full-screen effect audit above. Nonameno's wave test had
  a pre-existing 1.26e-12 tick-to-second rounding failure even on its pinned
  DCK version; its corrected tolerance now passes with both dependencies.
- `cmd/checkeffects` rendered 336 combinations of catalog bitmap fonts and
  shared scroll modes with mixed glyph sizes. Every cell had visible glyph
  pixels and matched a second Draw without stepping. Union's ramped tour path
  also matched the former hidden-pointer position within 1e-12 and the same
  raster pixel at every tick from 0 through 3,600 at both 50 and 60 Hz.
- DCK's complete `GOWORK=off go test ./...`, `go build ./...` and `go vet ./...`
  pass. The explicit `dck_gpu_rendercheck` suite also passes for `composite`,
  `scrolling` and `sprites`; these assertions read GPU pixels only after an
  Ebitengine game loop starts. The separate `dck_composition_rendercheck` suite
  preserves paired-background draw order and masked output pixels.
- The saved `authoring` formation compiles through the normal sprite group;
  tests compare its poses at four times, reject invalid cues and loops, and
  cover independent property clocks. The standalone JSON example rendered at
  frame 120 with a moving backdrop, text layers and a bent sprite phrase. A
  second saved project rendered four independent keyframed sprites with the
  same layered host at frame 90, without scene-specific Go callbacks.
- Union Replicants now uses 126 ordered, per-letter position banks over a
  25-second loop. Sixteen aligned Atari-video/DCK capture pairs cover the
  opening stacks, wave section, fast two-way arcs and loop join. The point
  data is authored from five-frame-per-second observations, so overlapped
  letters remain approximate rather than claimed pixel-identical. Its updated
  Pixel 10a APK yielded 1,986 distinct presentation intervals across 33.15
  seconds of frame history: p95 16.734 ms, maximum 17.091 ms, none above
  20 ms and thermal status 0.
- DCK's pure `motion`, `geometry`, `timeline`, `timeline/recipes`, `palette` and
  `modulation` test suites pass. Eleven opt-in GPU comparators ran successfully
  against DCK code `f9ebe64`: water reflection, staged materials, copper titles,
  tiled waves, windowed rasters, repeated backgrounds, bitmap pages, the jelly
  cube, Bilizir's batched cube train, bitmap scroll bands and projected logo
  rows. The jelly check covered 42 complete cube captures, the scroll bands
  included 120,000-character runs, and projected rows had 50 exact comparisons.
  The ordinary sandbox exposes no monitor, so these checks need native access.
- Bilizir's logo/scroll GPU comparator passed at eight checkpoints through
  frame 4,800, with zero scroll-reference pixel differences. Additional demo
  comparators passed for Coco's cube train (10 captures), Multiscreen's cube
  train (10) and DNA scroller (11), Phenomena's DNA scroller (11), and the
  Vectorballs shape presets. These sample specific effects, not whole demos.
- Bilizir now has a direct full-frame comparison with the preserved Go original.
  The original-logo and strict-scroll-reset options reproduce all eight
  complete PNG captures from frames 0 through 4,800 byte for byte. With the
  default seamless loop, the scrolling region still matches through frame
  2,400; after the first wrap it intentionally differs from the old blank reset.
  The optional warped logo remains an independent DCK variation. The published
  revision `94852f4` also passed `cmd/fidelity` against pinned original
  `4720c97`: zero different pixels out of 480,000 at every sampled frame.
- Before/after full-frame GPU captures of the trajectory-clock migrations have
  zero differing RGBA pixels: Cuddly Big Sprite and Ehhh at frames 0, 600,
  1,200, 2,400 and 4,800 each; Union Level 16 at 0, 600, 1,200 and 2,400;
  DMA Is Back at 0, 1, 60, 240, 600, 1,200 and 2,400. Earlier Cuddly captures
  also matched at 60 and 240. The compared revisions were `640b95a`/`a1823ca`
  for Cuddly, `e608b27`/`678afb7` for Union and `aaa9946`/`025a53f` for DMA.
  These checks cover the four migrated scenes, not the full catalog or every
  transition.
- Cuddly Colorshock II also moved its compiled formula orbit to the same
  trajectory clock. Full-frame captures before and after this migration have
  zero differing RGBA pixels at 0, 60, 240, 600, 1,200, 2,400 and 4,800,
  covering its indexed scroll-table wrap (`a1823ca`/`525b7fa`).
- The standalone Phenomena DNA scroller and its Multiscreen panel now use the
  row wave's owned 0.30-step frame clock. Each native comparator matched its
  independent previous-formula timer and strip pixels at eleven checkpoints
  through frame 48,000, including pauses and late loop playback.
- The Replicants foreground `sprites.Train` now owns its speed-scaled wave
  phase. Nine seeded full-frame captures before and after this extraction had
  zero differing RGBA pixels at frames 0, 1, 99, 100, 101, 240, 600, 1,200
  and 2,400. A separate 5,000-tick controller test includes live speed changes
  and allocation-free sampling.
- Standalone TCB Multi-Plane now has a direct complete-frame comparison with
  its preserved Go original, not only the preceding DCK renderer. Twenty-one
  PNG captures from frames 0 through 9,600 are byte-identical, including the
  two logo wave-section joins, its strict phase reset, mountain wrap and later
  scrolling forms. This does not establish parity with Atari ST hardware.
- TeamG1's corrected `Config.Feed` now inserts letters at the authored
  640-pixel stage edge and uses the source's glyph-change rule. Its pre-CRT
  intro surface is byte-identical to the preserved Go original at frames 0,
  1, 60, 240 and 600; complete main-scene frames at 1,200, 2,400 and 4,800
  also match. The flat CRT intentionally changes the final intro pixels to
  prevent the original curved pass from clipping the outer font rows.
- Vectorballs' four alternate object families now use one complete
  `sprites.ProjectedObject` instead of a screen-local point/flag/rotation loop.
  Thirty-six complete before/after GPU captures, including shape-switch ticks
  and the water reflection, had zero differing RGBA pixels. Pure tests compare
  5,000 updates for every shape, independent instances and no steady-update
  allocations. The original authored `PointSequence` remains unchanged.
- Vectorballs' fifteen color variants now use `sprites.BuildPaletteAtlas`
  instead of a local per-pixel recoloring and irregular-sheet assembly loop.
  Fifty-five complete before/after GPU captures through frame 659, including
  every optional-object switch and the water reflection, are byte-identical.
  Pure DCK tests cover irregular crops, copied source colors, transparent
  pixels, palette changes and invalid geometry.
- Vectorballs' authored point sequence now implements
  `sprites.IndexedPointReader` and is projected directly without copying the
  point bank on every frame. Eleven complete before/after GPU captures through
  frame 4,800, including the reflection, are byte-identical. Pure projector
  tests match both input paths with and without culling and verify zero
  steady-draw allocations for the mutable reader.
- Cuddly's menu and MegaTwist's intro now use the same parameterized CRT pass
  already used by DMA and Coco. A configurable copy blend reproduces their
  transparent-edge handling. Three complete before/after GPU captures per
  production, at frames 0, 60 and 240, differ by one Cuddly pixel per capture
  and by 0, 2 and 2 MegaTwist pixels. The remaining differences are chromatic
  sampling at boundaries where the former literal and new uniform shader
  constants round differently. A dedicated GPU check verifies default
  source-over and explicit copy blending on transparent and opaque source areas.
  Cuddly's Android menu now exposes this optional pass through a CRT OFF / CRT
  ON touch button beside its other controls. Both Cuddly Android apps use the
  same control, while desktop retains the C key.
- Union TNT Crew 3's faceted ball now comes from `effects.SolidSphereModel`.
  A pure comparison matched all 512 coordinates and 112 ordered colored faces
  exactly against the previous builder. Complete GPU captures at frames 1, 60,
  240 and 600 are byte-identical; the remaining four object models retain their
  authored point and face data.
- Union Delta Force now passes its three live YM levels to one configurable
  `sprites.SignalFrameBank` instead of keeping a screen-local change/decay/frame
  loop. Eight complete music-driven GPU captures at frames 0, 1, 2, 3, 30, 60,
  240 and 600 are byte-identical to the previous renderer. Pure tests also
  cover independent channels, rising-threshold module/PCM inputs, palette offsets,
  invalid frames and allocation-free YM updates.
- Cuddly DNA's 125-disc sphere now uses a configurable `geometry.SphereCloud`
  distribution with its seeded random source. Six complete before/after GPU
  captures, including main-scene ticks 1,200 and 2,400, are pixel-identical.
  Vectorballs additionally gains an evenly spaced sphere through the same
  point-bank family; its optional-object GPU check includes the new mode and
  its reflection. The recommended 144-ball mode and the dense 4,096-point mode
  have now both been selected and sampled on Pixel.
- Pixel 10a (Android 17/API 37) was reconnected and sampled at 60 Hz with
  current DCK APKs. Cuddly's Big Sprite, DNA main stage, Mega Scroller, Reset,
  Starwars, Fullscreen and introduction reported 59.81–60.25 observed FPS in
  repeated 250-update windows. The largest measured CPU Update time was about
  5.43 ms (DNA main); Draw submission was at most about 0.34 ms. These host
  timings exclude asynchronous GPU work.
- The Go `cmd/pixelprobe` samples SurfaceFlinger's actual-present timestamps.
  Twelve history windows at two-second spacing yielded 744 distinct intervals
  for each of Union Delta Force, Multi-Plane and Starballs, Multiscreen's four-demo tour,
  Bilizir, DMA Is Back, Phenomena DNA, standalone TCB Replicants, Vectorballs'
  authored sequence, 3D DOC, DMA 3D, Coco, MegaTwist, TeamG1, Nonameno,
  Viva TCB, TCB Multi-Plane, DOM, Grodan and Cuddly's standalone menu. The two
  Vectorballs sphere variants were sampled separately. All twenty-two traces
  had zero intervals above 20 ms; detailed p95/max values appear below. Union intro
  and its Replicants screen also had 63-frame spot checks with zero intervals
  above 20 ms. Screenshots confirmed the effects, and thermal status remained
  0 during the earlier eight-scene run. The newer eleven scenes and both sphere
  variants and Union Delta Force were each inspected at a visible stage frame.
- Two Union Pixel tour runs each visited the full introduction, all eleven
  doors and loading/menu returns, then began another cycle at six seconds per
  screen. Their sparse 105-sample traces each spanned about 232 seconds and
  retained 6,510 distinct
  intervals (about 109 seconds of actual-present history). Their p95 values
  were 16.731 and 16.738 ms; six and four intervals exceeded 20 ms. The first
  maximum was 267.077 ms. The second, timestamped run peaked at 133.552 ms,
  and all four slow intervals aligned within about 60 ms of loader-to-screen
  handoffs into Wow, Hidden or Disk Copier. No sampled slow interval landed
  inside a stable screen. Thermal status stayed 0; three later process PSS
  snapshots ranged from 450,355 to 458,699 KiB. These are not peak-memory
  or battery measurements.
- Union's longer Pixel tour stayed one minute on each of its eleven screens,
  completed one hall cycle and entered a second. Its 500 PixelProbe windows
  spanned 1,108.04 seconds, retaining 31,000 distinct intervals and 517.61
  seconds of frame history. The p95 was 16.735 ms; six intervals exceeded
  20 ms, with a 133.623 ms maximum. Forty memory snapshots reached at most
  419,572 KiB process PSS and 190,832 KiB graphics memory. Thermal status
  stayed 0; these maxima are sampled, not absolute peaks.
- Cuddly's opt-in Pixel tour visited its introduction, thirteen doors, Reset
  and the final menu before ending normally. With `-allow-end`, PixelProbe
  retained 127 of 170 requested windows: 7,874 unique intervals covering
  about 132 seconds across 282 elapsed seconds. The p95 interval was 16.736 ms;
  four exceeded 20 ms, with a 200.307 ms maximum. Their presentation times
  preceded loader handoffs into Ehhh, Mega Scroller, Starwars and Megaball by
  24–68 ms. No sampled slow interval occurred inside an already running
  screen. Thermal status remained 0. The highest mean CPU Update and Draw
  submission times among 71 logged windows were 4.474 and 3.044 ms.
- A second Cuddly tour stayed one minute on each screen and reached Reset and
  the final menu in 905.59 seconds. Its 407 PixelProbe windows retained 25,234
  distinct intervals covering 421.22 seconds of frame history: p95 16.733 ms,
  three above 20 ms, maximum 83.465 ms. All three slow timestamps occurred
  near loader/menu handoffs. Twenty-one memory snapshots over its last ten
  minutes reached at most 595,053 KiB process PSS and 317,992 KiB graphics
  memory; thermal status stayed 0. These are sampled maxima, not peaks.
- Unmeasured optional combinations, varying device refresh rates, true memory
  peaks and battery consumption still need checks. Sparse presentation
  windows do not prove every cue boundary stays at 60 FPS.

| Pixel 10a DCK scene | Distinct intervals | p95 present interval | Maximum | Above 20 ms |
| --- | ---: | ---: | ---: | ---: |
| Union Delta Force | 744 | 16.720 ms | 17.699 ms | 0 |
| Union Multi-Plane | 744 | 16.717 ms | 16.831 ms | 0 |
| Union Starballs | 744 | 16.718 ms | 16.895 ms | 0 |
| Multiscreen tour | 744 | 16.736 ms | 17.267 ms | 0 |
| Multiscreen TCB canvas isolated | 744 | 16.740 ms | 16.879 ms | 0 |
| Bilizir | 744 | 16.719 ms | 17.168 ms | 0 |
| DMA Is Back | 744 | 16.728 ms | 16.901 ms | 0 |
| Phenomena DNA | 744 | 16.712 ms | 16.919 ms | 0 |
| TCB Replicants | 744 | 16.768 ms | 18.365 ms | 0 |
| Vectorballs authored sequence | 744 | 16.780 ms | 17.072 ms | 0 |
| 3D DOC main | 744 | 16.764 ms | 16.897 ms | 0 |
| DMA 3D | 744 | 16.783 ms | 16.944 ms | 0 |
| Coco main | 744 | 16.772 ms | 17.654 ms | 0 |
| Coco exact CRT intro and handoff | 744 | 16.741 ms | 16.941 ms | 0 |
| MegaTwist main | 744 | 16.761 ms | 16.923 ms | 0 |
| TeamG1 main | 744 | 16.736 ms | 16.935 ms | 0 |
| Nonameno main | 744 | 16.748 ms | 16.970 ms | 0 |
| Viva TCB | 744 | 16.772 ms | 17.088 ms | 0 |
| TCB Multi-Plane | 744 | 16.762 ms | 16.989 ms | 0 |
| DOM Intro | 744 | 16.758 ms | 16.904 ms | 0 |
| Grodan | 744 | 16.755 ms | 18.022 ms | 0 |
| Cuddly menu, CRT off | 744 | 16.762 ms | 16.926 ms | 0 |
| Cuddly menu, CRT on | 744 | 16.726 ms | 16.874 ms | 0 |
| DCK saved CRT/water/lens composition | 744 | 16.737 ms | 16.956 ms | 0 |
| DCK progressive path laboratory | 744 | 16.742 ms | 16.878 ms | 0 |
| Vectorballs sphere, 144 balls, published APK | 744 | 16.776 ms | 17.594 ms | 0 |
| Vectorballs sphere, 4,096 points, batched | 744 | 16.745 ms | 16.961 ms | 0 |

The Pixel's physical landscape display was 2,424×1,080. The main logical
surfaces measured here include 768×540 for Cuddly, 768×536 for Union and
800×600 for Multiscreen. Single `dumpsys meminfo` snapshots reported process
PSS / graphics memory of 249,775 / 121,068 KiB for Cuddly's intro,
276,724 / 138,044 KiB for Union Starballs and 263,663 / 147,364 KiB for
Multiscreen. These snapshots are not peak-memory measurements. Android reported
thermal status 0 after the sampled runs.
The Vectorballs authored sequence additionally used 198,956 KiB process PSS
and 101,156 KiB graphics memory in one snapshot. Its optional object previews
were also inspected on Pixel 10a after their exact desktop GPU comparisons:

| Vectorballs Pixel preview | Distinct intervals | p95 | Maximum | Above 20 ms |
| --- | ---: | ---: | ---: | ---: |
| Cube, edges, six segments | 822 | 16.736 ms | 16.866 ms | 0 |
| Pyramid, surface, six segments | 808 | 16.730 ms | 16.912 ms | 0 |
| Plane, six segments | 806 | 16.745 ms | 16.938 ms | 0 |
| Flag, twelve segments | 790 | 16.733 ms | 16.904 ms | 0 |

These are approximately 13-second presentation samples, not long-run memory
or battery measurements. Thermal status was 0 after the four previews. Both
six- and twelve-segment flags were inspected; six leaves individual balls more
legible while twelve reads as a denser waving surface.
After the CRT migration, a separate 744-interval MegaTwist run reported p95
16.764 ms, maximum 16.933 ms and zero intervals above 20 ms. Cuddly's CRT-on
run above was launched with the opt-in Android extra and sampled for 12.4
seconds. The menu's button was also toggled on both Android app variants on
the Pixel; the touch controls remained outside the CRT material.
After the spherical-cloud update, another Vectorballs authored-sequence run
reported 744 intervals, p95 16.776 ms, maximum 16.961 ms and none above 20 ms.
Its 144-ball sphere had another zero-slow-interval trace with the recommended
20-pixel sprite. Before batching and perspective/rear-half options, 217 of 744
intervals in the 4,096-point variant exceeded 20 ms and p95 was 33.411 ms;
after the change, zero exceeded 20 ms. At that density the output reads as a
point surface rather than distinct vectorballs. A further 372-interval sample
from the published APK had p95 16.769 ms, maximum 16.903 ms and zero intervals
above 20 ms at that density. The recommended 144-ball published APK retained
zero slow intervals over 744 samples. Its foreground window still reports
`KEEP_SCREEN_ON` and Android remained awake with a 30-second screen timeout.
After the palette-atlas migration, the published DCK APK was reinstalled and
its recommended 144-ball sphere inspected on Pixel. A fresh 744-interval trace
reported p95 16.773 ms, maximum 16.897 ms and none above 20 ms.
After direct indexed projection, the next APK yielded 744 intervals for the
authored sequence (p95 16.765 ms, maximum 17.022 ms) and 744 for the inspected
144-ball sphere (p95 16.795 ms, maximum 17.051 ms). Neither trace exceeded
20 ms.
The updated Bilizir DCK APK was also reinstalled and visually inspected. Its
default logo-warp/seamless-scroll mode yielded 744 intervals, p95 16.763 ms,
maximum 17.015 ms and zero intervals above 20 ms.
TeamG1's updated DCK APK showed its intro and main stage on Pixel. A focused
744-interval presentation run had p95 16.736 ms, maximum 16.930 ms and zero
intervals above 20 ms.
Coco's updated DCK APK was inspected on Pixel after isolating its CRT source
surface. A 744-interval main-stage run had p95 16.731 ms, maximum 16.898 ms
and zero intervals above 20 ms; it does not measure intro-only GPU cost.
With the authored source origin restored, another 744-interval run across the
intro and main handoff had p95 16.741 ms, maximum 16.941 ms and zero above
20 ms. One process snapshot showed 215,793 KiB PSS and 116,028 KiB graphics
memory; thermal status was 0. The sample is not a peak-memory measurement.

The common `scrolling.New` repeat renderer already has coverage for short
horizontal and vertical messages, gap boundaries and mixed-font controls.
`RingConfig.SeamlessSeed` additionally fills the initial slot bank for a short
recycled message. It is opt-in so existing Cuddly and Union seed timing stays
unchanged; its cursor/slot policy has a pure test and a compiled image-backed
regression test.

## Direct source-to-DCK frame sweep (2026-09-27 to 2026-09-28)

`cmd/fidelity` compared preserved Go revisions with published DCK revisions
at ticks 0, 1, 60, 240, 600, 1,200, 2,400 and 4,800. Audio was disabled and
both captures used one deterministic 60 Hz clock. Thirteen of the seventeen
paired productions were exact at all eight sampled ticks. This is complete
RGBA-canvas evidence at those ticks, not proof for intermediate frames or
Atari ST hardware. Cuddly's row covers its standalone Go menu only; Union
has no preserved Go production for this comparison. FR-010 and Second Reality
remain outside the full-screen audit.

| Production | Original → DCK revision | Sampled result |
| --- | --- | --- |
| 3D DOC | `914a84a` → `9cfae0e` | Exact. |
| Bilizir | `4720c97` → `94852f4` | Exact with original-logo and original-scroll-reset options; the default DCK variations remain available. |
| DMA 3D | `e161039` → `37b6f46` | One pixel differs by one channel level at tick 600; all other ticks exact. |
| DMA Is Back | `b47c293` → `25b9fdd` | Exact with historical cube transitions selected. |
| Coco is the best | `0a9b678` → `7ab225b` | Exact at the eight shared checkpoints and the additional intro tick 241. |
| Cuddly menu | `556d023` → `3cb4f30` | Exact for the standalone menu. |
| DOM intro | `a02358c` → `f7d89e8` | Exact. |
| Multiscreen | `2cd2bc1` → `0cdca5a` | One and seven pixels differ at ticks 600 and 1,200; six other shared checkpoints exact. A later tick 9,600 differs by ten pixels. |
| Vectorballs | `22ccd09` → `cd9a23c` | Exact for the authored sequence; optional objects have separate before/after checks. |
| Grodan | `5338712` → `522432b` | Exact. |
| MegaTwist | `df3ebf3` → `6d065d9` | Two or three pixels differ at ticks 60, 240, 600 and 1,200; four other ticks exact. |
| Nonameno | `e755a34` → `0865717` | Exact after the probe freezes `audio.Now()` as well as `time.Now()`. |
| Phenomena DNA | `8759087` → `0666ad7` | Exact. |
| TCB Multi-Plane | `2fa1b3f` → `a6de271` | Exact; a separate 21-frame check covers wave joins and reset through tick 9,600. |
| TCB Replicants | `1e5e55e` → `554080d` | Exact. |
| TeamG1 | `83e6452` → `91e598c` | Raw intro glyph surface exact at five ticks; the final intro differs because the DCK CRT is deliberately flat. Main scene exact at 1,200 and later sampled ticks. |
| Viva TCB | `f4b985e` → `a2aa151` | Exact, including tick 0 after preparing the logo formation's initial pose. |

Coco's raw intro surface is byte-identical before CRT at ticks 0, 1, 60, 240
and 241. A GPU coordinate probe found that its original CRT source begins at
texture X=4, while the previous DCK copy began at X=0. The historical shader
uses that coordinate directly. `CRTOverlayConfig.SourceOrigin` now prepares a
bounded independent source at any specified pixel origin, with no extra copy
for the default zero value. `presets.CocoCRTOverlay()` sets X=4; Coco no longer
owns a separate intro-strip surface. All nine full 800×600 source-to-DCK
captures through tick 4,800 now match pixel for pixel. The opt-in normalized
CRT variant still produces its previous five sampled outputs unchanged.
TeamG1's raw intro source also matches at ticks 0, 1, 60, 240 and 600. Its
authored 640-pixel insertion edge fixes the previously delayed scene handoff,
while the flat CRT intentionally prevents clipping the font's outer rows.
The remaining sampled full-frame differences in DMA 3D, MegaTwist and
Multiscreen stay visible in reports rather than being hidden by a tolerance.
The former Multiscreen captures differed by 182 and 146 pixels at ticks 600
and 1,200. The `-inspect-multiscreen-tcb` option captures the retained TCB tile
and glyph positions inside the actual tour. All 30 glyph positions match, and
removing only that text makes both transitions exact. The difference follows
the tile's GPU storage: giving just the TCB tile an independent texture through
`SceneTourConfig.UnmanagedMask` leaves one and seven differing full-frame pixels
at those ticks. Eleven checkpoints from 0 through 9,600 are exact at eight
ticks; tick 9,600 keeps the same ten-pixel difference as before. Applying the
storage change to all four tiles instead introduced 399 differences at 9,600,
so the other three retain their default allocation. Pixel 10a presentation
over 744 intervals through the TCB transition had p95 16.740 ms, maximum
16.879 ms and none above 20 ms; one memory snapshot showed 273,005 KiB PSS
and 153,808 KiB graphics memory, with thermal status 0. This short sample
does not establish long-run battery or peak-memory behavior.
The shared projected transport exposes its borrowed, depth-sorted positions
through `ProjectedController` for inspectors and future authoring controls.
From this module, reproduce the tile captures and position files with:

```sh
go run ./cmd/fidelity -demo go-multiscreen -frames 600,1200 \
  -inspect-multiscreen-tcb -out /tmp/dck-fidelity
```

The command exits nonzero while the documented pixel differences remain.

## Complete Cuddly and Union migration sweep (2026-09-28)

The collection extension of `cmd/fidelity` compares entire native canvases for
each screen, menu and door loader, with no excluded rectangle or pixel tolerance.
The [record](fidelity/collections-20260928.json) contains 607 exact comparisons
in 60 cases: 289 Cuddly frames and 318 Union frames. Both archived revisions use
the same local DCK checkout. This proves those sampled native migrations;
it does not claim original Atari hardware parity or unsampled behavior.

| Collection | Native reference → candidate | Scope |
| --- | --- | --- |
| Cuddly | `f3a921e` → `6f01eb4`, with DCK `23599dc` | Fifteen screens including intro and Reset; idle menu; all thirteen door loaders; editable Megaball input sequence. |
| Union | `f12b045` → `77b56b5` | Introduction, eleven screens, idle menu and all eleven door loaders; moving Hidden pointer; five TNT3 objects; ten Starballs populations; TNT2 speed/direction controls; Replicants speed changes; Disk Copier start/stop stages. |

Replicants uses `8a2fcb3` as its explicit reference to retain the intentional
Atari-based trajectory correction. Delta Force synthesizes its real YM track
without an audio device and feeds the three register levels to the ordinary
scene, rather than comparing frozen music-driven sprites. Core screen frames
include startup and late playback through tick 9,600; loader captures bracket
counter, blank and reveal boundaries. Spreadpoint additionally covers its final
card and three staggered main-effect entrances.

The first complete sweep exposed 29 differing Cuddly frames in six screens.
Their extracted presets had changed authored linear sampling to nearest:
Spreadpoint's initial inner logo, DOC's inner raster, Fullscreen's logo/raster
band, Starwars' sprite train, DNA's ribbon and sampled logo rows, and Reset's
raster pairs. Restoring those parameters removes every sampled difference
without adding screen-local rendering logic. The generic effects, geometry,
clock order and owned surfaces stay unchanged. Cuddly now pins the published
corrected DCK revision; its obsolete per-frame Spreadpoint filter setter is
removed. A separate run of its actual capture command with `GOWORK=off` and
no local replacement matched 60 complete frames across the six screens, from
the same source tree committed as `8c979b0`.

Reproduce a screen or input sequence with an explicit native reference:

```sh
go run ./cmd/fidelity -demo go-cuddlymenu -screen spreadpoint \
  -reference f3a921e -frames 0,1,723,724,725,1528,2008,2676,4800,9600
go run ./cmd/fidelity -demo go-uniondemo -screen tnt3 -scenario controls \
  -reference f12b045 -frames 0,1,599,600,601,1200,1800,2400,3000
```

Both corrected Cuddly Android variants were built and installed on Pixel 10a.
The complete app's six affected main scenes were visually inspected and sampled
after their ordinary introductions/warmup. Each 12.4-second presentation sample
contains 744 distinct intervals; all six have zero intervals above 20 ms.

| Corrected Cuddly screen | p95 | Maximum | PSS snapshot (KiB) | Graphics snapshot (KiB) |
| --- | ---: | ---: | ---: | ---: |
| Fullscreen | 16.729 ms | 16.848 ms | 274,448 | 144,740 |
| Starwars | 16.680 ms | 16.813 ms | 258,148 | 129,228 |
| 3D DOC | 16.720 ms | 16.894 ms | 263,627 | 135,452 |
| DNA | 16.685 ms | 16.817 ms | 242,522 | 123,228 |
| Spreadpoint | 16.704 ms | 16.837 ms | 256,577 | 125,536 |
| Reset | 16.720 ms | 16.996 ms | 322,329 | 173,980 |

Thermal status was 0 after every sample. These are process snapshots and short
presentation traces, not memory peaks or battery measurements. The filtering
correction adds no surface, point bank or simulation loop. The complete app was
restored to its ordinary introduction after measurement. The JSON record also
contains the installed APK hashes, warmup ticks and published dependency check.

## Unified scrolling entry points (2026-09-28)

The common `scrolling.New` constructor now includes `DualProfiled`, `Caption`
and `Reveal` alongside the existing fifteen specialized transport choices.
These bind the existing complete engines rather than introducing another
renderer. Every recycled ring in Cuddly/Union, both collections' loaders,
Cuddly Big Sprite's separated font lanes, Starwars' back-mask-front text and
Union TNT3's caption now enter through this facade. The former direct
constructors remain available for compatibility. Bitmap layout helpers,
precomputed DNA artwork and explicit history inputs remain separate asset or
composition operations.

`DrawOffset` places regular/recycled text and synchronized lanes without another
image or a transport reset. Borrowed controllers expose cue cursors or permit
per-lane masks; the facade still owns their Update. Reveal samples its caller's
clock in Update so high refresh rates cannot accelerate the loading text.

The [facade record](fidelity/scroll-facade-20260928.json) stores another 607
unmasked, exact complete-frame comparisons: Cuddly `8c979b0` → `aa94f97` and
Union `77b56b5` → `30cae2a`, including all 24 door loaders and interactive
screen fixtures. DCK and both production test/vet suites pass against published
DCK `e9cb323`. The native `dck_scroll_facadecheck` additionally compares
independent font dimensions, horizontal/vertical offsets, output-pass order,
caption line transitions and reverse-column cell reveals. Repeated Draw does
not change the sampled cursors, poses or cue clock.

Both Cuddly APKs and Union were rebuilt and installed on Pixel 10a. Active
Starwars and TNT3 each yielded 744 distinct presentation intervals with none
above 20 ms: p95/max 16.720/16.969 ms and 16.776/17.608 ms respectively.
The two main scenes were visually inspected; these short runs do not establish
memory peaks or battery use.

## DNA composition through the common facade (2026-09-28)

`CuedSlices` combines `SliceProgram`, configurable pause/rotation cues and
borrowed rotating-glyph artwork through `scrolling.New`. Phenomena and its
Multiscreen panel now use it for Update and Draw. Their explicit 320-tick
pre-roll uses the borrowed controller; late film binding retains every cursor
and cue state. Binding rejects a film whose frame count differs from the clock,
and premature drawing reports an unbound-artwork error. Both old direct APIs
remain available. Pure tests cover 5,000 updates, independent control clocks,
odd glyph widths and zero steady-update allocations.

Cuddly DNA's opposing colored faces and Spreadpoint's delayed lane now select
`Output.Feedback` on the same facade. Their message, font and profile remain
scene data; DCK presets configure the insertion, direction, gradient, position
and authored phase. Source surfaces stay 320×25 and replace the preceding
scene-owned surfaces. Phase callbacks are cached at setup and Update; Draw
cannot advance history or reevaluate a live callback.

The [DNA facade record](fidelity/dna-facade-20260928.json) contains 42 exact,
unmasked complete-frame comparisons: Phenomena `03abf51` → `aa20b19`,
Multiscreen `0cdca5a` → `23928fd`, and Cuddly DNA/Spreadpoint `aa94f97` →
`f811259`/`4756742`. Cuddly DNA retains its earlier texture upload order,
removing a former forty-pixel, one-channel-level filtering difference at tick
1,200. Separate source-formula comparators verify eleven DNA images per
Phenomena variant through tick 48,000. The common GPU suite compares delayed
artwork binding, repeated draws and both feedback directions pixel for pixel.
DCK and the three production test/vet suites pass.

Phenomena, Multiscreen and both Cuddly Android variants were rebuilt and
installed on Pixel 10a. Four active-scene samples contain 744 distinct present
intervals each, with no interval above 20 ms:

| DNA scene | p95 | Maximum |
| --- | ---: | ---: |
| Cuddly DNA | 16.723 ms | 16.852 ms |
| Spreadpoint | 16.746 ms | 16.864 ms |
| Phenomena | 16.760 ms | 16.972 ms |
| Multiscreen's Phenomena panel | 16.759 ms | 17.488 ms |

The main scenes were visually inspected. The short presentation samples and
process-memory snapshots do not measure true memory peaks or battery use.

## Saved composition and progressive path edges

The authoring schema now stores ordered image passes on a single layer or on
the complete scene. CRT, water reflection and magnifier configurations compile
directly to the existing DCK components and `kit.Pipeline`. Pass windows can
fade and repeat from their first start with an explicit period; the source
still updates once per frame. CRT's opt-in transparent curved edge permits
alpha layers without changing existing productions' black-edge default.
`Project.PostSurfaceBytes()` estimates owned post-processing surfaces before
compilation, and the schema limits them to 64 MiB. The saved 640×360
`examples/authoring/projects/composed-effects.json` uses 3.52 MiB for these
surfaces and runs unchanged in the desktop and Android effects-lab hosts.
Native GPU checks verify inactive/second-cycle lens windows, bounded lens
coverage, cumulative title passes and correct partial-alpha CRT fades.

The effects laboratory's trajectory scroll now uses the common automatic
repeat transport instead of a character-index window. `PathConfig.Extrapolate`
continues open endpoint tangents; `Viewport` clips actual glyph pixels. The
historical `Clip` option keeps its whole-origin rule for existing callers.
A native GPU check moves an eight-pixel solid glyph through both edges and
verifies every one-column reveal/removal step. The updated Pixel display was
visually accepted by the user, including the recurring lens. A 1,800-tick
profile measured 59.99 updates/s and 59.92 Draw callbacks/s, with mean CPU
cost of 12.9 µs per update and 1.68 ms per Draw submission. Its 744-interval
presentation trace had no interval above 20 ms; one snapshot showed 207,983
KiB PSS, 106,392 KiB graphics memory and thermal status 0.

The repeating saved composition separately measured 59.97 updates/s and 59.90
Draw callbacks/s over a 900-tick profile, with 32.2 µs mean Update CPU and
1.12 ms mean Draw submission CPU. Its 744 presented intervals had p95
16.737 ms, maximum 16.956 ms and none above 20 ms. One process snapshot
reported 205,304 KiB PSS and 107,056 KiB graphics memory; thermal status was
0. These short runs are not peak-memory, GPU-execution or battery measurements.

## Extraction order and acceptance

1. Keep the scrolling facade cohesive: DMA, Coco and MegaTwist share the
   proportional scanline transport; 3D DOC uses ordered row bands; DMA 3D
   and Replicants share fixed-cell row-source/column-destination sampling.
   Each sampling rule is named and compared against original pixels.
2. Extract projected fields and sprite ensembles with explicit spawn, depth,
   trajectory, timing and painter configuration. Pixel/sprite/streak choice must
   preserve the same simulation, not force one visual behavior on every screen.
3. Extract raster/mask materials and cue programs. Keep event clocks, draw order,
   alpha blend and handoff behavior explicit so the future editor can serialize
   common cases while Go callbacks remain available for special cases.
   The `authoring` sprite layer serializes `CuedFormation` pose, independent
   origin/spacing/arc tracks and overlapping harmonic cues, or absolute
   per-sprite `KeyframedFormation` paths. The saved formation example renders
   through the ordinary desktop host without scene-specific Go.
4. Continue measuring combined scene classes on Pixel. The ledger now samples
   scanline/warp, projected field, masks and multi-layer compositions, with
   Cuddly CPU timings, twenty-two baseline presentation traces across twenty
   catalog scenes and two optional sphere variants, one-minute-per-screen
   Cuddly and Union tours, and focused cube/pyramid/plane/flag runs. Unmeasured
   optional combinations, varying refresh rates, true memory peaks and battery
   consumption still need device evidence.

The 2026-09-27 host-source pass found no further active screen-local sine/cosine,
triangle/shader submission or pixel recoloring loop in the audited DCK versions
and Cuddly/Union screens. Remaining loops prepare authored assets, compose
ordered layers, process menu/input state or project points through the shared
components. DOM's one-time repeated raster image and TeamG1's missing-asset
checkerboard remain asset preparation, not animated effects. This source pass
does not replace complete original-versus-port visual review.
The follow-up check of DMA Is Back and TeamG1 confirms that their intro/music
handoffs already use `timeline.IntroHandoff`; the remaining conditionals only
select the active scene, own resources or place layers. No extra local cue
controller was found in either DCK version.
A separate AST inventory found 69 non-test loops in the audited DCK sources
and Union screens (excluding command/mobile hosts, FR-010 and Second Reality).
The live loops inspected in Cuddly's hall integrate its collision-aware menu
character, Grodan's loops preserve authored layer order, and Union Delta
dispatches events produced by DCK's shared word-motion controller. Other loops
load or prepare artwork, resolve input, or close borrowed resources; none is
another local pixel, sprite-trajectory or shader engine. This narrows the
remaining work to fidelity and runtime evidence rather than a known missing
effect family.

## Catalog acceptance review (2026-09-28)

The [requirement-by-requirement review](CATALOG_ACCEPTANCE.md) concludes the
reusable native effect-engine extraction for the eighteen in-scope repositories.
Its [revision and check record](fidelity/catalog-acceptance-20260928.json) refreshes
all twenty demo modules' test/vet checks against DCK `02f4f62`, the 336 visible
font/mode combinations and a source inventory covering 81 files and 80 loops.
The larger loop count includes collection application/menu hosts and controls;
it does not use the earlier 69-loop inventory's narrower file selection.
All eighty tracked original Go files in the seventeen paired source snapshots
remain present and unchanged. Local Git rules exclude hosting exchange folders.

The final composition regression reproduced a missing-artwork failure which
previously returned nil from feedback Update. DCK now returns that source error
immediately, preserving the last valid history and phase sample. Native GPU
checks retain a populated history after its borrowed film is closed; a separate
test covers an exceeded glyph budget before phase evaluation. Successful source
rendering keeps the existing sampling, draw order and resources.

The review retains the map's known sampling differences and runtime limits.
It does not assert continuous Atari-hardware equivalence, true memory peaks,
battery measurements or completion of the separate graphical editor.

For any migration, compare deterministic complete-frame captures before and
after at startup, state changes, text/texture wrap and late playback. Keep the
original packages intact. Use bounded reusable GPU surfaces and geometry arrays;
none of these effects should allocate a message-width texture during Draw.
