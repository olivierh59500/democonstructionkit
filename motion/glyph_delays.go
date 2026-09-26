package motion

import "fmt"

func glyphGridSize(columns, rows int) (int, error) {
	if columns <= 0 || rows <= 0 || columns > 1<<16/rows {
		return 0, fmt.Errorf("motion: invalid glyph delay grid")
	}
	return columns * rows, nil
}

// SerpentineGlyphDelays visits alternating rows in opposite directions.
func SerpentineGlyphDelays(columns, rows int) ([]int, error) {
	count, err := glyphGridSize(columns, rows)
	if err != nil {
		return nil, err
	}
	delays := make([]int, count)
	for y := 0; y < rows; y++ {
		for x := 0; x < columns; x++ {
			index := y*columns + x
			if y%2 == 0 {
				delays[index] = index
			} else {
				delays[index] = y*columns + columns - 1 - x
			}
		}
	}
	return delays, nil
}

// MirroredColumnGlyphDelays alternates downward and upward column waves.
// Adjacent column pairs span twice the grid height, so columns must be even.
func MirroredColumnGlyphDelays(columns, rows int) ([]int, error) {
	count, err := glyphGridSize(columns, rows)
	if err != nil {
		return nil, err
	}
	if columns%2 != 0 {
		return nil, fmt.Errorf("motion: mirrored column grid needs an even width")
	}
	delays := make([]int, count)
	for y := 0; y < rows; y++ {
		for pair := 0; pair < columns/2; pair++ {
			delays[y*columns+pair*2] = pair*rows*2 + y
			delays[y*columns+pair*2+1] = pair*rows*2 + rows*2 - 1 - y
		}
	}
	return delays, nil
}

// SpiralGlyphDelays visits the outside edge clockwise, then moves inward.
// The result is a delay value for each row-major grid cell.
func SpiralGlyphDelays(columns, rows int) ([]int, error) {
	count, err := glyphGridSize(columns, rows)
	if err != nil {
		return nil, err
	}
	delays := make([]int, count)
	left, top, right, bottom, step := 0, 0, columns-1, rows-1, 0
	for left <= right && top <= bottom {
		for x := left; x <= right; x++ {
			delays[top*columns+x] = step
			step++
		}
		top++
		for y := top; y <= bottom; y++ {
			delays[y*columns+right] = step
			step++
		}
		right--
		if top <= bottom {
			for x := right; x >= left; x-- {
				delays[bottom*columns+x] = step
				step++
			}
			bottom--
		}
		if left <= right {
			for y := bottom; y >= top; y-- {
				delays[y*columns+left] = step
				step++
			}
			left++
		}
	}
	return delays, nil
}
