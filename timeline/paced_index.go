package timeline

import "fmt"

// PacedIndexConfig advances an image or palette index every Every ticks after
// the first change at First. Zero First defaults to Every. Count wraps frames;
// Start selects the initial source index.
type PacedIndexConfig struct{ Count, Every, First, Start int }

// PacedIndex can follow a global scene clock or be stepped only by selected
// movement/music events. Reading Current does not advance the counter.
type PacedIndex struct {
	config PacedIndexConfig
	tick   int
}

func NewPacedIndex(config PacedIndexConfig) (*PacedIndex, error) {
	if config.Count < 1 || config.Count > 1<<20 || config.Every < 1 || config.First < 0 || config.Start < 0 || config.Start >= config.Count {
		return nil, fmt.Errorf("timeline: invalid paced index")
	}
	if config.First == 0 {
		config.First = config.Every
	}
	return &PacedIndex{config: config}, nil
}

func (paced *PacedIndex) At(tick int) int {
	if tick < 0 {
		tick = 0
	}
	if tick < paced.config.First {
		return paced.config.Start
	}
	steps := 1 + (tick-paced.config.First)/paced.config.Every
	return (paced.config.Start + steps%paced.config.Count) % paced.config.Count
}
func (paced *PacedIndex) Current() int { return paced.At(paced.tick) }
func (paced *PacedIndex) Step() int    { paced.tick++; return paced.Current() }
func (paced *PacedIndex) Reset()       { paced.tick = 0 }
