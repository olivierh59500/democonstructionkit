# Effect map for the native demo catalog

This map covers every DCK production under `demos/` except FR-010 and Second
Reality. The screen and workspace audits began on 2026-09-23; the wrap/train
inventory was checked across all 20 DCK repositories on 2026-09-25. The
module-wide build/vet check was refreshed on 2026-09-26. A complete component
owns its transport,
animation state, geometry and rendering resources. The production supplies
assets, messages, presets, input and scene order. A shared draw helper alone is
not counted as a complete effect.

`Shared` names a complete component already used by that screen. The third
column distinguishes intentional production composition from reusable logic
still to extract. Authored artwork, input and layer order do not become DCK
effects simply to shorten a source file. The entries are acceptance work, not
claims of pixel-perfect or Pixel-device verification.

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
| `dma-is-back` | Cube, intro feed, CRT and scanline scroll use their dedicated components; timed cues are separate work. |
| `go-cocoisthebest` | Sprite grid/translation already uses `sprites.Group`; scanline scroll and cube are separate families. |
| `go-cuddlymenu` | Colorshock's orbit/table clock, Ehhh raster train, LED backdrop/gradient wraps and harmonic letters, Big Sprite front/back `sprites.AxisFlip`, Big Sprite/Fullscreen/Digi Weave groups, and Mega Scroller's directional entrance bounce now use shared controllers. |
| `go-dom-intro` | Its sampled background uses `VerticalStripTrain`, its font-size program uses `scrolling.Config.SizeBank`, and animated stars use `sprites.AnimatedField`. |
| `go-fr010` | Software-rendered parts were searched; no direct GPU image-train or simple threshold-wrap equivalent was migrated. |
| `go-multiscreen` | Embedded Viva title rasters and TCB mountain strips use the same DCK presets as their standalone versions. The camera tour is a separate director effect. |
| `go-secondreality` | Indexed software effects were searched separately; their palette/VRAM clocks are not interchangeable with these Ebitengine image controllers. |
| `go-uniondemo` | Intro harmonic cell waves, Replicants raster trains, Beat Dis/Wow/TNT2/Level 16 wraps and Disk Copier's six-strip raster are shared. Beat Dis letters use the harmonic formation with a common wobble and individual phases. |
| `go-vectorballs` | Ball projection, morphing and reflection are different shared families; no direct image train/wrap candidate found. |
| `grodan-kvack-kvack-demo` | Four scroll lanes use `Config.Ribbon`, three `SurfaceLayer` compositions, `GatedBackgroundPair` owns the scenery, and the sprite chain uses `motion.HarmonicFormation`. |
| `megatwist` | Its multi-frequency, clamped sprite motion now uses the same formation family; `sprites.GlowPainter` owns the configurable halo passes. |
| `nonameno-demo` | Its text pages use `motion.GlyphPageCycle` and `sprites.GlyphPages`; the bottom ribbon uses the common scrolling transport and harmonic mode. |
| `phenomena-dna-scroll-intro` | Raster-bar thresholds trigger scene-state changes; treating them as a periodic wrap would change the sequence. |
| `tcb-multi-plane-3d-scroller` | All 32 mountain strips now use the relative `WrapBank` preset also used by Multiscreen. |
| `tcb-replicants-demo` | Layered stars, coupled logos, stepped splash and the foreground sprite pair use complete DCK components. |
| `teamg1-demo` | Circular sprite formation already uses `sprites.Group`; other timed presentation cues remain local. |
| `viva_tcb` | Paired title rasters now use the shared `WrapBank` preset; harmonic logos remain a separate formation. |

## Individual demos and intro screens

