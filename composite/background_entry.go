package composite

// backgroundEntryPeriod keeps the first image alone until its origin reaches
// the viewport's leading edge. A repeated image can then fill the whole view.
func backgroundEntryPeriod(origin, viewportMin, period float64, singleCopy bool) float64 {
	if singleCopy && origin > viewportMin {
		return 0
	}
	return period
}
