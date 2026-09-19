package sound

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"sync"
	"testing"
	"time"
)

type ramp struct{ frame, limit int }

func (r *ramp) render(dst []float32) (int, error) {
	n := 0
	for n < len(dst) && r.frame < r.limit {
		dst[n], dst[n+1] = float32(r.frame)/100, -float32(r.frame)/100
		r.frame++
		n += 2
	}
	if r.frame == r.limit {
		return n, io.EOF
	}
	return n, nil
}
func (r *ramp) reset() error { r.frame = 0; return nil }
func (r *ramp) close() error { return nil }

func TestFragmentedReadsAndSampleAccurateSeek(t *testing.T) {
	reference := newStream(&ramp{limit: 6000}, 48000)
	want, err := io.ReadAll(reference)
	if err != nil {
		t.Fatal(err)
	}
	s := newStream(&ramp{limit: 6000}, 48000)
	var got []byte
	for _, size := range []int{1, 2, 3, 5, 4097, 9000, 7, 17} {
		b := make([]byte, size)
		if _, err := io.ReadFull(s, b); err != nil {
			t.Fatal(err)
		}
		got = append(got, b...)
	}
	if !bytes.Equal(got, want[:len(got)]) {
		t.Fatal("partial reads lost stereo continuity")
	}
	for _, offset := range []int64{3, 8003, 7, 0, 32000} {
		if n, err := s.Seek(offset, io.SeekStart); err != nil || n != offset {
			t.Fatal(n, err)
		}
		b := make([]byte, 71)
		if _, err := io.ReadFull(s, b); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(b, want[offset:offset+71]) {
			t.Fatalf("seek %d differs from linear playback", offset)
		}
	}
	before, _ := s.Seek(0, io.SeekCurrent)
	if _, err := s.Seek(-1, io.SeekStart); err == nil {
		t.Fatal("accepted negative offset")
	}
	after, _ := s.Seek(0, io.SeekCurrent)
	if before != after {
		t.Fatal("failed seek changed position")
	}
}
func TestEOFAndClose(t *testing.T) {
	s := newStream(&ramp{limit: 1}, 48000)
	buffer := make([]byte, 20)
	n, err := s.Read(buffer)
	if n != 8 || err != io.EOF {
		t.Fatal(n, err)
	}
	if n, err = s.Read(buffer); n != 0 || err != io.EOF {
		t.Fatal(n, err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Read(buffer); err != io.ErrClosedPipe {
		t.Fatal(err)
	}
}
func TestCallbacksDoNotAllocate(t *testing.T) {
	s := newStream(&ramp{limit: math.MaxInt}, 48000)
	buffer := make([]byte, 4097)
	if n := testing.AllocsPerRun(100, func() {
		if _, err := s.Read(buffer); err != nil {
			panic(err)
		}
	}); n != 0 {
		t.Fatalf("%g allocations per callback", n)
	}
}
func TestCloseDuringReads(t *testing.T) {
	s := newStream(&ramp{limit: math.MaxInt}, 48000)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		b := make([]byte, 33)
		for i := 0; i < 100; i++ {
			_, _ = s.Read(b)
		}
	}()
	_ = s.Close()
	wg.Wait()
}

// Synthetic fixtures contain no music or assets copied from the demos.
func testYM() []byte {
	const frames = 100
	data := make([]byte, 4+frames*14)
	copy(data, "YM3!")
	for f := 0; f < frames; f++ {
		data[4+f] = 80
		data[4+7*frames+f] = 0x3e
		data[4+8*frames+f] = 15
	}
	return data
}
func testMOD() []byte {
	data := make([]byte, 1084+1024+64)
	copy(data, "DCK test signal")
	copy(data[1080:], "M.K.")
	data[950] = 1
	binary.BigEndian.PutUint16(data[42:], 32)
	data[45] = 64
	binary.BigEndian.PutUint16(data[48:], 32)
	copy(data[1084:], []byte{0x01, 0xac, 0x10, 0})
	for i := 0; i < 64; i++ {
		data[1084+1024+i] = byte(int8((i%32 - 16) * 7))
	}
	return data
}
func TestYMAndModuleDecodeSeekAndLoop(t *testing.T) {
	for name, open := range map[string]func() (*Stream, error){
		"ym":  func() (*Stream, error) { return NewYM(testYM(), YMOptions{Loop: true}) },
		"mod": func() (*Stream, error) { return NewModule(testMOD(), ModuleOptions{Interpolation: true}) },
	} {
		t.Run(name, func(t *testing.T) {
			s, err := open()
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			b := make([]byte, 8192)
			if _, err = io.ReadFull(s, b); err != nil {
				t.Fatal(err)
			}
			nonzero := false
			for i := 0; i < len(b); i += 4 {
				value := math.Float32frombits(binary.LittleEndian.Uint32(b[i:]))
				if math.IsNaN(float64(value)) || math.Abs(float64(value)) > 1 {
					t.Fatal(value)
				}
				nonzero = nonzero || value != 0
			}
			if !nonzero {
				t.Fatal("decoder produced only silence")
			}
			if _, err = s.Seek(3, io.SeekStart); err != nil {
				t.Fatal(err)
			}
			again := make([]byte, 317)
			if _, err = io.ReadFull(s, again); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(again, b[3:320]) {
				t.Fatal("decoder seek is not sample accurate")
			}
			buf := make([]byte, 4096)
			if n := testing.AllocsPerRun(20, func() {
				if _, err := s.Read(buf); err != nil {
					panic(err)
				}
			}); n != 0 {
				t.Fatalf("%g callback allocations", n)
			}
		})
	}
	s, err := NewModule(testMOD(), ModuleOptions{Duration: time.Millisecond, Loop: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	pcm := make([]byte, 48*8*2)
	if _, err = io.ReadFull(s, pcm); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pcm[:384], pcm[384:]) {
		t.Fatal("explicit module loop changed samples")
	}
}
func TestInvalidAudioConfiguration(t *testing.T) {
	if _, err := NewYM(nil, YMOptions{SampleRate: 1}); err == nil {
		t.Fatal("accepted invalid rate")
	}
	if _, err := NewYM([]byte("bad"), YMOptions{}); err == nil {
		t.Fatal("accepted invalid YM")
	}
	if _, err := NewModule(testMOD(), ModuleOptions{Loop: true}); err == nil {
		t.Fatal("accepted unbounded loop")
	}
	if _, err := NewModule([]byte("bad"), ModuleOptions{}); err == nil {
		t.Fatal("accepted invalid module")
	}
}