| Production / screen | Shared today | Remaining responsibility or extraction |
| --- | --- | --- |
| 3D DOC | Atlas, `Config.RowBands` using `composite.RowWarp`, `PerspectiveCheckerboard`, `ProjectedBallTrain`, `timeline.IntroHandoff` | Whole-scene transform remains scene composition data. |
| Bilizir | Atlas, scrolling with a relative `motion.WrapBank` and `scrolling.CyclicWindow` over two bounded virtual text copies, two `StripWarp` instances on one `motion.WarpTableClock` with independent phase/gain variations, `motion.HarmonicTransform` logo paths on one `motion.WaveClock`, batched `SolidCubeTrain`, WaterReflection and `CopperBars` | Authored text and artwork stay production data; the DCK text loop bridges its former blank reset while the original source remains intact. |
| DMA 3D | Atlas, `Config.RowColumn`, `sprites.BatchedSolidField` with timed wrap, and complete `effects.MorphingMesh` with ordered per-face materials | Artwork, text, soundtrack and scene layer order stay production data. |
| DMA Is Back, intro | `Config.Feed`, configurable CRTOverlay, `timeline.IntroHandoff` | Message and scene materials stay production data. |
| DMA Is Back, main | `Config.Scanline`, ImageGrid, `motion.TrajectoryClock` with a `NestedOrbit` logo path, JellyCube, shared fade/music cue | Other whole-scene timed layers can use the general cue director. |
| Coco, intro | `Config.Feed`, configurable CRTOverlay, immediate `timeline.IntroHandoff` music cue | Scene materials stay production data. |
| Coco, main | `Config.Scanline`, `SolidCubeTrain`, sprites.Group grid/translation, shared font metrics, `RotozoomBackground` source quad and complete `CopperTitleBand` with a retained surface | Authored art and scene layer order. |
| DOM intro | Atlas, `scrolling.Config.SizeBank`, `motion.ScaledTextClock`, `VerticalStripTrain` background, `RasterOverlay` copies, `effects.Mask` and `sprites.AnimatedField` stars | Authored text and scene layer order. |
| Multiscreen, four embedded productions | `Config.Scanline`, recurrent `SolidCubeTrain`, `RotozoomBackground`, `sprites.Group` recurrent translation and direct `CopperTitleBand` in Coco, `RotozoomBackground` in Viva, Phenomena's main-only `scrolling.SliceProgram`, `motion.RecurrentRowWave` with its owned scene clock, `composite.ScalarStagePainter` HSL materials and reusable gradient/mask assets, projected-plane engines, atlas recipes, Viva `RasterTitle`, `scrolling.Config.Pseudo3D` text banks and `sprites.RecurrentFormation` logos, complete TCB `effects.MultiPlaneScene` | Authored images, messages, panel sizes, placement and scene order remain in the host. No local trigonometric path, mesh submission or scrolling renderer remains in its DCK panels; the standalone Phenomena intro is preserved separately. |
| Multiscreen, camera tour | Complete `composite.SceneTour` over `motion.CameraTour`: ten held/eased poses, direct fixed views, continuously updated sources, masked retained transition canvases, one-pass shader and fallback | Authored screen sources, world placement and music stay production configuration. |
| Vectorballs | Projected shape factories, `sprites.Projector`, source-order `geometry.RotateXYZScaled` model matrix, reflection and complete `geometry.PointSequence` with point morphing, sine grid, rotors, Y orbit and bounce | Artwork, authored shape/action data and layer placement stay production parameters. |
| Grodan | Atlas, four `scrolling.Config.Ribbon` lanes, three `SurfaceLayer` raster compositions, `GatedBackgroundPair` and harmonic `sprites.Group` | Authored art and scene layer order. |
| MegaTwist, intro | `Config.Feed`, atlas, `timeline.HoldRamp` 90-tick splash | Authored intro image and CRT material remain scene data. |
| MegaTwist, main | `Config.Scanline`, independent `ScanlineBackground` and DisplacementPrograms, harmonic `sprites.Group` and `sprites.GlowPainter` | Authored artwork and layer order remain scene composition. The former black transition overlay was visually inert after the splash reset and has been removed from the DCK version. |
| Nonameno, stars | `sprites.ProjectedField`, editable radial pattern and vector pixel/trail material | The star field is complete. |
| Nonameno, text pages and bottom scroll | Atlas, `motion.GlyphPageCycle`, three delay generators, `sprites.GlyphPages`, `scrolling.Config` with `HarmonicSineWith` and editable presets | Page text and bottom message remain production data. |
| Phenomena DNA intro | Atlas, `scrolling.BitmapPage` retained intro stills, DNAFrames, `scrolling.SliceProgram` with two-pixel insertion, loop-start and configurable control cues, `motion.RecurrentRowWave` baseline and owned scene clock, `motion.GravityBounce` photon path, `motion.WrapBank` hue clock, `timeline.ScalarStages` intro/outro thresholds, `composite.ScalarStagePainter` ordered reveal/fade and HSL main materials, and reusable gradient/silhouette/inverted-font assets | Live DNA scroller, input/music and final black mask remain production composition. |
| TCB multiplane | Complete `effects.MultiPlaneScene`: projected scrolling, font-independent forms, snapped `composite.Bands` mountains, per-row `composite.ProfileImage` logo, `sprites.AxisFlip` emblem | Text, artwork and music remain production parameters. |
| Replicants | Atlas, `Config.RowColumn` with variable-speed controls, `sprites.AnimatedField` stars, `CoupledLogoPair`, `BlockReveal` and `sprites.Train` foreground pair with owned speed-scaled phase | Authored text, input and scene order. |
| TeamG1, intro | `Config.Feed` with progressive right-edge entry, atlas, flat TimedCRTOverlay, `timeline.IntroHandoff` | Message and materials remain production data. |
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
| Menu | TileAlphabet, `sprites.Atlas` image banks, `sprites.FrameSequence` character animation, `sprites.FormationCarousel` with seven compiled formula modes, `composite.CachedTileParallax`, `motion.CameraFollow`, scrolling | Door/input semantics and map content stay local. |
| Loader | Bitmap font recipes, scrolling, `timeline.Countdown`, `timeline.CueClock` overlapping fade/hold windows and `timeline.CueRamp` gain applied by the playback host | Initial pre-render, text and layer placement remain authored. |
| Introduction | `composite.WaveChain` with an editable row-then-column logo preset, `ProfileStrips` with compiled `motion.WaveWrite` table, configurable sparkles and `timeline.StillThenMain` for the still/blank/music handoff | Authored artwork, star positions, music filename and layer order remain scene data. |
| Big Sprite | `sprites.ProjectedField` with strict-wrap vector lines, `scrolling.RingLanes` dual-font transport, `sprites.Group` Weave formation, `RasterOverlay` fill, `sprites.AxisFlip` front/back material on a `motion.TrajectoryClock` orbit | Authored layer placement and image order stay scene data. |
| Colorshock II | Background sampler, scrolling, editable `FormulaFormation` two-frequency orbit on `motion.TrajectoryClock` and `WrapBank` position-table clock | Source position samples and artwork remain screen data. |
| Ehhh | Scrolling, `composite.TableWarpLogo` over the compiled row profile shared with Digi, `sprites.Train` raster bars, `motion.TrajectoryClock` backdrop orbit, `motion.CuedWaveClock` lookahead/landing roller, `motion.WrapBank` inner strip and `RasterOverlay` middle-text fill | Authored art, text and layer order remain screen composition. |
| Mega Scroller | `composite.TiledWaveBackdrop` for the cached two-wave tile field, finite overlapping `composite.Background` mask copies, scrolling, directional `motion.BounceBank` text-surface transport and reusable `RasterOverlay` source-atop fill | Authored text, bar artwork and scene layer order remain production data. |
| Spreadpoint | `Config.Bands`, feedback DNA, atlas, a 20-ball `sprites.Group` with pixel-snapped formula, a dynamic `SurfaceLayer` for the raster-filled logo, `motion.PhaseSequence` with `motion.HarmonicTransform` for its orbit/zoom and `timeline.TintedCards` for the four stills and handoff cues | Authored card images, music filenames and layer order remain scene data. |
| Digi | Scrolling, complete `composite.TableWarpLogo` over compiled `motion.WaveWrite` rows and one rectified bounce shared with the Union logo, `sprites.Group` Weave formation, and independent `motion.WaveClock` scroll bounce | Authored artwork, text and layer order stay screen data. |
| LED Scroller | `composite.TiledWaveBackdrop` with retained scrolling source and explicit preload, cached bubble matrix, scrolling, `motion.WrapBank` gradient offset, harmonic `sprites.Group`, rectified `motion.WaveClock` and `palette.UniformGradient` 11-color raster material | Authored artwork and layer order remain scene composition. |
| 3D DOC | Scrolling, paired `composite.RowWarp` source-sampling programs, `PerspectiveCheckerboard`, `ProjectedBallTrain`, first-main-tick `timeline.IntroHandoff` cue and two ordered `RasterOverlay` source-atop passes for the inner text | Authored artwork, text and scene placement remain production data. |
| Fullscreen | `BackgroundLayer` velocity, `sprites.Group` Weave formation, `scrolling.Config.RingLanes` with paced vertical wrap and a bounded `SurfaceLayer` logo bar | Authored layer order and screen text. |
| Starwars | `Config.Crawl`, `scrolling.DualProfiledRing` with two independent fonts, segmented wave profile, source-in raster and exact first-frame filtering, `sprites.ProjectedField` with camera velocity/angle and depth shade, `sprites.SampledSpriteTrain` over one authored XY path, and `modulation.PeriodicDecay` backdrop flash | Authored text, image assets and layer placement remain scene data. |
| Knucklebuster | Scrolling and complete `sprites.LatchedOverlay` over `motion.LatchedTriggers`: seeded five-tick hit sampling, three held channels and four ordered sprite images | Authored artwork and message remain screen data; a separate signal preset supports music-driven variations. |
| DNA | FeedbackDNA, wave strips, complete `sprites.RotatingDiscCloud`, `composite.TwistingRibbon` with exact occlusion and two `composite.SampledRows` logo programs | Authored intro/music cue and layer order. |
| Megaball | Scrolling, rectified `motion.WaveClock` text bounce and `sprites.Group` with an ordered `motion.CoupledOrbitFormation` for two interleaved ball trains | The nine-value input mapping and authored labels remain screen data. |
| Reset | `Config.Slots`, scrolling, atlas, `timeline.StageSequence` with immediate events and cue windows, `composite.PairedRasterOrbit` for ordered pairs, `motion.WrapBank` for vertical raster fill and `motion.FormulaTrajectory` for the two-clock backdrop orbit | Source-atop compositing and layer order remain screen composition. |

