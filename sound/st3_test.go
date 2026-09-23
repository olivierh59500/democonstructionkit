package sound

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
	"time"
)

func silentS3M(packed bool) []byte {
	data := make([]byte, 112+2+64)
	copy(data, "Synthetic tracker timing")
	data[29] = 16
	binary.LittleEndian.PutUint16(data[32:], 2)
	binary.LittleEndian.PutUint16(data[36:], 1)
	copy(data[44:], "SCRM")
	data[48], data[49], data[50], data[51] = 64, 6, 125, 64
	for i := 64; i < 96; i++ {
		data[i] = 255
	}
	data[96], data[97] = 0, 255
	binary.LittleEndian.PutUint16(data[98:], 7)
	binary.LittleEndian.PutUint16(data[112:], 64)
	if packed {
		for i := 0; i < 64; i++ {
			data[114+i] = byte((i + 2) ^ ((i + 2) * 4))
		}
	}
	return data
}

func TestCompatibilityProfilePackedDetectionAndDuration(t *testing.T) {
	for _, packed := range []bool{false, true} {
		name := "music.s3m"
		profile := ModuleProfileScreamTracker3
		if packed {
			name = "music.fc"
			profile = ModuleDefault
		}
		options := Options{SampleRate: 8000, PCMFormat: PCM16, Duration: time.Millisecond, ModuleProfile: profile}
		s, err := Open(name, silentS3M(packed), options)
		if err != nil {
			t.Fatal(err)
		}
		if s.Metadata().Format != FormatS3M || s.Length() != 32 {
			t.Fatal("tracker metadata lost")
		}
		first, err := io.ReadAll(s)
		if err != nil || len(first) != 32 {
			t.Fatal("explicit tracker duration was ignored", len(first), err)
		}
		if _, ok := s.TrackerPositionAt(0); !ok {
			t.Fatal("tracker markers unavailable")
		}
		s.Close()
		if _, ok := s.TrackerPositionAt(0); ok {
			t.Fatal("closed stream retained markers")
		}
		options.Loop = true
		s, err = Open(name, silentS3M(packed), options)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		got := make([]byte, 96)
		if _, err = io.ReadFull(s, got); err != nil {
			t.Fatal(err)
		}
		for cycle := 0; cycle < 3; cycle++ {
			if !bytes.Equal(first, got[cycle*32:(cycle+1)*32]) {
				t.Fatal("loop changed tracker PCM")
			}
		}
		if _, err = s.Seek(0, io.SeekStart); err != nil {
			t.Fatal(err)
		}
		if _, err = io.ReadFull(s, got[:32]); err != nil || !bytes.Equal(first, got[:32]) {
			t.Fatal("seek did not restore the beginning")
		}
		options.Duration = time.Nanosecond
		if s, err := Open(name, silentS3M(packed), options); err == nil {
			s.Close()
			t.Fatal("accepted sub-frame duration")
		}
	}
}
