package effects

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

// JellyCubeMode selects a complete movement and deformation behavior.
type JellyCubeMode uint8

const (
	JellyNormal JellyCubeMode = iota
	JellyTumble
	JellyPulsate
	JellySwing
	JellyBounce
)

// JellyCubeStep selects a mode for Duration seconds. Repeated modes and any
// order are supported; the sequence loops, retaining orientation at boundaries.
type JellyCubeStep struct {
	Mode     JellyCubeMode
	Duration float64
}

// Keep fixed-step counters within exact float64 integer precision, well below
// the int64 conversion boundary used by absolute-time sampling.
const jellyMaxTicks int64 = 1 << 52

// JellyCubeDeformation independently scales the five deformation families.
// A zero gain disables that family; gains of one reproduce the classic effect.
type JellyCubeDeformation struct{ Wobble, Ripple, Squash, Twist, Translation float64 }

// JellyCubeConfig describes a complete animated, solid, deformable cube.
// Size is half an edge in world units. Colors are front, back, right, left,
// top, bottom. X and Y locate the projection center in destination pixels.
// CameraFOV and CameraOffset control perspective. CameraOffset must keep the
// object in front of the camera; faces crossing the near plane are skipped.
// Speed multiplies the whole animation clock; RotationSpeed is radians/second.
// Phase offsets the initial animation time, including the mode sequence.
// Delay, ZoomDuration, Transition and step durations are seconds of animation.
// Motion uses deterministic 60 Hz integration, regardless of display rate.
// Deformation and rotation are updated only by Update, never by Draw.
// Use DefaultJellyCubeConfig, then change the options required by the scene.
type JellyCubeConfig struct {
	X, Y, Size                      float64
	CameraFOV, CameraOffset         float64
	Colors                          [6]color.RGBA
	Steps                           []JellyCubeStep
	Speed, RotationSpeed, Phase     float64
	Delay, ZoomDuration, Transition float64
	Deformation                     JellyCubeDeformation
	// MaxReplaySteps bounds work in one Update, including initial phase and seeks.
	// Zero selects 100000 fixed steps (about 28 minutes of animation).
	MaxReplaySteps int64
}

// DefaultJellyCubeConfig returns an immediately visible cube with five modes.
func DefaultJellyCubeConfig() JellyCubeConfig {
	pink := color.RGBA{R: 0xe0, G: 0xa0, B: 0xc0, A: 0xff}
	purple := color.RGBA{R: 0xe0, G: 0x60, B: 0xc0, A: 0xff}
	light := color.RGBA{R: 0xe0, G: 0xe0, B: 0xe0, A: 0xff}
	return JellyCubeConfig{
		X: 320, Y: 200, Size: 80, CameraFOV: 2000, CameraOffset: 300,
		Colors: [6]color.RGBA{pink, pink, purple, purple, light, light},
		Steps:  []JellyCubeStep{{JellyNormal, 5}, {JellyTumble, 5}, {JellyPulsate, 5}, {JellySwing, 5}, {JellyBounce, 5}},
		Speed:  1, RotationSpeed: 3, Transition: .75,
		Deformation: JellyCubeDeformation{1, 1, 1, 1, 1},
	}
}

// DMAJellyCubeConfig includes the original entrance, palette and choreography.
func DMAJellyCubeConfig() JellyCubeConfig {
	c := DefaultJellyCubeConfig()
	c.Delay = 25.0 / 60
	c.ZoomDuration = 100.0 / 60
	return c
}

type jellyFace struct {
	P1, P2, P3, P4 int
	Color          color.RGBA
}
type jellyFaceDepth struct {
	face  jellyFace
	depth float64
}

// JellyCube owns its complete controller, geometry and render buffers. Multiple
// instances may be layered with kit.Group; each has independent timing and pose.
// Ordinary updates, mode changes and geometry submission allocate no Go memory.
// Seeking backwards replays the fixed-step controller from its initial state.
type JellyCube struct {
	config                                                           JellyCubeConfig
	steps                                                            []int64
	tick, delay                                                      int64
	step                                                             int
	timer                                                            int64
	pos, zoom, pulse, swingAmplitude, bounceVelocity, bouncePosition float64
	rotation                                                         geometry.Vec3
	swingX, swingZ                                                   float64
	handoff                                                          geometry.Handoff
	deformer                                                         *geometry.DeformProgram
	vertices, transformed, previous, incoming                        [8]geometry.Vec3
	faces                                                            [6]jellyFace
	depths                                                           [6]jellyFaceDepth
	drawVertices                                                     [24]ebiten.Vertex
	drawIndices                                                      [36]uint16
	texture                                                          *ebiten.Image
	smooth, closed                                                   bool
}

