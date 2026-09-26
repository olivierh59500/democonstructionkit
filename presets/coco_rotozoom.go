package presets

// CocoRotozoom reuses Viva's harmonic rotozoom without its staged entrance.
// Coco draws an oversized source quad with a half-bright texture tint; its
// source phase is centered at half the quad dimensions.
func CocoRotozoom(width, height float64) VivaRotozoomConfig {
	c := DefaultVivaRotozoomConfig(width, height)
	c.StartStage = 3
	c.TilePhaseX, c.TilePhaseY = width*2, height*2
	c.Color = [4]float32{.5, .5, .5, 1}
	return c
}
