package motion

import "math"

// CircleFormation arranges independently indexed sprites around a circle, then
// adds two secondary harmonics and a scale envelope. IndexCount uses the exact
// multiply-then-divide angle recurrence of fixed-count authored formations;
// IndexAngle is used when IndexCount is zero.
type CircleFormation struct {
	RadiusX, RadiusY, IndexAngle                                      float64
	IndexCount                                                        int
	XAmplitude, XRate, XIndexPhase, XPhase                            float64
	YAmplitude, YRate, YIndexPhase, YPhase                            float64
	ScaleBase, ScaleAmplitude, ScaleRate, ScaleIndexPhase, ScalePhase float64
}

func (c CircleFormation) At(time float64, index int) (Point, float64) {
	indexPhase := float64(index) * c.IndexAngle
	if c.IndexCount > 0 {
		indexPhase = float64(index) * math.Pi * 2 / float64(c.IndexCount)
	}
	angle := time + indexPhase
	x := math.Cos(angle) * c.RadiusX
	y := math.Sin(angle) * c.RadiusY
	x += math.Sin(time*c.XRate+float64(index)*c.XIndexPhase+c.XPhase) * c.XAmplitude
	y += math.Cos(time*c.YRate+float64(index)*c.YIndexPhase+c.YPhase) * c.YAmplitude
	scale := c.ScaleBase + c.ScaleAmplitude*math.Sin(time*c.ScaleRate+float64(index)*c.ScaleIndexPhase+c.ScalePhase)
	return Point{X: x, Y: y}, scale
}
