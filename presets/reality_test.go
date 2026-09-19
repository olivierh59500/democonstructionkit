package presets

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestRealityPatternDecodingAndSharedPointers(t *testing.T) {
	container := make([]byte, 16+256)
	binary.LittleEndian.PutUint32(container, 16)
	binary.LittleEndian.PutUint32(container[4:], uint32(len(container)))
	song := container[16:]
	copy(song[44:], "SCRM")
	song[29] = 16
	binary.LittleEndian.PutUint16(song[32:], 1)
	binary.LittleEndian.PutUint16(song[36:], 2)
	binary.LittleEndian.PutUint16(song[97:], 8)
	binary.LittleEndian.PutUint16(song[99:], 8)
	binary.LittleEndian.PutUint16(song[128:], 64)
	for j := 0; j < 64; j++ {
		song[130+j] = byte((j + 2) ^ ((j + 2) * 4))
	}
	original := append([]byte(nil), container...)
	decoded, err := RealityModule(container, 0)
	if err != nil {
		t.Fatal(err)
	}
	ptr := int(binary.LittleEndian.Uint16(decoded[97:])) * 16
	if ptr != int(binary.LittleEndian.Uint16(decoded[99:]))*16 || !bytes.Equal(decoded[ptr+2:ptr+66], make([]byte, 64)) || !bytes.Equal(container, original) {
		t.Fatal("decoding failed or changed input")
	}
	for _, bad := range [][]byte{nil, container[:100]} {
		if _, err := RealityModule(bad, 0); err == nil {
			t.Fatal("accepted truncated container")
		}
	}
}

func TestRealityShortPatternUsesLegacyZeroPadding(t *testing.T) {
	// A note event missing both values reads them as zero, followed by 64 empty rows.
	p := normalizeRealityPattern([]byte{0x20 ^ 10})
	if len(p) != 67 || p[0] != 0x20 || !bytes.Equal(p[1:], make([]byte, 66)) {
		t.Fatal(p)
	}
}
