# Effect map for the native demo catalog

This map covers every DCK production under `demos/` except FR-010 and Second
Reality. It was checked against the screen implementations and the detailed
workspace audits on 2026-09-23; the wrap/train inventory was checked across
all 20 DCK repositories on 2026-09-25. A complete component owns its transport,
animation state, geometry and rendering resources. The production supplies
assets, messages, presets, input and scene order. A shared draw helper alone is
not counted as a complete effect.

`Shared` names a complete component already used by that screen. `Extract next`
identifies remaining reusable logic and the parameters that must stay editable.
The entries are acceptance work, not claims that all screens are already small.

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
| `dma-3d` | Star field and mesh animation are depth/geometry effects, not an image train or threshold wrap. |
| `dma-is-back` | Cube, intro feed, CRT and scanline scroll use their dedicated components; timed cues are separate work. |
| `go-cocoisthebest` | Sprite grid/translation already uses `sprites.Group`; scanline scroll and cube are separate families. |
| `go-cuddlymenu` | Colorshock's orbit/table clock, Ehhh raster train, LED backdrop/gradient wraps and harmonic letters, Big Sprite front/back `sprites.AxisFlip`, Big Sprite/Fullscreen/Digi Weave groups, and Mega Scroller's directional entrance bounce now use shared controllers. |
| `go-dom-intro` | Both vertically wrapped raster/background offsets use `WrapBank`; star atlas choreography remains local. |
| `go-fr010` | Software-rendered parts were searched; no direct GPU image-train or simple threshold-wrap equivalent was migrated. |
| `go-multiscreen` | Embedded Viva title rasters and TCB mountain strips use the same DCK presets as their standalone versions. The camera tour is a separate director effect. |
| `go-secondreality` | Indexed software effects were searched separately; their palette/VRAM clocks are not interchangeable with these Ebitengine image controllers. |
| `go-uniondemo` | Intro harmonic cell waves, Replicants raster trains, Beat Dis/Wow/TNT2/Level 16 wraps and Disk Copier's six-strip raster are shared. Beat Dis letters use the harmonic formation with a common wobble and individual phases. |
| `go-vectorballs` | Ball projection, morphing and reflection are different shared families; no direct image train/wrap candidate found. |
| `grodan-kvack-kvack-demo` | Its phased twelve-sprite chain now uses `motion.HarmonicFormation` through `sprites.Group`, with independent phase clocks and a bouncing amplitude envelope. |
| `megatwist` | Its multi-frequency, clamped sprite motion now uses the same formation family; `sprites.GlowPainter` owns the configurable halo passes. |
| `nonameno-demo` | Text-page glyphs use staggered enter/exit and depth tweening, not cyclic image transport. |
| `phenomena-dna-scroll-intro` | Raster-bar thresholds trigger scene-state changes; treating them as a periodic wrap would change the sequence. |
| `tcb-multi-plane-3d-scroller` | All 32 mountain strips now use the relative `WrapBank` preset also used by Multiscreen. |
| `tcb-replicants-demo` | Stepped block reveal, zoom bank and layered stars are different motion/state programs. |
| `teamg1-demo` | Circular sprite formation already uses `sprites.Group`; other timed presentation cues remain local. |
| `viva_tcb` | Paired title rasters now use the shared `WrapBank` preset; harmonic logos remain a separate formation. |

## Individual demos and intro screens

