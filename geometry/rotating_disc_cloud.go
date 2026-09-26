package geometry

import (
	"fmt"
	"math"
	"slices"
)

// RotatingDiscCloudConfig projects fixed 3D points while rotating their X/Z
// plane around Y. Center, depth base, Y offset and radius scale are independent
// of point count or the drawing material.
type RotatingDiscCloudConfig struct {
	Points                        []Vec3
	CenterX, CenterY              float64
	DepthBase, YOffset, Focal     float64
	Angle, AngleStep, RadiusScale float64
}

// ProjectedDiscPose carries the depth used for stable painter ordering.
type ProjectedDiscPose struct{ X, Y, Z, Radius float64 }

// RotatingDiscCloud owns a copy of the source points and two reusable pose
// banks. Step sorts in ascending depth and never changes the authored points.
type RotatingDiscCloud struct {
	config  RotatingDiscCloudConfig
	poses   []ProjectedDiscPose
	scratch []ProjectedDiscPose
	angle   float64
	ready   bool
}

func NewRotatingDiscCloud(c RotatingDiscCloudConfig) (*RotatingDiscCloud, error) {
	if len(c.Points) < 1 || len(c.Points) > 1_000_000 || c.Focal <= 0 || c.RadiusScale <= 0 {
		return nil, fmt.Errorf("geometry: invalid rotating disc cloud count or projection")
	}
	for _, value := range [...]float64{c.CenterX, c.CenterY, c.DepthBase, c.YOffset,
		c.Focal, c.Angle, c.AngleStep, c.RadiusScale} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("geometry: nonfinite disc cloud setting")
		}
	}
	for _, point := range c.Points {
		if math.IsNaN(point.X) || math.IsInf(point.X, 0) || math.IsNaN(point.Y) || math.IsInf(point.Y, 0) ||
			math.IsNaN(point.Z) || math.IsInf(point.Z, 0) {
			return nil, fmt.Errorf("geometry: nonfinite disc cloud source point")
		}
	}
	c.Points = append([]Vec3(nil), c.Points...)
	count := len(c.Points)
	return &RotatingDiscCloud{config: c, poses: make([]ProjectedDiscPose, count),
		scratch: make([]ProjectedDiscPose, count), angle: c.Angle}, nil
}

// Step applies one angular increment before projecting, then sorts the frame.
// Invalid camera crossings leave the last valid pose and angle unchanged.
func (c *RotatingDiscCloud) Step() error {
	if c == nil {
		return fmt.Errorf("geometry: nil rotating disc cloud")
	}
	next := c.angle + c.config.AngleStep
	if math.IsNaN(next) || math.IsInf(next, 0) {
		return fmt.Errorf("geometry: disc cloud angle overflow")
	}
	sin, cos := math.Sincos(next)
	for i, p := range c.config.Points {
		x, z := p.X*cos+p.Z*sin, p.Z*cos-p.X*sin
		denominator := c.config.DepthBase - z
		if denominator == 0 {
			return fmt.Errorf("geometry: disc cloud crossed projection plane")
		}
		scale := c.config.Focal / denominator
		pose := ProjectedDiscPose{X: c.config.CenterX + x*scale,
			Y: c.config.CenterY - (p.Y+c.config.YOffset)*scale,
			Z: z, Radius: scale * c.config.RadiusScale}
		for _, value := range [...]float64{pose.X, pose.Y, pose.Z, pose.Radius} {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return fmt.Errorf("geometry: nonfinite disc cloud projection")
			}
		}
		c.scratch[i] = pose
	}
	slices.SortStableFunc(c.scratch, func(a, b ProjectedDiscPose) int {
		if a.Z < b.Z {
			return -1
		}
		if a.Z > b.Z {
			return 1
		}
		return 0
	})
	c.poses, c.scratch = c.scratch, c.poses
	c.angle = next
	c.ready = true
	return nil
}

func (c *RotatingDiscCloud) Poses() []ProjectedDiscPose { return c.poses }
func (c *RotatingDiscCloud) Angle() float64             { return c.angle }
func (c *RotatingDiscCloud) Ready() bool                { return c.ready }

// SetAngleStep changes the live rotation rate without resetting the pose.
func (c *RotatingDiscCloud) SetAngleStep(step float64) error {
	if c == nil || math.IsNaN(step) || math.IsInf(step, 0) {
		return fmt.Errorf("geometry: invalid disc cloud angular step")
	}
	c.config.AngleStep = step
	return nil
}

// Reset restores the first rotation cue while retaining point and pose memory.
func (c *RotatingDiscCloud) Reset() {
	if c == nil {
		return
	}
	c.angle = c.config.Angle
	c.ready = false
	clear(c.poses)
	clear(c.scratch)
}
