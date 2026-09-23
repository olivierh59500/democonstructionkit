package scrolling

import (
	"testing"

	kit "github.com/olivierh59500/democonstructionkit"
)

func TestFeedProgressiveEntryKeepsGlyphUntilItsFullWidthEnters(t *testing.T) {
	feed, err := NewFeed(FeedConfig{
		Font: scanlineTestAtlas(t), Text: "A", Width: 12, Height: 2,
		InsertX: 12, Scale: 1, Speed: 2, ProgressiveEntry: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer feed.Close()
	for tick := 1; tick <= 5; tick++ {
		if err := feed.Update(kit.Frame{}); err != nil {
			t.Fatal(err)
		}
		if feed.tile != 0 || feed.x != -2*tick {
			t.Fatalf("tick %d active glyph=%d, entry offset=%d", tick, feed.tile, feed.x)
		}
		if feed.Finished() {
			t.Fatalf("feed finished before its glyph fully entered at tick %d", tick)
		}
	}
	if err := feed.Update(kit.Frame{}); err != nil || !feed.Finished() {
		t.Fatalf("feed did not complete after the full glyph entered: %v", err)
	}
}
