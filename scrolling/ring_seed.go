package scrolling

// ringSeedIndex chooses an initial glyph without moving any recycled slot.
// Legacy seeding leaves slots beyond the first message blank. Seamless seeding
// reuses the message immediately, retaining the same slot geometry and speed.
func ringSeedIndex(index, length int, seamless bool) int {
	if seamless && length > 0 {
		return index % length
	}
	return index
}

func ringSeedCursor(slots, length int, seamless bool) int {
	if seamless && length > 0 {
		return slots % length
	}
	return slots
}
