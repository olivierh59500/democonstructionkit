package presets

import (
	"encoding/binary"
	"fmt"
)

// RealityModule extracts a standard S3M from Second Reality's two-song Reality.FC
// container. Song 0 is Skaven; song 1 is Purple Motion. FC stores XOR-obfuscated
// pattern payloads, as documented by the source demo's internal/st3/readPatByte.
// Short patterns are padded with the zero bytes supplied by that legacy reader.
// The input remains unchanged. This converts storage, not ST3 synchronization flags.
func RealityModule(container []byte, song int) ([]byte, error) {
	if song < 0 || song > 1 || len(container) < 8 {
		return nil, fmt.Errorf("presets: invalid Reality.FC song/header")
	}
	start := int64(binary.LittleEndian.Uint32(container[song*4:]))
	end := int64(len(container))
	if song == 0 {
		end = int64(binary.LittleEndian.Uint32(container[4:]))
	}
	if start < 8 || end <= start || end > int64(len(container)) || end-start < 96 {
		return nil, fmt.Errorf("presets: invalid Reality.FC offsets")
	}
	data := append([]byte(nil), container[start:end]...)
	if string(data[44:48]) != "SCRM" || data[29] != 16 {
		return nil, fmt.Errorf("presets: missing S3M header")
	}
	orders := int(binary.LittleEndian.Uint16(data[32:]))
	samples := int(binary.LittleEndian.Uint16(data[34:]))
	patterns := int(binary.LittleEndian.Uint16(data[36:]))
	table := 96 + orders + samples*2
	if patterns > 255 || table+patterns*2 > len(data) {
		return nil, fmt.Errorf("presets: truncated Reality.FC parapointers")
	}
	seen := map[int]int{}
	sourceSize := len(data)
	for i := 0; i < patterns; i++ {
		ptr := int(binary.LittleEndian.Uint16(data[table+2*i:])) * 16
		if ptr == 0 {
			continue
		}
		if target, ok := seen[ptr]; ok {
			binary.LittleEndian.PutUint16(data[table+2*i:], uint16(target/16))
			continue
		}
		if ptr+2 > sourceSize {
			return nil, fmt.Errorf("presets: pattern %d pointer out of range", i)
		}
		length := int(binary.LittleEndian.Uint16(data[ptr:]))
		if ptr+2+length > sourceSize {
			return nil, fmt.Errorf("presets: pattern %d payload out of range", i)
		}
		payload := normalizeRealityPattern(data[ptr+2 : ptr+2+length])
		padding := (16 - len(data)%16) % 16
		target := len(data) + padding
		if target/16 > 65535 || len(payload) > 65535 {
			return nil, fmt.Errorf("presets: normalized S3M exceeds parapointer limits")
		}
		data = append(data, make([]byte, padding+2)...)
		binary.LittleEndian.PutUint16(data[target:], uint16(len(payload)))
		data = append(data, payload...)
		binary.LittleEndian.PutUint16(data[table+2*i:], uint16(target/16))
		seen[ptr] = target
	}
	return data, nil
}

func normalizeRealityPattern(encoded []byte) []byte {
	index := 0
	read := func() byte {
		i := index
		index++
		if i >= len(encoded) {
			return 0
		}
		return encoded[i] ^ byte((i+2)^((i+2)*4))
	}
	decoded := make([]byte, 0, len(encoded)+64)
	for row := 0; row < 64; {
		flag := read()
		decoded = append(decoded, flag)
		if flag == 0 {
			row++
			continue
		}
		count := 0
		if flag&32 != 0 {
			count += 2
		}
		if flag&64 != 0 {
			count++
		}
		if flag&128 != 0 {
			count += 2
		}
		for i := 0; i < count; i++ {
			decoded = append(decoded, read())
		}
	}
	return decoded
}
