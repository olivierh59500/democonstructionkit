package composite

import (
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestVerticalStripTrainSamplesCachedRowsAndSkipsSourceOverflow(t *testing.T) {
	source := ebiten.NewImage(1, 8)
	defer source.Deallocate()
	for y := 0; y < 8; y++ {
		source.Set(0, y, color.RGBA{R: uint8(y * 25), A: 255})
	}
	wrap := &RasterWrap{Boundary: 6, Restart: 0, Inclusive: true}
	train, err := NewVerticalStripTrain(VerticalStripTrainConfig{
		Image: source, SourceStride: 2, StripHeight: 1, SampleStep: 2,
		Count: 3, RowStep: 1, Velocity: 2, Wrap: wrap,
	})
	if err != nil {
		t.Fatal(err)
	}
	wrap.Restart = 99 // Construction owns the wrap recipe.
	train.Step()
	if train.Phase() != 2 {
		t.Fatalf("first phase %v", train.Phase())
	}
	dst := ebiten.NewImage(1, 3)
	defer dst.Deallocate()
	train.Draw(dst)
	for row, sourceRow := range []int{2, 4, 6} {
		got := color.RGBAModel.Convert(dst.At(0, row)).(color.RGBA)
		if got.R != uint8(sourceRow*25) {
			t.Fatalf("destination row %d samples source %d: %+v", row, sourceRow, got)
		}
	}
	train.Step()
	dst.Clear()
	train.Draw(dst)
	if got := color.RGBAModel.Convert(dst.At(0, 2)).(color.RGBA); got.A != 0 {
		t.Fatalf("overflow row should be transparent: %+v", got)
	}
	train.Step()
	if train.Phase() != 0 {
		t.Fatalf("inclusive wrap phase %v", train.Phase())
	}
	if err := train.SetVelocity(math.NaN()); err == nil {
		t.Fatal("accepted nonfinite strip velocity")
	}
	if err := train.SetVelocity(1); err != nil {
		t.Fatal(err)
	}
	train.Step()
	if train.Phase() != 1 {
		t.Fatalf("live velocity cue produced phase %v", train.Phase())
	}
	if got := testing.AllocsPerRun(100, train.Step); got != 0 {
		t.Fatalf("vertical strip step allocates %.2f objects", got)
	}
}
