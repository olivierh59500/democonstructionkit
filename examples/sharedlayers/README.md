# Shared indexed artwork, points and meters

Run `go run ./examples/sharedlayers` from the module root. Two retained slots
show one indexed image with independent palettes. Two point planes overlay the
artwork using the same reciprocal projector, with OR and XOR collision policies.
Some point pairs deliberately coincide: OR retains them, while XOR cancels them.
Gradient bars compose a third layer without another image surface.

The meter consumes synthetic YM register snapshots so the example needs no audio
asset. In a production, call `registers, available := stream.YMRegisters()` and
pass available snapshots to `YMPeriodMeter.Step`. Level values can drive sprite
size, light intensity or deformation instead of bars. Each configuration sets
its own volume curve, column mapping, gating and decay.

Artwork is generated once. Image slots and palette banks remain separate;
changing palettes does not upload artwork. Point sampling clears only touched
pixels and deduplicates collisions independently of the resulting mask. Drawing
reuses each prepared state without advancing its clock. The three layers close
their owned resources, while the example releases its borrowed image afterward.

Use `-capture /tmp/dck-sharedlayers -frame 200` for one native PNG, or add
`-frames 300` to export six seconds at the authored 50 Hz cadence. Capture reads
pixels for export; interactive effects do not use GPU pixel readback.
