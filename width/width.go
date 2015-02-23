// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate stringer -type=Kind
//go:generate go run gen.go gen_common.go gen_trieval.go

// Package width provides functionality for handling different widths in text.
//
// Wide characters behave like ideographs; they tend to allow line breaks after
// each character and remain upright in vertical text layout. Narrow characters
// are kept together in words or runs that are rotated sideways in vertical text
// layout.
//
// For more information, see http://unicode.org/reports/tr11/.
package width

import (
	"unicode/utf8"

	"golang.org/x/text/transform"
)

// TODO
// 1) Reduce table size by compressing blocks.
// 2) API proposition for computing display length
//    (approximation, fixed pitch only).
// 3) Implement display length.

// Kind indicates the type of width property as defined in http://unicode.org/reports/tr11/.
type Kind int

const (
	// Neutral characters do not occur in legacy East Asian character sets.
	Neutral Kind = iota

	// EastAsianAmbiguous characters that can be sometimes wide and sometimes
	// narrow and require additional information not contained in the character
	// code to further resolve their width.
	EastAsianAmbiguous

	// EastAsianWide characters are wide in its usual form. They occur only in
	// the context of East Asian typography. These runes may have explicit
	// halfwidth counterparts.
	EastAsianWide

	// EastAsianNarrow characters are narrow in its usual form. They have
	// fullwidth counterparts.
	EastAsianNarrow

	// EastAsianFullwidth characters have a compatibility decompositions of type
	// wide that map to a narrow counterpart.
	EastAsianFullwidth

	// EastAsianHalfwidth characters have a compatibility decomposition of type
	// narrow that map to a wide counterpart, plus U+20A9 ₩ WON SIGN.
	EastAsianHalfwidth
)

// TODO: the generated tries need to return size 1 for invalid runes for the
// width to be computed correctly (each byte should render width 1)

var trie = newWidthTrie(0)

// Lookup reports the Properties of the first rune in b and the number of bytes
// of its UTF-8 encoding.
func Lookup(b []byte) (p Properties, size int) {
	v, sz := trie.lookup(b)
	return Properties{elem(v)}, sz
}

// LookupString reports the Properties of the first rune in s and the number of
// bytes of its UTF-8 encoding.
func LookupString(s string) (p Properties, size int) {
	v, sz := trie.lookupString(s)
	return Properties{elem(v)}, sz
}

// LookupRune reports the Properties of rune r.
func LookupRune(r rune) Properties {
	var buf [4]byte
	n := utf8.EncodeRune(buf[:], r)
	v, _ := trie.lookup(buf[:n])
	return Properties{elem(v)}
}

// Properties provides access to width properties of a rune.
type Properties struct {
	elem elem
}

// Kind returns the Kind of a rune as defined in Unicode TR #11.
// See http://unicode.org/reports/tr11/ for more details.
func (p Properties) Kind() Kind {
	v := p.elem
	if v < hasMappingMask {
		return Kind(v)
	}
	if v&tagMappingMask == tagHalfwidth {
		return EastAsianHalfwidth
	}
	return EastAsianFullwidth
}

// Folded returns the folded variant of a rune or 0 if the rune is canonical.
func (p Properties) Folded() rune {
	if p.elem > hasMappingMask {
		return rune(p.elem &^ tagMappingMask)
	}
	return 0
}

// TODO for Properties:
// - Add Fullwidth/Halfwidth or Inverted methods for computing variants
// mapping.
// - Add width information (including information on non-spacing runes).

// Transformer implements the transform.Transformer interface.
type Transformer struct {
	t transform.Transformer
}

// Reset implements the transform.Transformer interface.
func (t Transformer) Reset() { t.t.Reset() }

// Transform implements the Transformer interface.
func (t Transformer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	return t.t.Transform(dst, src, atEOF)
}

// Bytes returns a new byte slice with the result of applying t to b.
func (t Transformer) Bytes(b []byte) []byte {
	b, _, _ = transform.Bytes(t, b)
	return b
}

// String returns a string with the result of applying t to s.
func (t Transformer) String(s string) string {
	s, _, _ = transform.String(t, s)
	return s
}

// Fold returns a transform that maps all runes to their canonical width.
//
// Note that the NFKC and NFKD transforms in golang.org/x/text/unicode/norm
// provide a more generic folding mechanism.
func Fold() Transformer {
	return Transformer{foldTransform{}}
}

// TODO: Widening or narrowing all possible runes:
//   - Widen   Convert anything with a wide variant to wide (including to Fullwidth).
//   - Narrow  Convert anything with a narrow variant to narrow (including Halfwidth).
//
// We opt to not support folding only fullwidth or halfwidth variants. We could
// later do so with options, but the preferred approach is allow applying
// filters to transforms.
