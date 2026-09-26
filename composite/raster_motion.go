package composite

// RasterWrap resets a moving image after it crosses Boundary. Inclusive
// controls whether equality triggers the reset. Direction follows Velocity.
// One authored reset is applied per logical step, preserving scene timing.
type RasterWrap struct {
	Boundary, Restart float64
	Inclusive         bool
}

func rasterNext(position, velocity float64, wrap *RasterWrap) float64 {
	position += velocity
	if wrap == nil || velocity == 0 {
		return position
	}
	if velocity < 0 && (position < wrap.Boundary || wrap.Inclusive && position == wrap.Boundary) ||
		velocity > 0 && (position > wrap.Boundary || wrap.Inclusive && position == wrap.Boundary) {
		return wrap.Restart
	}
	return position
}