| Production / screen | Shared today | Extract next, with editable parameters |
| --- | --- | --- |
| 3D DOC | Atlas, `Config.RowBands` using `composite.RowWarp`, `PerspectiveCheckerboard`, `ProjectedBallTrain`, `timeline.IntroHandoff` | Whole-scene transform remains scene composition data. |
| Bilizir | Atlas, scrolling, independent StripWarp for text/logo, SolidCube, WaterReflection, `CopperBars` | Historical scroll timing stays a recipe; extra raster palettes and masks can use the shared bank. |
| DMA 3D | Atlas, `Config.RowColumn` and mesh primitives | Multi-material morphing mesh with per-face blend, sort and winding; masked three-speed starfield. |
| DMA Is Back, intro | `Config.Feed`, configurable CRTOverlay, `timeline.IntroHandoff` | Message and scene materials stay production data. |
| DMA Is Back, main | `Config.Scanline`, ImageGrid, NestedOrbit, JellyCube, shared fade/music cue | Other whole-scene timed layers can use the general cue director. |
| Coco, intro | `Config.Feed`, configurable CRTOverlay, immediate `timeline.IntroHandoff` music cue | Scene materials stay production data. |
| Coco, main | `Config.Scanline`, SolidCubeBatch, sprites.Group grid/translation, shared font metrics and `CopperBars` | Repeating rotozoom and remaining title-layer presentation. |
| DOM intro | Atlas, synchronized FontProgram, scrolling, `motion.WrapBank` background and raster offsets | Whole-bank font switch triggered at viewport entry; animated star atlas instances. |
| Multiscreen, four embedded productions | SolidCube, sampled DNA and projected-plane engines, atlas recipes, `CopperBars` in Coco, shared Viva raster and `scrolling.Config.Pseudo3D` text banks, complete TCB `effects.MultiPlaneScene` | Make the other three production scenes reusable constructors rather than copies of their standalone controllers. |
| Multiscreen, camera tour | Layer/viewport helpers | Camera/zoom director with visibility culling, retained outputs and serialized handoffs. |
| Vectorballs | Projected shape factories, reflection | Reusable point-morph/deformation/action sequence with explicit inherited vs cleared settings and smooth handoffs. |
| Grodan | Atlas, scrolling, bounded `Background` repetition and harmonic `sprites.Group` with bounce envelope | Repeated vertical text columns. |
| MegaTwist, intro | `Config.Feed`, atlas, `timeline.HoldRamp` 90-tick splash | Authored intro image and CRT material remain scene data. |
| MegaTwist, main | `Config.Scanline`, independent `ScanlineBackground` and DisplacementPrograms, harmonic `sprites.Group` and `sprites.GlowPainter` | Transition overlay material and scene composition. |
| Nonameno, stars | `sprites.ProjectedField`, editable radial pattern and vector pixel/trail material | The star field is complete; staggered text-page choreography is tracked below. |
| Nonameno, text pages | Atlas | Staggered per-glyph enter/exit with scale/depth, easing, delays and completion barrier; baseline sine scroll. |
| Phenomena DNA intro | Atlas, DNAFrames, sliced transport | Whole screen's two-pixel insertion, loop-start and control events as a reusable configuration; separate intro/outro reveal and bounce cues. |
| TCB multiplane | Complete `effects.MultiPlaneScene`: projected scrolling, font-independent forms, snapped `composite.Bands` mountains, per-row `composite.ProfileImage` logo, `sprites.AxisFlip` emblem | Text, artwork and music remain production parameters. |
| Replicants | Atlas, `Config.RowColumn` with variable-speed controls | Stepped block reveal, quantized logo zoom bank and layered stars. |
| TeamG1, intro | `Config.Feed` with progressive right-edge entry, atlas, flat TimedCRTOverlay, `timeline.IntroHandoff` | Message and materials remain production data. |
| TeamG1, main | TexturedCube, HarmonicImage, ProfileImage, `Config.Profiled`, sprites.Group circular formation | Timed scene/audio cues and whole-scene presentation remain composition data. |
| Viva TCB | Atlas, four editable `scrolling.Config.Pseudo3D` glyph banks, staged `RotozoomBackground` and shared `motion.WrapBank` title raster | Ten-logo harmonic formation and raster title material. |

Second Reality remains outside this full-screen inventory, but its Rotozoomer
now uses the `indexed.Rotozoom256` backend. The live RGBA tile used by Viva and
the palette-indexed fixed-point sampler in Second Reality expose different
renderers so both retain their source pixels. Its 2,580-frame isolated capture
matches the previous implementation exactly.

## Cuddly presentation units