// Validate checks configuration without allocating GPU resources.
func (c JellyCubeConfig) Validate() error {
	finite := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
	for _, v := range []float64{c.X, c.Y, c.Size, c.CameraFOV, c.CameraOffset, c.Speed, c.RotationSpeed, c.Phase, c.Delay, c.ZoomDuration, c.Transition, c.Deformation.Wobble, c.Deformation.Ripple, c.Deformation.Squash, c.Deformation.Twist, c.Deformation.Translation} {
		if !finite(v) {
			return fmt.Errorf("effects: non-finite jelly cube configuration")
		}
	}
	if c.Size <= 0 || c.CameraFOV <= 0 || c.Speed < 0 || c.Phase < 0 || c.Delay < 0 || c.ZoomDuration < 0 || c.Transition < 0 || len(c.Steps) == 0 || c.Deformation.Squash < 0 || c.Deformation.Squash >= 1/.15 {
		return fmt.Errorf("effects: invalid jelly cube size, timing or deformation")
	}
	scale := c.Size / 80
	radiusSquared := 3 * (c.Size * 1.2) * (c.Size * 1.2)
	// Bound the intermediate radial arithmetic and the eventual float32 vertex
	// storage. These conservative bounds include rotation and every gain family.
	extent := 8*c.Size + 192*scale*math.Abs(c.Deformation.Wobble) + 30*scale*math.Abs(c.Deformation.Ripple) + 200*scale*math.Abs(c.Deformation.Translation)
	if scale <= 0 || !finite(.03/scale) || radiusSquared <= 0 || !finite(radiusSquared) || !finite(extent) || extent > math.MaxFloat32/16 || !finite(c.RotationSpeed*float64(jellyMaxTicks)) {
		return fmt.Errorf("effects: jelly cube scale or gain exceeds representable geometry")
	}
	if c.MaxReplaySteps < 0 || c.MaxReplaySteps > jellyMaxTicks || c.Phase*60 > float64(c.replayLimit()) || c.Delay*60 > float64(jellyMaxTicks) || c.ZoomDuration*60 > float64(jellyMaxTicks) {
		return fmt.Errorf("effects: jelly cube phase exceeds replay budget or timing is out of range")
	}
	for i, s := range c.Steps {
		if s.Mode > JellyBounce || !finite(s.Duration) || s.Duration < 1.0/60 || s.Duration*60 > float64(jellyMaxTicks) {
			return fmt.Errorf("effects: invalid jelly cube step %d", i)
		}
	}
	return nil
}
func (c JellyCubeConfig) replayLimit() int64 {
	if c.MaxReplaySteps == 0 {
		return 100000
	}
	return c.MaxReplaySteps
}

func NewJellyCube(c JellyCubeConfig) (*JellyCube, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	c.Steps = append([]JellyCubeStep(nil), c.Steps...)
	cube := &JellyCube{config: c, steps: make([]int64, len(c.Steps)), delay: int64(math.Round(c.Delay * 60)), smooth: c.Transition > 0}
	for i, s := range c.Steps {
		cube.steps[i] = int64(math.Round(s.Duration * 60))
	}
	size := c.Size
	cube.vertices = [8]geometry.Vec3{{X: -size, Y: -size, Z: -size}, {X: size, Y: -size, Z: -size}, {X: size, Y: size, Z: -size}, {X: -size, Y: size, Z: -size}, {X: -size, Y: -size, Z: size}, {X: size, Y: -size, Z: size}, {X: size, Y: size, Z: size}, {X: -size, Y: size, Z: size}}
	cube.faces = [6]jellyFace{{4, 5, 6, 7, c.Colors[0]}, {1, 0, 3, 2, c.Colors[1]}, {5, 1, 2, 6, c.Colors[2]}, {0, 4, 7, 3, c.Colors[3]}, {7, 6, 2, 3, c.Colors[4]}, {0, 1, 5, 4, c.Colors[5]}}
	var err error
	cube.deformer, err = newJellyDeformer(c)
	if err != nil {
		return nil, fmt.Errorf("effects: invalid jelly cube deformation: %w", err)
	}
	// Reserve handoff storage at construction, including the first transition.
	_ = cube.handoff.Begin(cube.previous[:], cube.incoming[:], -1, 1)
	cube.reset()
	cube.texture = ebiten.NewImage(1, 1)
	cube.texture.Fill(color.White)
	if err := cube.Update(kit.Frame{}); err != nil {
		cube.Close()
		return nil, err
	}
	return cube, nil
}

