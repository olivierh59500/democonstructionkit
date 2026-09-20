// Package scrolltext parses optional scrolling controls without a graphics backend.
package scrolltext

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Kind uint8

const (
	Text Kind = iota
	Font
	Speed
	Pause
	Shape
	Effect
	Scale
	Tracking
)

// Token is either visible text or a control. Control bytes never consume a glyph.
type Token struct {
	Kind          Kind
	Text          string
	Value, Second float64
}

// Decoder recognizes a control at the beginning of input. consumed=0 means text.
// Applications can retain original syntax by supplying their own decoder.
type Decoder func(input string) (token Token, consumed int, err error)

// Parse preserves literal text when decoder is nil, including punctuation that
// resembles commands. UTF-8 errors and malformed recognized controls are rejected.
func Parse(input string, decoder Decoder) ([]Token, error) {
	if !utf8.ValidString(input) {
		return nil, fmt.Errorf("scrolltext: invalid UTF-8")
	}
	var tokens []Token
	var literal strings.Builder
	flush := func() {
		if literal.Len() > 0 {
			tokens = append(tokens, Token{Kind: Text, Text: literal.String()})
			literal.Reset()
		}
	}
	for len(input) > 0 {
		if decoder != nil {
			t, n, err := decoder(input)
			if err != nil {
				return nil, err
			}
			if n < 0 || n > len(input) {
				return nil, fmt.Errorf("scrolltext: invalid decoder consumption")
			}
			if n > 0 {
				flush()
				tokens = append(tokens, t)
				input = input[n:]
				continue
			}
		}
		r, n := utf8.DecodeRuneInString(input)
		literal.WriteRune(r)
		input = input[n:]
	}
	flush()
	return tokens, nil
}

// Braces recognizes {font:name}, {speed:120}, {pause:2}, {shape:twist},
// {effect:gold}, {scale:2,3} and {tracking:4}. '{{' emits a literal '{'.
// Speeds are nonnegative pixels/second; pauses are measured in seconds.
func Braces(input string) (Token, int, error) {
	if strings.HasPrefix(input, "{{") {
		return Token{Kind: Text, Text: "{"}, 2, nil
	}
	if !strings.HasPrefix(input, "{") {
		return Token{}, 0, nil
	}
	end := strings.IndexByte(input, '}')
	if end < 0 {
		return Token{}, 0, fmt.Errorf("scrolltext: unterminated control")
	}
	name, value, ok := strings.Cut(input[1:end], ":")
	if !ok {
		return Token{}, 0, fmt.Errorf("scrolltext: control requires a value")
	}
	t := Token{Text: value}
	switch name {
	case "font":
		t.Kind = Font
	case "shape":
		t.Kind = Shape
	case "effect":
		t.Kind = Effect
	case "speed":
		t.Kind = Speed
	case "pause":
		t.Kind = Pause
	case "scale":
		t.Kind = Scale
	case "tracking":
		t.Kind = Tracking
	default:
		return Token{}, 0, fmt.Errorf("scrolltext: unknown control %q", name)
	}
	if t.Kind == Font || t.Kind == Shape || t.Kind == Effect {
		if value == "" {
			return Token{}, 0, fmt.Errorf("scrolltext: empty %s", name)
		}
		return t, end + 1, nil
	}
	first, second, two := strings.Cut(value, ",")
	var err error
	t.Value, err = strconv.ParseFloat(first, 64)
	if err != nil {
		return Token{}, 0, fmt.Errorf("scrolltext: invalid %s: %w", name, err)
	}
	t.Second = t.Value
	if two {
		if t.Kind != Scale {
			return Token{}, 0, fmt.Errorf("scrolltext: %s expects one value", name)
		}
		t.Second, err = strconv.ParseFloat(second, 64)
		if err != nil {
			return Token{}, 0, err
		}
	}
	return t, end + 1, nil
}

// DomSizes keeps the original Dom intro's ^CsN; syntax, selecting a named face.
// Register faces "0", "1", "2", "3" with their respective scale and advances.
func DomSizes(input string) (Token, int, error) {
	if !strings.HasPrefix(input, "^Cs") {
		return Token{}, 0, nil
	}
	end := strings.IndexByte(input, ';')
	if end < 0 {
		return Token{}, 0, fmt.Errorf("scrolltext: unterminated Dom font control")
	}
	name := input[3:end]
	if name == "" {
		return Token{}, 0, fmt.Errorf("scrolltext: empty font index")
	}
	return Token{Kind: Font, Text: name}, end + 1, nil
}
