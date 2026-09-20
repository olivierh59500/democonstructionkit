// Package indexed provides lossless palette expansion for software demo effects.
package indexed

import (
	"encoding/binary"
	"io"
)

// ExpandRGBA expands palette indices into premultiplied RGBA bytes. Entries are
// packed R | G<<8 | B<<16 | A<<24; callers choose alpha/palette cycling explicitly.
// No allocation or image-size assumption is made on the frame path.
func ExpandRGBA(dst, indices []byte, palette *[256]uint32) error {
	if palette == nil || len(indices) > len(dst)/4 {
		return io.ErrShortBuffer
	}
	for i, index := range indices {
		binary.LittleEndian.PutUint32(dst[i*4:], palette[index])
	}
	return nil
}
