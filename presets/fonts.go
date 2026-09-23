// Package presets describes source-demo assets without embedding or copying them.
package presets

import (
	"fmt"
	"image"
	"strings"

	"github.com/olivierh59500/democonstructionkit/font"
)

// Font identifies an atlas in a demos-root fs.FS. Kind selects its metric recipe.
type Font struct {
	ID, Path, Kind string
	Cell           image.Point
	Columns        int
}

// PhenomenaAlphabet is the authored order used by both DNA scroll atlases.
const PhenomenaAlphabet = " ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!'?/,.-@"

// Fonts returns fresh descriptors for every bitmap font family found in the audit,
// including duplicated assets in go-multiscreen. Vector/packed fonts are not atlases.
func Fonts() []Font {
	return []Font{
		{"3d_doc", "3d_doc/assets/font_out.png", "doc", image.Pt(62, 50), 10},
		{"3d_doc-in", "3d_doc/assets/font_in.png", "doc", image.Pt(62, 50), 10},
		{"3d_doc-intro", "3d_doc/assets/kh6.png", "doc", image.Pt(62, 50), 10},
		{"bilizir-demo", "bilizir-demo/assets/soap-font.png", "soap", image.Pt(32, 32), 10},
		{"dma-3d", "dma-3d/assets/tcb_rep_font.png", "replicants", image.Pt(64, 50), 10},
		{"dma-is-back", "dma-is-back/assets/font.png", "proportional", image.Pt(48, 36), 10},
		{"go-cocoisthebest", "go-cocoisthebest/assets/font.png", "proportional", image.Pt(48, 36), 10},
		{"go-cuddlymenu", "go-cuddlymenu/assets/menu/chrome.png", "chrome", image.Pt(96, 80), 59},
		{"go-dom-intro", "go-dom-intro/assets/rep_ik+_font0.png", "ascii", image.Pt(40, 32), 1},
		{"megatwist", "megatwist/assets/font.png", "proportional", image.Pt(48, 36), 10},
		{"teamg1-demo", "teamg1-demo/assets/font.png", "teamg1", image.Pt(48, 36), 10},
		{"viva_tcb", "viva_tcb/assets/font.png", "ascii", image.Pt(42, 40), 10},
		{"nonameno-demo", "nonameno-demo/assets/font.png", "noname", image.Pt(32, 32), 10},
		{"nonameno-small", "nonameno-demo/assets/font8.png", "noname-small", image.Pt(8, 8), 40},
		{"phenomena-dna-scroll-intro", "phenomena-dna-scroll-intro/assets/font.png", "phenomena", image.Pt(16, 26), 45},
		{"tcb-multi-plane-3d-scroller", "tcb-multi-plane-3d-scroller/assets/bgfont.png", "planes", image.Pt(32, 33), 10},
		{"tcb-replicants-demo", "tcb-replicants-demo/assets/tcb_rep_font.png", "replicants", image.Pt(64, 50), 10},
		{"grodan-kvack-kvack-demo", "grodan-kvack-kvack-demo/assets/bsfont.png", "grodan-big", image.Pt(24, 33), 10},
		{"grodan-up", "grodan-kvack-kvack-demo/assets/upfonts.png", "grodan-up", image.Pt(33, 29), 10},
		{"grodan-small", "grodan-kvack-kvack-demo/assets/lfont.png", "grodan-small", image.Pt(8, 8), 10},
		{"multiscreen-coco", "go-multiscreen/assets/coco/font.png", "proportional", image.Pt(48, 36), 10},
		{"multiscreen-viva", "go-multiscreen/assets/viva/font.png", "ascii", image.Pt(42, 40), 10},
		{"multiscreen-phenomena", "go-multiscreen/assets/phenomena/font.png", "phenomena", image.Pt(16, 26), 45},
		{"multiscreen-tcb", "go-multiscreen/assets/tcb/bgfont.png", "planes", image.Pt(32, 33), 10},
	}
}
func FindFont(id string) (Font, bool) {
	for _, f := range Fonts() {
		if f.ID == id {
			return f, true
		}
	}
	return Font{}, false
}

