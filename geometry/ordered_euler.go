package geometry

import (
	"fmt"
	"math"
)

// OrderedEuler preserves per-axis matrix arithmetic instead of collapsing
// three rotations into one matrix. Order names the axes applied first to last:
// 0=X, 1=Y, 2=Z. Mirror scales the final XYZ coordinates independently.
type OrderedEuler struct {
	order  [3]uint8
	mirror Vec3
	axes   [3]Rotation
}

func NewOrderedEuler(order [3]uint8, mirror Vec3) (*OrderedEuler, error) {
	if order[0] > 2 || order[1] > 2 || order[2] > 2 ||
		order[0] == order[1] || order[0] == order[2] || order[1] == order[2] ||
		math.IsNaN(mirror.X) || math.IsInf(mirror.X, 0) ||
		math.IsNaN(mirror.Y) || math.IsInf(mirror.Y, 0) ||
		math.IsNaN(mirror.Z) || math.IsInf(mirror.Z, 0) {
		return nil, fmt.Errorf("geometry: invalid ordered Euler transform")
	}
	return &OrderedEuler{order: order, mirror: mirror}, nil
}

func (e *OrderedEuler) SetAngles(angles Vec3) {
	e.axes[0] = RotateXYZ(Vec3{X: angles.X})
	e.axes[1] = RotateXYZ(Vec3{Y: angles.Y})
	e.axes[2] = RotateXYZ(Vec3{Z: angles.Z})
}

func (e *OrderedEuler) Apply(p Vec3) Vec3 {
	for _, axis := range e.order {
		p = e.axes[axis].Apply(p)
	}
	p.X *= e.mirror.X
	if e.mirror.Y == -1 {
		p.Y = -p.Y
	} else {
		p.Y *= e.mirror.Y
	}
	if e.mirror.Z == -1 {
		p.Z = -p.Z
	} else {
		p.Z *= e.mirror.Z
	}
	return p
}
