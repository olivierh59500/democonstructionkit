package motion

import "fmt"

// WaveAppend places a write after the current table. Other positions are
// absolute; later writes may replace earlier samples or extend zero-filled gaps.
const WaveAppend = -1

// WaveWrite places one locally sampled section into a mutable wave table.
type WaveWrite struct {
	At      int
	Section WaveSection
}

// CompileWaveProgram compiles an ordered series of append and overwrite
// operations. This preserves authored lookup tables whose writes overlap or
// leave blank frames while keeping all trigonometry out of the render loop.
func CompileWaveProgram(writes ...WaveWrite) ([]float64, error) {
	if len(writes) == 0 {
		return nil, fmt.Errorf("motion: empty wave program")
	}
	var values []float64
	for index, write := range writes {
		part, err := CompileWaveTable(write.Section)
		if err != nil {
			return nil, fmt.Errorf("motion: wave write %d: %w", index, err)
		}
		at := write.At
		if at == WaveAppend {
			at = len(values)
		}
		if at < 0 || at > 1<<20-len(part) {
			return nil, fmt.Errorf("motion: invalid wave write position")
		}
		end := at + len(part)
		if end > len(values) {
			values = append(values, make([]float64, end-len(values))...)
		}
		copy(values[at:end], part)
	}
	return values, nil
}
