package scrolling

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/font"
)

func scanlineTestAtlas(t *testing.T) *Atlas {
	t.Helper()
	img := ebiten.NewImage(30, 2)
	img.Fill(color.White)
	metrics, err := font.New(font.Config{
		Bounds: image.Rect(0, 0, 30, 2), LineHeight: 2, SpaceAdvance: 10,
		Glyphs: map[rune]font.Glyph{
			'A': {Rect: image.Rect(0, 0, 10, 2), Advance: 10},
			'B': {Rect: image.Rect(10, 0, 20, 2), Advance: 10},
			'C': {Rect: image.Rect(20, 0, 30, 2), Advance: 10},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	atlas, err := NewAtlas(img, metrics)
	if err != nil {
		t.Fatal(err)
	}
	return atlas
}

func TestScanlineCursorContinuesThroughThreeMessages(t *testing.T) {
	config := ScanlineConfig{
		Font: scanlineTestAtlas(t), Text: "ABC", Wave: []int{1},
		ViewportWidth: 12, ViewportHeight: 6, SurfaceWidth: 24,
		SourceRows: 3, RowHeight: 2, StripHeight: 2,
		Scale: 1, WaveStep: 1, CursorRows: 1,
	}
	scroll, err := NewScanlineScroll(config)
	if err != nil {
		t.Fatal(err)
	}
	defer scroll.Close()
	if err := scroll.Update(kit.Frame{Tick: 95}); err != nil {
		t.Fatal(err)
	}
	if scroll.letter != 9 || scroll.decal != 90 {
		t.Fatalf("third message cursor = (%d, %d), want (9, 90)", scroll.letter, scroll.decal)
	}
	if state := scroll.State(); state.Letter != 9 || state.Decal != 90 || state.WaveStart != 95 {
		t.Fatalf("public cursor state = %+v", state)
	}
	dst := ebiten.NewImage(12, 6)
	scroll.Draw(dst)
	if scroll.text.Window(scroll.letter, 12).End <= scroll.letter {
		t.Fatal("third message produced an empty glyph window")
	}
	if err := scroll.Update(kit.Frame{Tick: 5}); err != nil {
		t.Fatal(err)
	}
	if scroll.letter != 0 || scroll.decal != 0 {
		t.Fatalf("backward seek cursor = (%d, %d), want (0, 0)", scroll.letter, scroll.decal)
	}
}

func TestScanlineClampAndVariableSpeed(t *testing.T) {
	config := ScanlineConfig{
		Font: scanlineTestAtlas(t), Text: "ABC", Wave: []int{1},
		ViewportWidth: 12, ViewportHeight: 6, SurfaceWidth: 24,
		SourceRows: 3, RowHeight: 2, StripHeight: 2,
		Scale: 1, WaveStep: 1, CursorRows: 1,
		ClampCursor: true, UseTime: true,
	}
	scroll, err := NewScanlineScroll(config)
	if err != nil {
		t.Fatal(err)
	}
	defer scroll.Close()
	if err := scroll.Update(kit.Frame{Time: 95.5}); err != nil {
		t.Fatal(err)
	}
	if scroll.letter != 2 || scroll.decal != 20 {
		t.Fatalf("clamped cursor = (%d, %d), want (2, 20)", scroll.letter, scroll.decal)
	}
	if err := scroll.Update(kit.Frame{Time: -1}); err == nil {
		t.Fatal("negative time was accepted")
	}
	if scroll.letter != 2 || scroll.decal != 20 {
		t.Fatal("rejected time changed the selected character")
	}
}

func TestFeedCompletesWithoutAdvancingDuringDraw(t *testing.T) {
	atlas := scanlineTestAtlas(t)
	feed, err := NewFeed(FeedConfig{Font: atlas, Text: "aB", Width: 12, Height: 2, Margin: 10, Scale: 1, Speed: 2, UppercaseASCII: true})
	if err != nil {
		t.Fatal(err)
	}
	defer feed.Close()
	capitalA, _, _ := atlas.ExactGlyph('A')
	if feed.glyphs[0] != capitalA {
		t.Fatal("ASCII uppercase mapping did not resolve the supplied atlas")
	}
	for i := 0; i < 30 && !feed.Finished(); i++ {
		if err := feed.Update(kit.Frame{}); err != nil {
			t.Fatal(err)
		}
	}
	if !feed.Finished() {
		t.Fatal("finite feed did not finish")
	}
	image := feed.Image()
	position, letter := feed.x, feed.letter
	feed.Draw(ebiten.NewImage(12, 2))
	feed.Draw(ebiten.NewImage(12, 2))
	if feed.Image() != image || feed.x != position || feed.letter != letter {
		t.Fatal("drawing the same feed frame advanced its state")
	}
	if err := feed.Update(kit.Frame{}); err != nil || feed.x != position || feed.letter != letter {
		t.Fatal("finished feed advanced after completion")
	}
}
