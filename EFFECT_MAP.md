# Effect map for the native demo catalog

This map covers every DCK production under `demos/` except FR-010 and Second
Reality. It was checked against the screen implementations and the detailed
workspace audits on 2026-09-23. A complete component owns its transport,
animation state, geometry and rendering resources. The production supplies
assets, messages, presets, input and scene order. A shared draw helper alone is
not counted as a complete effect.

`Shared` names a complete component already used by that screen. `Extract next`
identifies remaining reusable logic and the parameters that must stay editable.
The entries are acceptance work, not claims that all screens are already small.

## Individual demos and intro screens

| Production / screen | Shared today | Extract next, with editable parameters |
| --- | --- | --- |
| 3D DOC | Atlas, ordinary scrolling, strip and quad helpers | Two-pass row-warped scroll; perspective XOR checkerboard; phased projected balls/shadows with movement sequences, depth order and palette. |
| Bilizir | Atlas, scrolling, independent StripWarp for text/logo, SolidCube, WaterReflection | Copper/raster bank: bars, palette, speed, spacing, order and masks. Historical scroll timing stays a recipe. |
| DMA 3D | Atlas, scrolling and mesh primitives | Multi-material morphing mesh with per-face blend, sort and winding; masked three-speed starfield; row-then-column scroll sampling. |
| DMA Is Back, intro | `Config.Feed`, configurable CRTOverlay | Music/fade cues can become a typed timeline; message and cue times stay production data. |
| DMA Is Back, main | `Config.Scanline`, ImageGrid, NestedOrbit, JellyCube | Soundtrack start and whole-scene fade can use the same cue/envelope system as other intros. |
| Coco, intro | `Config.Feed`, configurable CRTOverlay | Intro-to-main music cue. |
| Coco, main | `Config.Scanline`, SolidCubeBatch, shared font metrics | Repeating rotozoom; 16-logo formation with phase/spacing; copper-filled title material. |
| DOM intro | Atlas, synchronized FontProgram, scrolling | Whole-bank font switch triggered at viewport entry; timed raster backgrounds and animated star atlas instances. |
| Multiscreen, four embedded productions | SolidCube, sampled DNA and projected-plane engines, atlas recipes | Make the four production scenes reusable constructors rather than copies of their standalone controllers. |
| Multiscreen, camera tour | Layer/viewport helpers | Camera/zoom director with visibility culling, retained outputs and serialized handoffs. |
| Vectorballs | Projected shape factories, reflection | Reusable point-morph/deformation/action sequence with explicit inherited vs cleared settings and smooth handoffs. |
| Grodan | Atlas and scrolling | Bounded layered image repetition; phased sprite chain with count, image selection, spacing and wave envelope; repeated vertical text columns. |
| MegaTwist, intro | `Config.Feed`, atlas | Keep the screenshot-matched splash/transition timing as an editable cue program. |
| MegaTwist, main | `Config.Scanline` with a one-time/looping DisplacementProgram and strict X rejection | Background scanline layer with independent clock and source wrap; glowing sprite formation and transition overlay. |
| Nonameno, stars | Projected-field primitives | Seeded radial spawn/respawn, depth-based pixel size/color and streak history with reset policy. |
| Nonameno, text pages | Atlas | Staggered per-glyph enter/exit with scale/depth, easing, delays and completion barrier; baseline sine scroll. |
| Phenomena DNA intro | Atlas, DNAFrames, sliced transport | Whole screen's two-pixel insertion, loop-start and control events as a reusable configuration; separate intro/outro reveal and bounce cues. |
| TCB multiplane | Projected scrolling, font-independent forms, background bands | Logo row warp and central flip; share screen recipe with Multiscreen and Union while preserving phase-per-visible-slot timing. |
| Replicants | Atlas, scrolling | Stepped block reveal, quantized logo zoom bank, layered stars and two-pass row/column scroller. |
| TeamG1, intro | `Config.Feed`, atlas | Its time-dependent CRT material should accept a shader profile while preserving the authored time clock. |
| TeamG1, main | TexturedCube, harmonic plasma kernel, atlas | GPU plasma surface ownership; spiral logo formation; live row-warped logo and scroll, with sample/filter rules. |
| Viva TCB | Atlas, scrolling, image repetition | Four pseudo-3D glyph banks with per-glyph scale/order/snap; ten-logo harmonic formation; raster title material. |

## Cuddly presentation units