## Union presentation units

| Screen | Shared today | Remaining responsibility or extraction |
| --- | --- | --- |
| Introduction | Image repetition on `motion.WrapBank`, `HarmonicCellWarp` with two active editable row/column wave banks, `motion.HarmonicTransform` logo path and bitmap recipes | Authored artwork, text placement and layer order remain screen composition. |
| Menu | Scrolling, background sampler, `sprites.Atlas` character frames, `motion.WrapBank` panorama, `motion.LinearTick` uncover wipe, `timeline.PacedIndex` palette/walk cycles, `motion.HoldBounce` logo and `motion.WalkParallax` hall/banner | Door navigation and authored layer order stay local. |
| Loader | Reveal, bitmap recipes and `timeline.CueClock` exact-duration transition | Playback cue and layer placement remain host composition data; retain column-major reverse-row glyph order. |
| Beat Dis | Background sampler with independent single-copy entrance, scrolling, `motion.WrapBank` wallpaper and pattern offsets, harmonic `sprites.Group` letters with global wobble | Layer order stays authored scene data. |
| Delta Force | YM register snapshots, `modulation.Change` and `Decay` for the three voice-triggered ball frames, scrolling, WaveStrips, `motion.BounceToggle` logo, `motion.WrapBank` gold fill, `motion.EnterHoldExit` handoff and one reusable `RasterOverlay` source-atop material for both texts | Authored text, images and layer order remain screen composition. |
| TNT Crew 3 | Complete `effects.SolidMeshCarousel` over `motion.ModelCarousel` and `geometry.OrderedEuler`, plus `scrolling.CaptionCarousel` for ordered slide/hold/exit text | Authored vertices, face groups, messages and input mapping stay production data. |
| Wow Scroller | Scrolling, `RasterOverlay` source-atop fill and `motion.WrapBank` paired panel offsets | The two 640×1235 images are single oversized draws naturally clipped by the 640×400 stage; no additional repeated-image transport is present. |
| Hidden | `sprites.DelayedTrail` over `PointHistory` and `timeline.PacedIndex` for the one-tick palette offset | Authored palette colors, crosshair and border clipping remain scene composition. |
| Starballs | Camera, scrolling, `sprites.MaskedProjectedField` with one projected population, two borrowed materials, editable mask paint, live count and depth opacity | Authored logo, scroll message and input mapping remain screen data. |
| Replicants | Atlas, two synchronized bitmap text windows on a strict `motion.WrapBank` clock, `RasterOverlay` fill, `motion.CuedFormation` letter paths, `sprites.Train` rasters driven by `motion.BounceBank` | Authored text, key mapping and layer order remain scene data. |
| TNT Crew 2 | Background sampler, BitmapText.DrawWindow on a strict `motion.WrapBank` text clock, `motion.WrapBank` three-layer parallax and mutable speeds | Per-key control mapping remains scene data. |
| Level 16 | Vertical scrolling, background sampler, `motion.TrajectoryClock` with a `NestedOrbit` ball path and two independent `composite.RasterOverlay` materials with exact water/raster wraps | Authored artwork and layer occlusion remain scene data. |
| Multi-Plane | Complete `effects.MultiPlaneScene` in native-stage mode, with Union's source-index phases, `composite.ProfileImage` row renderer and strict `sprites.AxisFlip` cycle | Artwork, text, soundtrack and door routing remain production data. |
| Disk Copier | Bitmap recipes, `sprites.Atlas` LCD regions, `motion.GatedWrapBank` three LCD clocks, `composite.WindowedImageBank` six-strip raster with its shared phase, `timeline.CueRanges` stages and `timeline.SteppedEnvelope` palette | Input/state program and LED layer placement remain scene composition data. |

