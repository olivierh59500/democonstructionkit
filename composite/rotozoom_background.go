package composite

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

// RotozoomProgram supplies one complete repeated-texture transform per logical
// frame. A preset can own a multi-stage entrance while custom Go programs can
// synchronize the transform with music or another effect.
type RotozoomProgram interface {
	Update(kit.Frame) error
	Repetition() Repetition
}

// RotozoomVelocity adds simple independent linear motion to the base pose.
// Center and phase use pixels per second, rotation uses radians per second.
type RotozoomVelocity struct {
	CenterX, CenterY, Zoom, Rotation, PhaseX, PhaseY float64
}

// RotozoomBackgroundConfig borrows one source tile and chooses either a
// program or a base pose with velocity. Repetition handles texture wrapping,
// zoom, rotation and sampling in one GPU quad without a tiled render target.
type RotozoomBackgroundConfig struct {
	Image    *ebiten.Image
	Pose     Repetition
	Velocity RotozoomVelocity
	Program  RotozoomProgram
}

// RotozoomBackground is a complete kit.Effect with a fixed update/draw order.
// Several instances may share one tile while using different phases or speeds.
type RotozoomBackground struct {
	config RotozoomBackgroundConfig
	pose   Repetition
}

func NewRotozoomBackground(c RotozoomBackgroundConfig) (*RotozoomBackground, error) {
	if c.Image == nil {
		return nil, fmt.Errorf("composite: rotozoom source image is nil")
	}
	if c.Program == nil && c.Pose.Zoom == 0 {
		c.Pose.Zoom = 1
	}
	if !finiteRotoVelocity(c.Velocity) || c.Program == nil && !finiteRepetition(c.Pose) {
		return nil, fmt.Errorf("composite: invalid rotozoom pose or velocity")
	}
	return &RotozoomBackground{config: c, pose: c.Pose}, nil
}

func (r *RotozoomBackground) Update(f kit.Frame) error {
	if r == nil || math.IsNaN(f.Time) || math.IsInf(f.Time, 0) {
		return fmt.Errorf("composite: invalid rotozoom frame")
	}
	pose := r.config.Pose
	if r.config.Program != nil {
		if err := r.config.Program.Update(f); err != nil {
			return err
		}
		pose = r.config.Program.Repetition()
	}
	v := r.config.Velocity
	pose.CenterX += v.CenterX * f.Time
	pose.CenterY += v.CenterY * f.Time
	pose.Zoom += v.Zoom * f.Time
	pose.Rotation += v.Rotation * f.Time
	pose.PhaseX += v.PhaseX * f.Time
	pose.PhaseY += v.PhaseY * f.Time
	if !finiteRepetition(pose) {
		return fmt.Errorf("composite: rotozoom transform became invalid")
	}
	r.pose = pose
	return nil
}

func (r *RotozoomBackground) Draw(dst *ebiten.Image) {
	if r == nil || dst == nil {
		return
	}
	Repeat(dst, r.config.Image, r.pose)
}

// Repetition returns the currently sampled transform for another layer or a
// deterministic capture. It does not advance the animation.
func (r *RotozoomBackground) Repetition() Repetition { return r.pose }

func finiteRepetition(p Repetition) bool {
	for _, value := range [...]float64{p.CenterX, p.CenterY, p.Zoom, p.Rotation, p.PhaseX, p.PhaseY} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return p.Zoom > 0
}

func finiteRotoVelocity(v RotozoomVelocity) bool {
	for _, value := range [...]float64{v.CenterX, v.CenterY, v.Zoom, v.Rotation, v.PhaseX, v.PhaseY} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}
