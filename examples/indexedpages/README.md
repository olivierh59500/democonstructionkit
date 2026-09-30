# Indexed palette pages

This example uploads one procedural image containing 16 indices and shows it in
two independent effects. The left effect crops and places the artwork with an
RGB12 palette; the right effect keeps the complete image and uses RGB565. Both
palettes fade from black, then alternate between warm and cool colors. The
artwork stays unchanged throughout.

Run from the module root:

```sh
go run ./examples/indexedpages
go run ./examples/indexedpages -capture /tmp/dck-indexedpages -frame 350
```

The source stores `index * 17` in the red component. Set
`Channel: composite.BitplaneRed, Scale: 15` to decode it. This encoding is
independent of the packed palette format: the default `palette.PackedRGB` is
RGB12, while RGB565 uses five, six and five bits at shifts eleven, five and zero.
Images with another index encoding can choose another channel, scale and offset.

`effects.IndexedImage.Fade(from, to, start, ticks)` copies its endpoints and
uses the absolute time passed to `Update`. `FPS: 50` means that a frame with
`Time: 7` samples source tick 350, regardless of draw frequency. A 101-sample
fade includes both endpoints at ticks zero and one hundred, spanning two
seconds. Each packed channel follows signed delta interpolation, preserving
the integer steps of classic palette transitions in either direction.

`Crop` selects a rectangle in the converted image, starting at `(0, 0)`;
`Options.GeoM` then scales, rotates or places that result. A page implements
`kit.Effect`: add it to `kit.Group` or `kit.Layers` alongside scrolling text,
sprites and logos in the desired drawing order. To deform the whole result,
draw the group to a reusable render surface and use that image as the input to
a warp, reflection or other composition pass. Palette animation changes colors;
those separate effects change placement, geometry or the final image.

Each page owns one conversion surface, one shader and a small color bank. The
two pages borrow the same uploaded source image. `Close` releases each page's
resources; the example releases the source afterward. There are no per-frame
artwork uploads, CPU pixel reads or image allocations.
