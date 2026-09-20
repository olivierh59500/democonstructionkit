package motion

import "math"

// CoupledOrbit combines a shared depth oscillation with independent X/Y waves.
// PhaseStep controls advancement per sampled point, preserving interleaved groups.
// QOffset/QScale remain below 255 to keep both modulation denominators finite.
type CoupledOrbit struct {
	XIncrement, YIncrement, ZIncrement, QIncrement   float64
	XOffset, YOffset, ZOffset, QOffset, QScale       float64
	CenterX, CenterY, Radius, DepthRadius, PhaseStep float64
	x, y, z                                          float64
}

func (o *CoupledOrbit) Next(index float64) (float64, float64) {
	o.x += o.XIncrement * o.PhaseStep
	o.y += o.YIncrement * o.PhaseStep
	o.z += o.ZIncrement * o.PhaseStep
	z := o.DepthRadius * math.Sin(o.z+index*o.ZOffset*.02)
	xq := o.QIncrement / (255 - math.Min(254, o.QOffset))
	yq := o.QIncrement / (255 - math.Min(254, o.QScale))
	return o.CenterX + z*2 + o.Radius*math.Sin(o.x+index*o.XOffset*(.02*xq)), o.CenterY + z/2 + o.Radius*math.Cos(o.y+index*o.YOffset*(.02*yq))
}