## Verification ledger (2026-09-26)

- `GOWORK=off go build ./...` and `go vet ./...` passed in all 20 demo modules
  with their published dependency pins. FR-010 and Second Reality are included
  in these compatibility checks, not in the full-screen effect audit above.
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
- Pixel 10a (Android 17/API 37) was reconnected and sampled at 60 Hz with
  current DCK APKs. Cuddly's Big Sprite, DNA main stage, Mega Scroller, Reset,
  Starwars, Fullscreen and introduction reported 59.81–60.25 observed FPS in
  repeated 250-update windows. The largest measured CPU Update time was about
  5.43 ms (DNA main); Draw submission was at most about 0.34 ms. These host
  timings exclude asynchronous GPU work.
- The Go `cmd/pixelprobe` samples SurfaceFlinger's actual-present timestamps.
  Twelve history windows at two-second spacing yielded 744 distinct intervals
  for each of Union Multi-Plane and Starballs, Multiscreen's four-demo tour,
  Bilizir, DMA Is Back, Phenomena DNA and standalone TCB Replicants. All seven
  measured scenes had zero intervals above 20 ms; detailed p95/max values
  appear below. Union intro and its Replicants screen additionally had 63-frame
  spot checks with zero intervals above 20 ms. Screenshots confirmed the
  effects, and thermal status remained 0.
