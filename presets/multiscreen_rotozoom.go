package presets

// MultiscreenCocoRotozoom preserves the embedded panel's viewport-sampled
// texture phase and half-bright material with the shared harmonic pose.
func MultiscreenCocoRotozoom(width, height float64) VivaRotozoomConfig {
	c := CocoRotozoom(width, height)
	c.TilePhaseX, c.TilePhaseY = width*4, height*4
	return c
}

// MultiscreenVivaRotozoom uses the same running harmonic clocks with the
// embedded Viva panel's larger source phase and untinted tile.
func MultiscreenVivaRotozoom(width, height float64) VivaRotozoomConfig {
	c := CocoRotozoom(width, height)
	c.TilePhaseX, c.TilePhaseY = width*8, height*8
	c.Color = [4]float32{1, 1, 1, 1}
	return c
}