| Screen | Shared today | Extract next, with editable parameters |
| --- | --- | --- |
| Menu | TileAlphabet, atlas, background sampler, scrolling | Sprite ensemble's seven motion programs, camera parallax and atlas frame player. Door/input behavior stays in the menu. |
| Loader | Bitmap font recipes and scrolling | Typed countdown, gain/fade and hold cues; preserve the initial pre-render and decrement boundary. |
| Introduction | Wave/Profile/Cell strips, sparkles | Serializable logo warp program and pause/audio cues; preserve row-then-column order. |
| Big Sprite | Streaks, scrolling, Weave formation | Shared front/back flip material, synchronized dual-font transport and raster mask fill. |
| Colorshock II | Background sampler, scrolling | Position-table placement and two-frequency background orbit as an editable motion recipe. |
| Ehhh | Scrolling, row profile | Lookahead/landing control events, raster train and roller bounce states. |
| Mega Scroller | Tiled background, WaveStrips, scrolling | Ping-pong text transport and mask material with authored source-atop blending. |
| Spreadpoint | `Config.Bands`, feedback DNA, atlas | Card/audio cue sequence, 20-ball formation and raster-filled logo material. |
| Digi | Scrolling, lookup row warp, Weave formation | Shared bouncing logo and nested-sine letter formation preset with independent phases. |
| LED Scroller | Bounded tiled background, cached bubble matrix, scrolling | Raster/color ramp material and amplitude-modulated letter ensemble. |
| 3D DOC | Scrolling, batch primitives | Perspective XOR ground, shadowed projected-ball train, seven continuous motion handoffs. |
| Fullscreen | Background sampler, Weave formation, scrolling | Recycled multi-scroll lanes and raster-filled logo bar. |
| Starwars | `Config.Crawl`, RowProjection, scrolling, projected starfield | Sampled sprite train and dual-color wave strip material. |
| Knucklebuster | Scrolling and sprite images | Seeded hit trigger with hold/release envelope; optional music signal must be a distinct mode. |
| DNA | FeedbackDNA, wave strips, projected discs | Two-sided twisting ribbon with exact front/back occlusion and sampled row-source warp. |
| Megaball | Scrolling, CoupledOrbit | Two interleaved ball trains with index-dependent phase stepping and editable controls. |
| Reset | `Config.Slots`, scrolling, atlas | Multi-stage cue director, depth-ordered paired rasters and fill material. |

## Union presentation units

| Screen | Shared today | Extract next, with editable parameters |
| --- | --- | --- |
| Introduction | Image repetition, CellWarp, bitmap recipes | Reusable harmonic cell-offset program and timed music/still cues. |
| Menu | Scrolling, background sampler, sprite instances | Walkable panorama/parallax, palette cycling and uncover wipe; door navigation stays local. |
| Loader | Reveal and bitmap recipes | Audio-timed cue wrapper; retain column-major reverse-row glyph order. |
| Beat Dis | Background sampler, scrolling | Phase-spaced letter ensemble with global wobble. |
| Delta Force | YM register snapshots, scrolling, WaveStrips | Register-change trigger and seven-tick sprite release; source-alpha raster fill. |
| TNT Crew 3 | Mesh, geometry, bitmap recipes | Model-group draw material and camera recession handoff; authored shapes stay data. |
| Wow Scroller | Scrolling | Cropped oversized-image repetition with explicit wrap periods and raster fill. |
| Hidden | PointHistory, sprite instances | Delayed pointer trail, palette cycle and source clip. |
| Starballs | Camera, scrolling | Seeded depth field with interchangeable sprite material, quantized size/opacity and mask passes. |
| Replicants | Atlas, scrolling | Configurable bouncing raster train and historically reset scroll clock. |
| TNT Crew 2 | Background sampler, BitmapText.DrawWindow | Layered parallax phase/velocity program with user direction and exact reset policy. |
| Level 16 | Vertical scrolling, background sampler, NestedOrbit | Raster/water material and authored layer occlusion. |
| Multi-Plane | Bands, projected scrolling | Logo row lookup/center flip; package the form program and strip background as one editable screen recipe. |
| Disk Copier | Bitmap recipes and sprite regions | Stepped color envelope, LED/LCD atlas player and state/cue program. |

## Extraction order and acceptance

1. Finish the repeated scanline family: DMA, Coco and MegaTwist now share the
   complete proportional transport. 3D DOC, DMA 3D and Replicants need
   additional source-row and column-pass policies. A new policy must
   be named for its sampling rule and validated against the original pixels.
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
