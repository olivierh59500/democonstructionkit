package presets

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

// CuddlyKnucklebusterHits groups four borrowed overlays into three separately
// triggered channels with the authored seeded-random cadence.
func CuddlyKnucklebusterHits(head, bass, left, right *ebiten.Image,
	randomFloat func() float64) sprites.LatchedOverlayConfig {
	return sprites.LatchedOverlayConfig{
		Motion: motion.LatchedTriggersConfig{Count: 3, Period: 5, ReleaseAt: 1,
			Source: motion.TriggerRandom, RandomFloat: randomFloat,
			RandomRange: 750, RandomBias: 1, Threshold: 725},
		Sprites: []sprites.LatchedSprite{
			{Image: head, Channel: 0, X: 316, Y: 110},
			{Image: bass, Channel: 0, X: 298, Y: 210},
			{Image: left, Channel: 1, X: 196, Y: 138},
			{Image: right, Channel: 2, X: 381, Y: 131},
		},
		Filter: ebiten.FilterNearest, Blend: ebiten.BlendSourceOver,
	}
}

// CuddlyKnucklebusterSignalHits selects live music/event inputs on the same
// channels. A true signal reasserts its overlay until the configured release.
func CuddlyKnucklebusterSignalHits(head, bass, left, right *ebiten.Image,
	signal func(index int, tick int) bool) sprites.LatchedOverlayConfig {
	c := CuddlyKnucklebusterHits(head, bass, left, right, nil)
	c.Motion.Source = motion.TriggerSignal
	c.Motion.RandomFloat = nil
	c.Motion.Signal = signal
	c.Motion.SignalEveryTick = true
	c.Motion.Threshold = 1
	return c
}
