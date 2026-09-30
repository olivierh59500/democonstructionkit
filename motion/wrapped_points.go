package motion

import (
	"fmt"
	"image"
)

// WrappedPoint keeps authored word-sized coordinates separate from the camera.
type WrappedPoint struct{ X, Y, Z int16 }

// WrappedPointProjectionConfig makes fixed-point projection arithmetic explicit.
// Each coordinate is added to its offset as a wrapping 16-bit word, then masked.
// Bias is added to signed X/Y words before multiplication. Numerator divided by
// masked Z plus DepthBias supplies the reciprocal. Products wrap at 32 bits,
// shift arithmetically, and wrap at 16 bits before Center is added. An empty
// Bounds disables clipping. A zero denominator or negative depth is invisible.
type WrappedPointProjectionConfig struct {
	Mask                 [3]uint16
	Bias                 [2]int16
	Numerator, DepthBias int32
	Shift                uint8
	Center               image.Point
	Bounds               image.Rectangle
}

// WrappedPointPose includes masked depth for independent palette or sprite rules.
type WrappedPointPose struct {
	X, Y  int
	Depth uint16
}

// WrappedPointProjection owns immutable parameters and samples without allocation.
// It can drive an indexed pixel plane, a sprite field or a caller's own renderer.
type WrappedPointProjection struct{ config WrappedPointProjectionConfig }

func NewWrappedPointProjection(c WrappedPointProjectionConfig) (*WrappedPointProjection, error) {
	if c.Numerator <= 0 || c.Shift > 31 || c.Center.X < -(1<<30) || c.Center.X > 1<<30 || c.Center.Y < -(1<<30) || c.Center.Y > 1<<30 {
		return nil, fmt.Errorf("motion: invalid wrapped point reciprocal, shift or center")
	}
	if !c.Bounds.Empty() {
		dx, dy := int64(c.Bounds.Max.X)-int64(c.Bounds.Min.X), int64(c.Bounds.Max.Y)-int64(c.Bounds.Min.Y)
		if dx > 8192 || dy > 8192 || c.Bounds.Min.X < -(1<<30) || c.Bounds.Min.Y < -(1<<30) || c.Bounds.Max.X > 1<<30 || c.Bounds.Max.Y > 1<<30 {
			return nil, fmt.Errorf("motion: wrapped point clip exceeds projection budget")
		}
	}
	return &WrappedPointProjection{config: c}, nil
}

func (p *WrappedPointProjection) Project(point WrappedPoint, offset [3]int16) (WrappedPointPose, bool) {
	if p == nil {
		return WrappedPointPose{}, false
	}
	c := p.config
	x := int32(int16((uint16(point.X)+uint16(offset[0]))&c.Mask[0]) + c.Bias[0])
	y := int32(int16((uint16(point.Y)+uint16(offset[1]))&c.Mask[1]) + c.Bias[1])
	z := (uint16(point.Z) + uint16(offset[2])) & c.Mask[2]
	denominator := int64(z) + int64(c.DepthBias)
	if denominator <= 0 {
		return WrappedPointPose{Depth: z}, false
	}
	factor := int32(int64(c.Numerator) / denominator)
	// Multiplication deliberately wraps before the shift, as in a signed
	// long-word multiply whose low result is used by the word projection.
	pose := WrappedPointPose{X: int(int16((x*factor)>>c.Shift)) + c.Center.X,
		Y: int(int16((y*factor)>>c.Shift)) + c.Center.Y, Depth: z}
	if !c.Bounds.Empty() && !image.Pt(pose.X, pose.Y).In(c.Bounds) {
		return pose, false
	}
	return pose, true
}
