# Demo Construction Kit roadmap

## 1.0.0 — Developer Edition

The first release is a library for Go developers building demos and intros with
Ebitengine. Effects are configured through Go APIs; executable examples and
JSON projects show how to combine them. A developer supplies artwork, fonts,
messages and music, then chooses the composition and its timing.

The existing runtime, asset descriptions, effect configurations, timelines and
saved-project schema provide the foundation for the graphical application.

## Planned 2.0.0 — Visual Editor

The next major version is intended for everyone, including people who do not
know how to program. The complete creation workflow will be graphical, with
no mandatory Go code or manual JSON editing:

1. Create a project and import images, sprite sheets, bitmap fonts and music.
2. Choose effects from a library of configurable presets and add instances to
   the scene.
3. Adjust parameters visually, including font dimensions and character order,
   speeds, phases, colors, deformation, cameras and music-driven properties.
4. Draw or edit paths and curves, then arrange and superpose layers in the
   desired order.
5. Organize screens on a timeline, schedule transitions and synchronize effects
   with the soundtrack.
6. Preview the result, save the project and export a runnable demo or intro.

The editor will use the shared Go/Ebitengine effects so its preview and
exported production follow the same rendering and animation rules.
The programmable runtime remains available for developers extending the kit.

This is the direction planned for 2.0.0. The graphical editor is a future
release; no release date is announced.
