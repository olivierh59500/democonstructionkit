package sound

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"math"
	"testing"
	"testing/fstest"
	"time"
)

func testWave() []byte {
	data := make([]byte, 44+400)
	copy(data, "RIFF")
	binary.LittleEndian.PutUint32(data[4:], uint32(len(data)-8))
	copy(data[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(data[16:], 16)
	binary.LittleEndian.PutUint16(data[20:], 1)
	binary.LittleEndian.PutUint16(data[22:], 2)
	binary.LittleEndian.PutUint32(data[24:], 48000)
	binary.LittleEndian.PutUint32(data[28:], 48000*4)
	binary.LittleEndian.PutUint16(data[32:], 4)
	binary.LittleEndian.PutUint16(data[34:], 16)
	copy(data[36:], "data")
	binary.LittleEndian.PutUint32(data[40:], 400)
	for i := 0; i < 100; i++ {
		binary.LittleEndian.PutUint16(data[44+i*4:], uint16(i*31))
		binary.LittleEndian.PutUint16(data[46+i*4:], uint16(i*31))
	}
	return data
}

func testXM() []byte {
	b := make([]byte, 345)
	copy(b, "Extended Module: ")
	copy(b[17:], "DCK silent module")
	b[37] = 0x1a
	copy(b[38:], "DCK fixture")
	binary.LittleEndian.PutUint16(b[58:], 0x0104)
	binary.LittleEndian.PutUint32(b[60:], 276)
	for offset, value := range map[int]uint16{64: 1, 68: 1, 70: 1, 76: 6, 78: 125, 341: 64} {
		binary.LittleEndian.PutUint16(b[offset:], value)
	}
	binary.LittleEndian.PutUint32(b[336:], 9)
	return b
}
func testIT() []byte {
	b := make([]byte, 198)
	copy(b, "IMPM")
	copy(b[4:], "DCK silent module")
	for offset, value := range map[int]uint16{32: 2, 38: 1, 40: 0x0214, 42: 0x0200} {
		binary.LittleEndian.PutUint16(b[offset:], value)
	}
	b[48], b[49], b[50], b[51], b[52] = 128, 48, 6, 125, 128
	for i := 0; i < 64; i++ {
		b[64+i] = 128
	}
	b[64], b[128] = 32, 64
	b[192], b[193] = 0, 255
	return b
}

func TestOpenAllTrackerFamiliesThroughOneEntry(t *testing.T) {
	s3m := silentS3M(false)
	s3m[28] = 0x1a
	s3m[64] = 0
	for _, test := range []struct {
		kind Format
		data []byte
	}{{FormatMOD, testMOD()}, {FormatS3M, s3m}, {FormatXM, testXM()}, {FormatIT, testIT()}} {
		t.Run(string(test.kind), func(t *testing.T) {
			s, err := Open("asset.bin", test.data, Options{Loop: true})
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			if s.Metadata().Format != test.kind {
				t.Fatal(s.Metadata())
			}
			if n, err := io.ReadFull(s, make([]byte, 8192)); err != nil || n != 8192 {
				t.Fatal(n, err)
			}
		})
	}
}

func TestOpenSelectsFormatFromContentAndPreservesPCM(t *testing.T) {
	for _, tt := range []struct {
		name   string
		data   []byte
		format Format
	}{{"misleading.s3m", testYM(), FormatYM}, {"misleading.fc", testYM(), FormatYM}, {"wrong.fc", testMOD(), FormatMOD}, {"misleading.ym", testMOD(), FormatMOD}, {"RECORD.WAV", testWave(), FormatWAV}} {
		f, err := Open(tt.name, tt.data, Options{PCMFormat: PCM16, Gain: .5})
		if err != nil {
			t.Fatal(err)
		}
		if f.Metadata().Format != tt.format || f.Format() != PCM16 || f.SampleRate() != 48000 {
			t.Fatal(f.Metadata(), f.Format())
		}
		data := make([]byte, 16)
		if _, err := io.ReadFull(f, data); err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
	// Positive module signatures win even when a title begins with MP3-like text.
	mod := testMOD()
	copy(mod, "ID3 module title")
	if f, err := Detect("song.mp3", mod); err != nil || f != FormatMOD {
		t.Fatal(f, err)
	}
}

func TestOpenNaturalModuleEndAndLoop(t *testing.T) {
	once, err := Open("song.mod", testMOD(), Options{Interpolation: true})
	if err != nil {
		t.Fatal(err)
	}
	defer once.Close()
	pcm, err := io.ReadAll(io.LimitReader(once, 8<<20))
	if err != nil {
		t.Fatal(err)
	}
	if len(pcm) == 0 || len(pcm) >= 8<<20 {
		t.Fatal("module did not stop at its natural end", len(pcm))
	}
	if _, err := once.Read(make([]byte, 8)); err != io.EOF {
		t.Fatal("natural end is not EOF", err)
	}
	loop, err := Open("song.mod", testMOD(), Options{Interpolation: true, Loop: true})
	if err != nil {
		t.Fatal(err)
	}
	defer loop.Close()
	repeated := make([]byte, len(pcm)*2+37)
	if _, err := io.ReadFull(loop, repeated); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(repeated[:len(pcm)], pcm) || !bytes.Equal(repeated[len(pcm):len(pcm)*2], pcm) || !bytes.Equal(repeated[len(pcm)*2:], pcm[:37]) {
		t.Fatal("natural module loop changed PCM")
	}
	if _, err := loop.Seek(31, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	b := make([]byte, 257)
	io.ReadFull(loop, b)
	if !bytes.Equal(b, pcm[31:288]) {
		t.Fatal("module seek lost decoder state")
	}
}

func TestOpenGzipRecordingLoopPrefixAndMetadata(t *testing.T) {
	var compressed bytes.Buffer
	z := gzip.NewWriter(&compressed)
	z.Write(testWave())
	z.Close()
	s, err := OpenFS(fstest.MapFS{"music.wav.gz": &fstest.MapFile{Data: compressed.Bytes()}}, "music.wav.gz", Options{PCMFormat: PCM16, Loop: true, LoopStartFrame: 25})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if s.Metadata().Format != FormatWAV || s.Metadata().Duration != 100*time.Second/48000 {
		t.Fatal(s.Metadata())
	}
	if s.Length() != 400 {
		t.Fatal("duration rounding truncated the exact PCM length", s.Length())
	}
	got := make([]byte, 400+300)
	if _, err := io.ReadFull(s, got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got[:400], testWave()[44:]) || !bytes.Equal(got[400:], testWave()[44+100:]) {
		t.Fatal("recording loop replayed the one-time prefix")
	}
}

func TestVolumeChangeDoesNotTearPartialStereoFrames(t *testing.T) {
	s := newStream(&ramp{limit: 6000}, 48000)
	s.format = PCM16
	// Consume frame zero plus the first byte of frame one at full volume.
	initial := make([]byte, 5)
	io.ReadFull(s, initial)
	s.SetVolume(.5)
	got := make([]byte, 7)
	io.ReadFull(s, got)
	whole := append(initial, got...)
	want := []int16{0, 0, 327, -327, 327, -327}
	for i, v := range want {
		if sample := int16(binary.LittleEndian.Uint16(whole[i*2:])); sample != v {
			t.Fatal(i, sample, v)
		}
	}
	if s.Volume() != .5 {
		t.Fatal(s.Volume())
	}
}

func TestOpenQuantizedFloatMatchesPCM16AndDoesNotAllocate(t *testing.T) {
	integer, _ := Open("song.ym", testYM(), Options{PCMFormat: PCM16, Gain: .5, Loop: true})
	defer integer.Close()
	floating, _ := Open("song.ym", testYM(), Options{Gain: .5, Quantize16: true, Loop: true})
	defer floating.Close()
	a, b := make([]byte, 4096), make([]byte, 8192)
	io.ReadFull(integer, a)
	io.ReadFull(floating, b)
	for i := 0; i < len(a)/2; i++ {
		want := float32(int16(binary.LittleEndian.Uint16(a[i*2:]))) / 32768
		got := math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
		if got != want {
			t.Fatal(i, got, want)
		}
	}
	if n := testing.AllocsPerRun(100, func() {
		if _, err := integer.Read(a); err != nil {
			panic(err)
		}
	}); n != 0 {
		t.Fatal(n)
	}
}

func TestOpenInvalidFilesAndOptionsFail(t *testing.T) {
	for _, name := range []string{"bad.ym", "bad.mod", "bad.xm", "bad.s3m", "bad.it", "bad.wav", "bad.mp3", "bad.ogg", "bad.sndh"} {
		if s, err := Open(name, []byte("invalid"), Options{}); err == nil {
			s.Close()
			t.Fatal(name)
		}
	}
	for _, c := range []Options{{Gain: math.NaN()}, {Gain: -1}, {PCMFormat: 9}, {BlockFrames: -1}, {BlockFrames: 1025}, {Track: 2}, {LoopStartFrame: -1}, {ModuleProfile: 9}} {
		if s, err := Open("test.ym", testYM(), c); err == nil {
			s.Close()
			t.Fatal(c)
		}
	}
}

func TestOpenDecodeQuantumPreservesRegisterCueClock(t *testing.T) {
	s, err := Open("track.ym", registerYM(), Options{PCMFormat: PCM16, BlockFrames: 800, Loop: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	buf := make([]byte, 800*4)
	io.ReadFull(s, buf)
	if registers, ok := s.YMRegisters(); !ok || registers != registerFrame(0) {
		t.Fatal(registers, ok)
	}
	io.ReadFull(s, buf)
	if registers, ok := s.YMRegisters(); !ok || registers != registerFrame(1) {
		t.Fatal(registers, ok)
	}
}