// Build creates metrics validated against the actual decoded image bounds.
func (s Font) Build(bounds image.Rectangle) (*font.Font, error) {
	if s.Kind == "proportional" || s.Kind == "teamg1" {
		return proportional(bounds, s.Kind == "teamg1")
	}
	if s.Kind == "chrome" {
		return chrome(bounds)
	}
	order := make([]rune, 59)
	for i := range order {
		order[i] = rune(32 + i)
	}
	aliases := map[rune]rune{}
	blanks := ""
	switch s.Kind {
	case "ascii":
	case "doc":
		order = sparse(" !'(),-.0123456789:;?ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		aliases['@'] = ' '
	case "soap":
		order = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789(),.!")
	case "phenomena":
		order = []rune(PhenomenaAlphabet)
	case "replicants":
		order = sparse("!\"'(),-.0123456789:;?ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		order[26] = 0
		order[27] = ':'
		order[28] = ';'
	case "noname":
		order = sparse(" !\"'(),-.0123456789:?ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	case "noname-small":
		order = sparse(" !\"'(),-./0123456789:;?ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	case "planes":
		order = sparse("!(),.:;ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	case "grodan-big":
		blanks = " -"
		order = sparse("!.0123456789:?ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		order[5] = '\''
		order[6] = '"'
		order[7] = '('
		order[8] = ')'
		order[15] = ','
	case "grodan-up":
		blanks = "0123456789 -,'"
		order = sparse("!().:?ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		order[25] = '#'
	case "grodan-small":
		blanks = " -,\""
		order = sparse("!'()./0123456789:?ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	default:
		return nil, fmt.Errorf("presets: unknown font kind %q", s.Kind)
	}
	return font.NewGrid(font.Grid{Bounds: bounds, Cell: s.Cell, Columns: s.Columns, Order: string(order), Blanks: blanks, Uppercase: true, Aliases: aliases})
}
func sparse(characters string) []rune {
	order := make([]rune, 59)
	for _, r := range characters {
		order[int(r)-32] = r
	}
	return order
}

func proportional(bounds image.Rectangle, logo bool) (*font.Font, error) {
	c := font.Config{Bounds: bounds, Glyphs: map[rune]font.Glyph{}, LineHeight: 36, SpaceAdvance: 32, Uppercase: true}
	chars := " !\"'()+,-.0123456789:;<=>?ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for _, r := range chars {
		index := int(r) - 32
		w := 48
		switch {
		case strings.ContainsRune("!',.:;I", r):
			w = 16
		case strings.ContainsRune(" \"()-<=>", r):
			w = 32
		}
		x, y := index%10*48, index/10*36
		c.Glyphs[r] = font.Glyph{Rect: image.Rect(x, y, x+w, y+36).Add(bounds.Min), Advance: float64(w)}
	}
	if logo {
		c.Glyphs['#'] = font.Glyph{Rect: image.Rect(432, 180, 480, 216).Add(bounds.Min), Advance: 48}
	}
	return font.New(c)
}
func chrome(bounds image.Rectangle) (*font.Font, error) {
	widths := chromeTileWidths[:59]
	c := font.Config{Bounds: bounds, Glyphs: map[rune]font.Glyph{}, LineHeight: 80, SpaceAdvance: 96, Uppercase: true}
	for i, width := range widths {
		w := width * 32
		c.Glyphs[rune(i+32)] = font.Glyph{Rect: image.Rect(i*96, 0, i*96+w, 80).Add(bounds.Min), Advance: float64(w)}
	}
	return font.New(c)
}

var chromeTileWidths = [...]int{3, 1, 2, 3, 3, 3, 3, 1, 1, 1, 3, 2, 1, 2, 1, 2, 3, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1, 1, 3, 2, 3, 2, 3, 3, 2, 2, 2, 2, 2, 2, 2, 1, 2, 2, 2, 3, 2, 3, 2, 3, 2, 2, 2, 2, 3, 3, 3, 2, 2, 3, 3, 3, 3, 3, 3}

// CuddlyChromeTiles compiles the menu's variable-width, three-tile glyphs.
func CuddlyChromeTiles(text string) []int {
	result, err := CuddlyChromeAlphabet().Compile(text)
	if err != nil {
		panic(err)
	}
	return result
}

// CuddlyChromeAlphabet returns a fresh editable copy of the menu's tile recipe.
func CuddlyChromeAlphabet() font.TileAlphabet {
	return font.TileAlphabet{First: 32, Widths: append([]int(nil), chromeTileWidths[:]...), Stride: 3, Ignore: "\r\n"}
}