func (c *JellyCube) reset() {
	c.tick, c.step, c.timer = 0, 0, 0
	c.pos, c.pulse, c.swingX, c.swingZ, c.bounceVelocity, c.bouncePosition = 0, 0, 0, 0, 0, 0
	c.rotation = geometry.Vec3{}
	c.swingAmplitude = 1
	c.zoom = 1
	if c.config.ZoomDuration > 0 {
		c.zoom = 0
	}
	// Expire any old handoff without discarding its preallocated storage.
	_ = c.handoff.Begin(c.previous[:], c.incoming[:], -1, 1)
	c.sampleRaw(c.transformed[:])
	if c.config.Steps[0].Mode == JellyBounce {
		c.bounceVelocity = .08
	}
}

// Update samples absolute local time. Identical times are idempotent, and
// skipped frames and backward seeks give the same pose as sequential playback.
func (c *JellyCube) Update(f kit.Frame) error {
	if c.closed {
		return fmt.Errorf("effects: jelly cube is closed")
	}
	if math.IsNaN(f.Time) || math.IsInf(f.Time, 0) {
		return fmt.Errorf("effects: invalid jelly cube time")
	}
	seconds := max(0, f.Time)*c.config.Speed + c.config.Phase
	if math.IsNaN(seconds) || math.IsInf(seconds, 0) || seconds*60 > float64(jellyMaxTicks) {
		return fmt.Errorf("effects: invalid jelly cube time")
	}
	target := int64(math.Floor(seconds*60 + 1e-8))
	replay := target - c.tick
	if target < c.tick {
		replay = target
	}
	if replay > c.config.replayLimit() {
		return fmt.Errorf("effects: jelly cube seek requires %d steps (limit %d)", replay, c.config.replayLimit())
	}
	if target < c.tick {
		c.reset()
	}
	for c.tick < target {
		c.advance()
	}
	return nil
}

func (c *JellyCube) advance() {
	copy(c.previous[:], c.transformed[:])
	c.tick++
	c.pos += .014
	changed := false
	if c.tick > c.delay {
		c.timer++
		if c.timer >= c.steps[c.step] {
			c.timer = 0
			c.step = (c.step + 1) % len(c.steps)
			changed = true
			switch c.config.Steps[c.step].Mode {
			case JellyBounce:
				c.bounceVelocity, c.bouncePosition = .08, 0
			case JellySwing:
				c.swingX, c.swingZ = c.rotation.X, c.rotation.Z
				if !c.smooth {
					c.swingX, c.swingZ = 0, 0
				}
				c.swingAmplitude = 1
			case JellyPulsate:
				c.pulse = math.Atan2(c.rotation.Y, c.rotation.X)
			}
		}
		speed := c.config.RotationSpeed / 60
		timer := float64(c.timer)
		switch c.config.Steps[c.step].Mode {
		case JellyNormal:
			c.rotation.X += speed
			c.rotation.Y += speed
			c.rotation.Z -= speed
		case JellyTumble:
			v := math.Sin(timer * .02)
			c.rotation.X += speed * (1 + v)
			c.rotation.Y += speed * (1.5 - v*.5)
			c.rotation.Z -= speed * (.5 + v*.5)
		case JellyPulsate:
			c.pulse += .05
			v := 1 + .3*math.Sin(c.pulse)
			c.rotation.X += speed * v
			c.rotation.Y += speed * .7 * v
			c.rotation.Z -= speed * .3
		case JellySwing:
			v := math.Sin(timer*.03) * c.swingAmplitude
			c.rotation.X = c.swingX + v*.8
			c.rotation.Y += speed * .5
			c.rotation.Z = c.swingZ + v*.4
			c.swingAmplitude *= .998
		case JellyBounce:
			c.bounceVelocity -= .001
			c.bouncePosition += c.bounceVelocity
			if c.bouncePosition < -.3 {
				c.bouncePosition = -.3
				c.bounceVelocity = math.Abs(c.bounceVelocity) * .85
			}
			c.rotation.X += speed * .3
			c.rotation.Y += speed * (1 + math.Max(0, c.bouncePosition))
			c.rotation.Z += speed * .1
		}
		if c.zoom < 1 {
			c.zoom = math.Min(1, c.zoom+1/(c.config.ZoomDuration*60))
		}
	}
	c.sampleRaw(c.transformed[:])
	if changed && c.smooth {
		_ = c.handoff.Begin(c.previous[:], c.transformed[:], float64(c.tick)/60, c.config.Transition)
	}
	if c.smooth {
		c.handoff.Apply(c.transformed[:], c.transformed[:], float64(c.tick)/60)
	}
}

// Pose returns a copy of the eight current world-space vertices, before camera
// projection and entrance zoom. Reading it never changes the animation.
func (c *JellyCube) Pose() [8]geometry.Vec3 { return c.transformed }

