# Native effect consolidation acceptance

Reviewed on 2026-09-28. The effect inventory covers eighteen native production
repositories, including every Cuddly and Union presentation unit. FR-010 and
Second Reality retain their existing shared kernels and are included in the
twenty-module compatibility check; their complete part renderers are outside
this inventory's scope.

The reusable animated-effect engines found in this inventory have been
extracted into DCK. The production adapters provide assets, text, parameters,
menu/input semantics and ordered scene composition. The [effect map](EFFECT_MAP.md)
identifies the components used by each screen. The remaining host functions do
not constitute another scrolling, deformation, sprite-trajectory, mesh or shader
engine. This conclusion is limited to the audited native sources and revision
set in the [acceptance record](fidelity/catalog-acceptance-20260928.json).

| Requirement | Result and evidence |
| --- | --- |
| Inventory every native DCK screen | The effect map covers the eighteen in-scope repositories, both collections' introductions, menus and loaders, and each selectable screen. A refreshed syntax-tree pass reviews 81 non-test files and 80 loops, including controls, asset preparation and the collection hosts. |
| One scrolling entry point with independent fonts and modes | `scrolling.New` owns ordinary glyph transport or nineteen specialized configurations. Fonts carry metrics, character order, dimensions and materials. Horizontal, vertical, page, path, zoom, perspective, recycled, cued-strip and feedback forms use the same facade. The native font check renders 336 mixed-size font/mode combinations with visible, repeatable pixels. |
| Reusable sprites, fields and 3D objects | `sprites.Group`, projected/animated fields, histories and formations own their motion and rendering. Point-cloud and solid/textured mesh components accept independent instances, palettes, trajectories and materials. Cube, pyramid, plane, flag and sphere variations have geometry tests, native comparisons and Pixel samples recorded in the map. |
| Reusable backgrounds, logos and image effects | Repeated/tiled backgrounds, rotozoom, scanline displacement, row/column/cell warps, raster fills, plasma, water reflection, magnifier and CRT components are used by the catalog. They accept different images and expose their dimensions, phase, sampling and composition settings. |
| Timed and cumulative composition | Timelines and owned Update clocks preserve authored handoffs. Ordered image passes operate on a layer or an entire scene. Saved examples combine CRT, reflection and a periodic magnifier; the normal desktop and Android hosts load the same project. Repeated Draw checks preserve cursors, history and sampled phases. |
| Shared music entry point | In-scope DCK production sources import the DCK audio facade rather than either playback backend. `sound.Open` selects YM, module or recorded audio. The existing PCM compatibility checks are part of the module test pass. |
| Preserve production behavior and originals | The source comparison ledger covers seventeen paired productions, with explicit historical switches for intentional variations. Both collections have 607 complete-frame facade comparisons; the later DNA migration adds 42 complete-frame and 22 isolated comparisons. The refreshed retention check finds no missing or changed tracked original Go file in the pinned source snapshots. |
| Verify runtime and resource behavior | DCK and all twenty demo modules pass tests and vet against the current DCK checkout. The map records stable-scene Pixel traces for the catalog, optional object previews and one-minute-per-screen collection tours. Latest DNA scenes contribute 2,976 distinct presentation intervals with none above 20 ms. Bounded history surfaces and zero-allocation transport tests cover the new DNA composition. |
| Document and version the work | README provides parameter and composition examples; the effect map and JSON records identify tested revisions. The separate portfolio DCK page supplies bilingual recipes and videos. Changes are committed in English; hosting exchange directories are excluded through local Git rules. |

## Fidelity and measurement limits

Full-canvas comparisons use deterministic native before/after captures without
pixel masks or error tolerance. They establish the sampled checkpoints, not
every intermediate frame or pixel equivalence with Atari hardware. The map
keeps DMA 3D, MegaTwist and Multiscreen's small sampling differences visible,
TeamG1's intentional unclipped intro, and Union Replicants' authored trajectory
approximation. Their source implementations remain available.

Stable-scene Pixel samples stay near a 16.7 ms presentation interval. The longer
collection tours expose occasional loading/menu handoff stalls. Memory values
are snapshots or sampled maxima; actual peaks, battery consumption and every
future combination have not been measured. New authored compositions can use
the existing Pixel probe and native capture tools for their own acceptance.

The planned [2.0.0 graphical editor](ROADMAP.md) remains a separate application
task. Version 1.0.0 is the Developer Edition. DCK already provides inspectable
controllers and a saved-project schema for common scenes; custom Go callbacks
remain available for additional authored behavior.
