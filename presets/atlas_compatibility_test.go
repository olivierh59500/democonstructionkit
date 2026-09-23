package presets

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

// Golden metrics were recorded from the demo constructors before migration.
// They cover every explicit glyph, including supported blanks and irregular cells.
func TestAtlasRecipesPreserveCompleteOriginalMappings(t *testing.T) {
	cases := []struct{ id, hash string }{
		{"dma-is-back", "56b044d698568d2f86ba145c07091cb44147901930df258d57332227bc8e4c09"},
		{"teamg1-demo", "a3672201daa33ec792cd51467795ec0587d66a84b2772c63a71dbf31f7f3b81c"},
		{"megatwist", "56b044d698568d2f86ba145c07091cb44147901930df258d57332227bc8e4c09"},
		{"go-cocoisthebest", "56b044d698568d2f86ba145c07091cb44147901930df258d57332227bc8e4c09"},
		{"multiscreen-coco", "56b044d698568d2f86ba145c07091cb44147901930df258d57332227bc8e4c09"},
		{"nonameno-demo", "1e1d024e78e73ea6c027055c473bcedbb05fa40cab23ca2e52ace0acda9ce8e9"},
		{"nonameno-small", "e088ad9557ddae99fcf390e2cf64582799d2f459747290c933d419d0d3933b06"},
		{"grodan-kvack-kvack-demo", "c46c65822fc8a62605ae1a426d7cd7966d4e31027fc5b051cf244245ec1e61d5"},
		{"grodan-up", "958a61925536edc85ab55a048edc0217ee3cee2577db86db154f54676dd205f4"},
		{"grodan-small", "4d5edfc78c1f547384b59f9ae4fa98fb9fc0342d395dd45de4cc1789fe215e23"},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			a, err := FontAtlas(tc.id, nil)
			if err != nil {
				t.Fatal(err)
			}
			var out strings.Builder
			for _, r := range a.Metrics().Characters() {
				_, g, _ := a.Glyph(r)
				fmt.Fprintf(&out, "%d:%d,%d,%d,%d:%g\n", r, g.Rect.Min.X, g.Rect.Min.Y, g.Rect.Dx(), g.Rect.Dy(), g.Advance)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(out.String()))); got != tc.hash {
				t.Fatalf("complete atlas mapping changed: %s", got)
			}
		})
	}
}
