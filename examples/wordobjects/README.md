# Word mesh objects

Run `go run ./examples/wordobjects` from the module root. Three independent
objects share one copied model recipe, one compiled sine/matrix bank and one
borrowed three-column row palette. Their angle rates and depth waves differ.

`WordEulerMatrix` configures the angular period, quantization, quarter phase,
product shifts and nested-product association. `WordProjectionConfig` chooses
depth/translation scaling, center, zero-depth handling and division overflow.
The classic keep-low-word policy reproduces signed word division overflow;
clamping and rejection support other compositions.

`WordMesh` owns projection, projected-word winding, clipped face caches,
contours and material submission. `SetPose` updates an authored object manually;
`Animate` can select a pose from a frame. Equivalent poses retain the projection
cache. A retained instance window can draw several copies from one cached pose,
with different placement. Materials choose texture, UVs, tint and fill rules.

Only immutable artwork/model preparation allocates geometry. Drawing uses
bounded buffers and does not advance a clock or read pixels from the GPU. No
offscreen framebuffer or color-conversion pass is needed by this component.

Use `-capture /tmp/dck-wordobjects -frame 200` to save a native PNG, or add
`-frames 300` for six seconds of consecutive 50 Hz frames. Export reads pixels;
interactive drawing does not. Source images remain alive after mesh Close and
are released separately by their owner.
