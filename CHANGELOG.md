# Changelog

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
