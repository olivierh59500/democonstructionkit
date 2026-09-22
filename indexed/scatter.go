package indexed

import (
	"encoding/binary"
	"fmt"
)

// ScatterConfig maps each source pixel to any number of destination pixels.
// Destinations[Offsets[i]:Offsets[i+1]] belongs to source pixel i. Offsets must
// contain SourceSize+1 entries, start at zero and end at len(Destinations).
// The constructor copies both slices; source and destination layouts may differ.
type ScatterConfig struct {
	SourceSize, DestinationSize int
	Offsets, Destinations       []uint32
}

// ScatterMode selects how source indices combine with a fixed background.
type ScatterMode uint8

const (
	// ScatterReplace copies every source index, including zero.
	ScatterReplace ScatterMode = iota
	// ScatterOverBackground restores the background when the source index is zero.
	ScatterOverBackground
	// ScatterAddBackground adds source and background indices modulo 256.
	ScatterAddBackground
)

type scatterRow struct{ source, end uint32 }

// ScatterMap is a compiled source-to-destination mapping. Source ordering and
// duplicate destinations are preserved: the last source pixel wins. Unmapped
// destinations are untouched. Render needs no allocations or table decoding.
type ScatterMap struct {
	sourceSize, destinationSize int
	rows                        []scatterRow
	destinations                []uint32
}

func NewScatterMap(cfg ScatterConfig) (*ScatterMap, error) {
	if cfg.SourceSize < 0 || cfg.DestinationSize < 1 || uint64(cfg.SourceSize) >= 1<<32 || uint64(cfg.DestinationSize) >= 1<<32 ||
		len(cfg.Offsets) != cfg.SourceSize+1 || uint64(len(cfg.Destinations)) >= 1<<32 {
		return nil, fmt.Errorf("indexed: invalid scatter dimensions")
	}
	if cfg.Offsets[0] != 0 || uint64(cfg.Offsets[cfg.SourceSize]) != uint64(len(cfg.Destinations)) {
		return nil, fmt.Errorf("indexed: invalid scatter offset endpoints")
	}
	for i := 0; i < cfg.SourceSize; i++ {
		if cfg.Offsets[i] > cfg.Offsets[i+1] {
			return nil, fmt.Errorf("indexed: scatter offsets must be ordered")
		}
	}
	for _, destination := range cfg.Destinations {
		if uint64(destination) >= uint64(cfg.DestinationSize) {
			return nil, fmt.Errorf("indexed: scatter destination out of bounds")
		}
	}
	compiled := &ScatterMap{sourceSize: cfg.SourceSize, destinationSize: cfg.DestinationSize,
		destinations: append([]uint32(nil), cfg.Destinations...)}
	for source := 0; source < cfg.SourceSize; source++ {
		if cfg.Offsets[source] != cfg.Offsets[source+1] {
			compiled.rows = append(compiled.rows, scatterRow{uint32(source), cfg.Offsets[source+1]})
		}
	}
	return compiled, nil
}

// DecodeScatterMap16 compiles count/address streams: one little-endian uint16
// count followed by that many little-endian uint16 destinations per source
// pixel. Out-of-canvas addresses are clipped; trailing padding is ignored.
// Truncated streams are rejected before a renderer is returned.
func DecodeScatterMap16(data []byte, sourceSize, destinationSize int) (*ScatterMap, error) {
	if sourceSize < 0 || sourceSize > len(data)/2 || destinationSize < 1 || uint64(destinationSize) >= 1<<32 {
		return nil, fmt.Errorf("indexed: invalid scatter stream dimensions")
	}
	var rows []scatterRow
	previous := uint32(0)
	destinations := make([]uint32, 0, len(data)/2-sourceSize)
	position := 0
	for source := 0; source < sourceSize; source++ {
		if position+2 > len(data) {
			return nil, fmt.Errorf("indexed: truncated scatter count at source %d", source)
		}
		count := int(binary.LittleEndian.Uint16(data[position:]))
		position += 2
		if count > (len(data)-position)/2 {
			return nil, fmt.Errorf("indexed: truncated scatter addresses at source %d", source)
		}
		for i := 0; i < count; i++ {
			destination := uint32(binary.LittleEndian.Uint16(data[position:]))
			position += 2
			if uint64(destination) < uint64(destinationSize) {
				destinations = append(destinations, destination)
			}
		}
		end := uint32(len(destinations))
		if end != previous {
			rows = append(rows, scatterRow{uint32(source), end})
			previous = end
		}
	}
	return &ScatterMap{sourceSize: sourceSize, destinationSize: destinationSize, rows: rows, destinations: destinations}, nil
}

// Render applies the map to caller-owned indexed buffers. All required lengths
// and the mode are checked before output changes. For predictable mapping the
// source must not overlap destination; the background may alias destination.
func (m *ScatterMap) Render(dst, source, background []byte, mode ScatterMode) error {
	if m == nil || len(dst) < m.destinationSize || len(source) < m.sourceSize || mode > ScatterAddBackground {
		return fmt.Errorf("indexed: invalid scatter buffers or mode")
	}
	if mode != ScatterReplace && len(background) < m.destinationSize {
		return fmt.Errorf("indexed: short scatter background")
	}
	// Select the operation once per map, outside the destination-pixel loop.
	switch mode {
	case ScatterReplace:
		start := uint32(0)
		for _, row := range m.rows {
			value := source[row.source]
			for _, destination := range m.destinations[start:row.end] {
				dst[destination] = value
			}
			start = row.end
		}
	case ScatterOverBackground:
		start := uint32(0)
		for _, row := range m.rows {
			value := source[row.source]
			if value == 0 {
				for _, destination := range m.destinations[start:row.end] {
					dst[destination] = background[destination]
				}
			} else {
				for _, destination := range m.destinations[start:row.end] {
					dst[destination] = value
				}
			}
			start = row.end
		}
	case ScatterAddBackground:
		start := uint32(0)
		for _, row := range m.rows {
			value := source[row.source]
			for _, destination := range m.destinations[start:row.end] {
				dst[destination] = background[destination] + value
			}
			start = row.end
		}
	}
	return nil
}
