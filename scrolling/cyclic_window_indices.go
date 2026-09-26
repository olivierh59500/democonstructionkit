package scrolling

type cyclicPenOffsets interface{ offsetAt(int) float64 }

// cyclicWindowIndices finds only pen positions whose transformed origins can
// intersect the caller's bounds. The source offset sequence is monotonic;
// fractional advances and different fonts require no fixed glyph width.
func cyclicWindowIndices[T cyclicPenOffsets](source T, count int, origin, scale, minimum, maximum float64) (int, int) {
	left, right := 0, count
	for left < right {
		middle := left + (right-left)/2
		if origin+source.offsetAt(middle)*scale > minimum {
			right = middle
		} else {
			left = middle + 1
		}
	}
	first := left
	left, right = first, count
	for left < right {
		middle := left + (right-left)/2
		if origin+source.offsetAt(middle)*scale >= maximum {
			right = middle
		} else {
			left = middle + 1
		}
	}
	end := left
	if end < first {
		end = first
	}
	return first, end
}
