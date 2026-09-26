package motion

import (
	"fmt"
	"math"
)

// HarmonicTransformConfig describes a scalable image pose from one phase.
// Each coordinate may mix sine, cosine and its corresponding output scale;
// independent X/Y scale waves allow squash or stretch. Phase offsets the
// whole pose for another logo or sprite using the same clock.
type HarmonicTransformConfig struct {
	Base, Sin, Cos, ScaleCoupling Point
	ScaleBase, ScaleSin, ScaleCos Point
	Phase                         float64
}

type HarmonicTransformPose struct {
	X, Y, ScaleX, ScaleY float64
}

type HarmonicTransform struct{ config HarmonicTransformConfig }

func NewHarmonicTransform(config HarmonicTransformConfig) (*HarmonicTransform, error) {
	for _, point := range [...]Point{config.Base, config.Sin, config.Cos, config.ScaleCoupling,
		config.ScaleBase, config.ScaleSin, config.ScaleCos} {
		if !pathFinite(point.X) || !pathFinite(point.Y) {
			return nil, fmt.Errorf("motion: nonfinite harmonic transform parameter")
		}
	}
	if !pathFinite(config.Phase) {
		return nil, fmt.Errorf("motion: nonfinite harmonic transform phase")
	}
	return &HarmonicTransform{config: config}, nil
}

// At samples without changing the clock. Its output is ready for an image
// transform or for applying the same pose to several material passes.
func (transform *HarmonicTransform) At(phase float64) HarmonicTransformPose {
	if transform == nil || !pathFinite(phase) {
		return HarmonicTransformPose{}
	}
	c := transform.config
	angle := phase + c.Phase
	if !pathFinite(angle) {
		return HarmonicTransformPose{}
	}
	sine, cosine := math.Sin(angle), math.Cos(angle)
	sx := c.ScaleBase.X + c.ScaleSin.X*sine + c.ScaleCos.X*cosine
	sy := c.ScaleBase.Y + c.ScaleSin.Y*sine + c.ScaleCos.Y*cosine
	return HarmonicTransformPose{
		X:      c.Base.X + c.Sin.X*sine + c.Cos.X*cosine + c.ScaleCoupling.X*sx,
		Y:      c.Base.Y + c.Sin.Y*sine + c.Cos.Y*cosine + c.ScaleCoupling.Y*sy,
		ScaleX: sx, ScaleY: sy,
	}
}
