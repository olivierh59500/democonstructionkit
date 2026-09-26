package geometry

import "math"

// RotateXYZScaled returns an XYZ Euler matrix with model scale applied to each
// coefficient after its authored products and sums. This order retains the
// subpixel coordinates of vectorball point clouds within floating-point
// compiler rounding while remaining usable for
// any projected sprite, cube or mesh that supplies angles and a uniform scale.
func RotateXYZScaled(angles Vec3, scale float64) Rotation {
	sinX, cosX := math.Sincos(angles.X)
	sinY, cosY := math.Sincos(angles.Y)
	sinZ, cosZ := math.Sincos(angles.Z)
	return Rotation{
		cosY * cosZ * scale,
		(sinX*sinY*cosZ - cosX*sinZ) * scale,
		(cosX*sinY*cosZ + sinX*sinZ) * scale,
		cosY * sinZ * scale,
		(sinX*sinY*sinZ + cosX*cosZ) * scale,
		(cosX*sinY*sinZ - sinX*cosZ) * scale,
		-sinY * scale,
		sinX * cosY * scale,
		cosX * cosY * scale,
	}
}