- The remaining catalog screens, longer playback, varying device refresh rates
  and battery consumption still need device checks. The sampled windows do not
  prove that every cue boundary or full megademo tour stays at 60 FPS.

| Pixel 10a DCK scene | Distinct intervals | p95 present interval | Maximum | Above 20 ms |
| --- | ---: | ---: | ---: | ---: |
| Union Multi-Plane | 744 | 16.717 ms | 16.831 ms | 0 |
| Union Starballs | 744 | 16.718 ms | 16.895 ms | 0 |
| Multiscreen tour | 744 | 16.736 ms | 17.267 ms | 0 |
| Bilizir | 744 | 16.719 ms | 17.168 ms | 0 |
| DMA Is Back | 744 | 16.728 ms | 16.901 ms | 0 |
| Phenomena DNA | 744 | 16.712 ms | 16.919 ms | 0 |
| TCB Replicants | 744 | 16.768 ms | 18.365 ms | 0 |

The Pixel's physical landscape display was 2,424×1,080. The main logical
surfaces measured here include 768×540 for Cuddly, 768×536 for Union and
800×600 for Multiscreen. Single `dumpsys meminfo` snapshots reported process
PSS / graphics memory of 249,775 / 121,068 KiB for Cuddly's intro,
276,724 / 138,044 KiB for Union Starballs and 263,663 / 147,364 KiB for
Multiscreen. These snapshots are not peak-memory measurements. Android reported
thermal status 0 after the sampled runs.

The common `scrolling.New` repeat renderer already has coverage for short
horizontal and vertical messages, gap boundaries and mixed-font controls.
`RingConfig.SeamlessSeed` additionally fills the initial slot bank for a short
recycled message. It is opt-in so existing Cuddly and Union seed timing stays
unchanged; its cursor/slot policy has a pure test and a compiled image-backed
regression test.

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
4. Measure combined scene classes on Pixel. The current three-cube benchmark
   measures that scene only; scanline, field/mask and multi-layer scenes need
   their own CPU/frame and logical-surface reports.

For any migration, compare deterministic complete-frame captures before and
after at startup, state changes, text/texture wrap and late playback. Keep the
original packages intact. Use bounded reusable GPU surfaces and geometry arrays;
none of these effects should allocate a message-width texture during Draw.
