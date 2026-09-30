package geometry

import "math"

// ClipNearSegment retains the visible part of a line at the camera's near
// plane. Wholly hidden segments return false. Endpoint order is preserved.
func ClipNearSegment(a, b Vec3, near float64) (Vec3, Vec3, bool) {
	if near <= 0 || math.IsNaN(near) || math.IsInf(near, 0) || a.Z < near && b.Z < near {
		return a, b, false
	}
	if a.Z < near {
		a = Lerp(a, b, (near-a.Z)/(b.Z-a.Z))
		a.Z = near
	}
	if b.Z < near {
		b = Lerp(b, a, (near-b.Z)/(a.Z-b.Z))
		b.Z = near
	}
	return a, b, true
}