| Screen | Shared today | Extract next, with editable parameters |
| --- | --- | --- |
| Menu | TileAlphabet, `sprites.Atlas` image banks, `sprites.FrameSequence` character animation, `sprites.FormationCarousel` with seven compiled formula modes, `composite.CachedTileParallax`, `motion.CameraFollow`, scrolling | Door/input semantics and map content stay local. |
| Loader | Bitmap font recipes, scrolling, `timeline.Countdown` and `timeline.CueClock` overlapping fade/hold windows | Bind the typed volume envelope to the playback host; initial pre-render and decrement boundary remain authored. |
| Introduction | Wave/Profile/Cell strips, sparkles, compiled `motion.WaveWrite` deformation program | Serializable logo warp program and pause/audio cues; preserve row-then-column order. |
| Big Sprite | `sprites.ProjectedField` with strict-wrap vector lines, `scrolling.RingLanes` dual-font transport, `sprites.Group` Weave formation, `RasterOverlay` fill, `sprites.AxisFlip` front/back material | Authored layer placement and image order stay scene data. |
| Colorshock II | Background sampler, scrolling, editable `FormulaFormation` two-frequency orbit and `WrapBank` position-table clock | Source position samples and artwork remain screen data. |
| Ehhh | Scrolling, compiled row-profile program shared with Digi, `sprites.Train` raster bars with phase-spaced cosine `motion.Wave`, cue-paced roller `motion.WaveClock` | Lookahead/landing control events and authored roller cue mapping. |
| Mega Scroller | Tiled background, WaveStrips, scrolling and directional `motion.BounceBank` text-surface transport | Mask material with authored source-atop blending. |
| Spreadpoint | `Config.Bands`, feedback DNA, atlas | Card/audio cue sequence, 20-ball formation and raster-filled logo material. |
| Digi | Scrolling, compiled overlapping `motion.WaveWrite` row warp, `sprites.Group` Weave formation with owned phase step, rectified `motion.WaveClock` for logo and scroll | Shared bouncing logo material. |
| LED Scroller | Bounded tiled background, cached bubble matrix, scrolling, `motion.WrapBank` backdrop/gradient offsets, harmonic `sprites.Group` and rectified `motion.WaveClock` | Raster/color ramp material. |
| 3D DOC | Scrolling, paired `composite.RowWarp` source-sampling programs, `PerspectiveCheckerboard`, `ProjectedBallTrain`, first-main-tick `timeline.IntroHandoff` cue | Raster-filled inner text material remains scene composition data. |
| Fullscreen | `BackgroundLayer` velocity, `sprites.Group` Weave formation, `scrolling.Config.RingLanes` with paced vertical wrap | Raster-filled logo bar. |
| Starwars | `Config.Crawl`, RowProjection, scrolling, `sprites.ProjectedField`, `RasterOverlay` source-in fill, rectified `motion.Wave` profile | Sampled sprite train and dual-color wave strip material. |
| Knucklebuster | Scrolling and sprite images | Seeded hit trigger with hold/release envelope; optional music signal must be a distinct mode. |
| DNA | FeedbackDNA, wave strips, projected discs | Two-sided twisting ribbon with exact front/back occlusion and sampled row-source warp. |
| Megaball | Scrolling, CoupledOrbit, rectified `motion.WaveClock` text bounce | Two interleaved ball trains with index-dependent phase stepping and editable controls. |
| Reset | `Config.Slots`, scrolling, atlas, `timeline.StageSequence` with immediate events, cue windows and paired fades | Depth-ordered paired rasters and fill material. |

## Union presentation units

| Screen | Shared today | Extract next, with editable parameters |
| --- | --- | --- |
| Introduction | Image repetition, `HarmonicCellWarp` with four editable row/column wave banks, bitmap recipes | Timed music/still cues and logo placement remain screen composition. |
| Menu | Scrolling, background sampler, `sprites.Atlas` character frames, `motion.WrapBank` panorama, `motion.LinearTick` uncover wipe, `timeline.PacedIndex` palette/walk cycles, `motion.HoldBounce` logo and `motion.WalkParallax` hall/banner | Door navigation and authored layer order stay local. |
| Loader | Reveal, bitmap recipes and `timeline.CueClock` exact-duration transition | Playback cue and layer placement remain host composition data; retain column-major reverse-row glyph order. |
| Beat Dis | Background sampler, scrolling, `motion.WrapBank` wallpaper and pattern offsets, harmonic `sprites.Group` letters with global wobble | Backdrop entrance and layer order stay authored scene data. |
| Delta Force | YM register snapshots, scrolling, WaveStrips | Register-change trigger and seven-tick sprite release; source-alpha raster fill. |
| TNT Crew 3 | Mesh, geometry, bitmap recipes | Model-group draw material and camera recession handoff; authored shapes stay data. |
| Wow Scroller | Scrolling, `RasterOverlay` source-atop fill, `motion.WrapBank` paired panel offsets | Cropped oversized-image repetition with explicit wrap periods. |
| Hidden | PointHistory, sprite instances | Delayed pointer trail, palette cycle and source clip. |
| Starballs | Camera, scrolling, `sprites.ProjectedField` with two materials, live count and depth opacity | Compose the two field materials and logo mask as an editable screen layer recipe. |
| Replicants | Atlas, scrolling, `RasterOverlay` fill, `motion.CuedFormation` letter paths, `sprites.Train` rasters driven by `motion.BounceBank` | Historically reset scroll clock. |
| TNT Crew 2 | Background sampler, BitmapText.DrawWindow, `motion.WrapBank` three-layer parallax and mutable speeds | Per-key control mapping remains scene data. |
| Level 16 | Vertical scrolling, background sampler, NestedOrbit, `motion.WrapBank` water and raster offsets | Raster/water material and authored layer occlusion. |
| Multi-Plane | Complete `effects.MultiPlaneScene` in native-stage mode, with Union's source-index phases, `composite.ProfileImage` row renderer and strict `sprites.AxisFlip` cycle | Artwork, text, soundtrack and door routing remain production data. |
| Disk Copier | Bitmap recipes, `sprites.Atlas` LCD regions, `motion.GatedWrapBank` three LCD clocks, `motion.WrapBank` six-strip raster offset, `timeline.CueRanges` stages and `timeline.SteppedEnvelope` palette | Input/state program and LED layer placement remain scene composition data. |

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