// SetSmoothTransitions selects matched transitions or historical abrupt modes.
// Configure this before the first Update to reproduce an entire historical run.
func (c *JellyCube) SetSmoothTransitions(enabled bool) { c.smooth = enabled && c.config.Transition > 0 }

// SetPosition moves this instance without resetting its animation or handoff.
func (c *JellyCube) SetPosition(x, y float64) {
	if !math.IsNaN(x) && !math.IsInf(x, 0) && !math.IsNaN(y) && !math.IsInf(y, 0) {
		c.config.X, c.config.Y = x, y
	}
}

func (c *JellyCube) Close() error {
	if c.texture != nil {
		c.texture.Deallocate()
		c.texture = nil
	}
	c.closed = true
	return nil
}

func (c *JellyCube) sampleRaw(dst []geometry.Vec3) {
	seconds := c.pos * 3
	extraScale, extraOffsetY, twist := 1.0, 0.0, 0.0
	switch c.config.Steps[c.step].Mode {
	case JellyPulsate:
		extraScale = 1 + .2*math.Sin(c.pulse)
	case JellyBounce:
		extraOffsetY = c.bouncePosition * 50
	case JellyTumble:
		twist = math.Sin(float64(c.timer)*.01) * .5
	}
	gains := c.config.Deformation
	sizeScale := c.config.Size / 80
	squash, stretch := 1+.15*gains.Squash*math.Sin(seconds*2), 1+.15*gains.Squash*math.Cos(seconds*2)
	sin25, cos25 := math.Sincos(seconds * 2.5)
	secondary := sin25 + .5*math.Sin(seconds*5)
	deformScale := 1.0
	if c.config.Steps[c.step].Mode == JellyPulsate {
		deformScale += .3 * math.Sin(c.pulse*2)
	}
	_ = c.deformer.SetVector(0, geometry.Vec3{X: extraScale, Y: extraScale, Z: extraScale})
	_ = c.deformer.SetWobbleAmount(2, 25*deformScale*sizeScale*gains.Wobble)
	_ = c.deformer.SetVector(3, geometry.Vec3{X: squash, Y: stretch, Z: 1 / (squash*stretch*.5 + .5)})
	_ = c.deformer.SetTwistAmount(5, twist*gains.Twist)
	_ = c.deformer.SetVector(6, c.rotation)
	_ = c.deformer.SetVector(7, geometry.Vec3{X: secondary * 8 * sizeScale * gains.Translation, Y: (cos25*8 + extraOffsetY) * sizeScale * gains.Translation, Z: math.Sin(seconds*3.7) * 4 * sizeScale * gains.Translation})
	c.deformer.Apply(dst, c.vertices[:], seconds)
}

func newJellyDeformer(c JellyCubeConfig) (*geometry.DeformProgram, error) {
	scale := c.Size / 80
	p, err := geometry.NewDeformProgram([]geometry.DeformStage{
		{Kind: geometry.DeformScale, Vector: geometry.Vec3{X: 1, Y: 1, Z: 1}},
		{Kind: geometry.DeformReference},
		{Kind: geometry.DeformWobble, Wobble: geometry.WobbleField{
			Key: geometry.Vec3{X: .01 / scale, Y: .02 / scale, Z: .03 / scale}, Amount: 25 * scale, Radius: c.Size, BaseInfluence: .5, RadialInfluence: .5,
			X: []geometry.Harmonic{{Gain: .4, Speed: 1, Spatial: 5}, {Gain: .2, Speed: 2.1, Spatial: 3}},
			Y: []geometry.Harmonic{{Gain: .4, Speed: 1.3, Spatial: 7, Cosine: true}, {Gain: .2, Speed: 1.7, Spatial: 4, Cosine: true}},
			Z: []geometry.Harmonic{{Gain: .3, Speed: .7, Spatial: 3}, {Gain: .15, Speed: 1.9, Spatial: 6, Cosine: true}},
		}},
		{Kind: geometry.DeformScale, Vector: geometry.Vec3{X: 1, Y: 1, Z: 1}},
		{Kind: geometry.DeformRipple, Ripple: geometry.RippleField{Amplitude: 5 * scale * c.Deformation.Ripple, Speed: 4, Spatial: 10, Radius: c.Size, CoordinateScale: c.Size, DirectionX: geometry.Vec3{Y: 1}, DirectionY: geometry.Vec3{X: 1}}},
		{Kind: geometry.DeformTwist, Twist: geometry.TwistField{Axis: 1, Radius: c.Size}},
		{Kind: geometry.DeformRotate},
		{Kind: geometry.DeformTranslate},
	})
	return p, err
}
