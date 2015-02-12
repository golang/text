// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate stringer -type=kind
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
)

// TODO
// 1) API proposition for transforms.
// 2) Implement Transforms.
// 3) Implement benchmarks.
// 4) Reduce table size by compressing blocks.
// 5) API proposition for computing display length
//    (approximation, fixed pitch only).
// 6) Implement display length.

// kind indicates the type of width property.
type kind int

const (
	// neutral characters do not occur in legacy East Asian character sets.
	neutral kind = iota

	// ambiguous characters that can be sometimes wide and sometimes narrow and
	// require additional information not contained in the character code to
	// further resolve their width.
	ambiguous

	// wide characters are wide in its usual form. They occur only in the
	// context of East Asian typography. Wide runes may have explicit halfwidth
	// counterparts.
	wide

	// narrow characters are narrow in its usual form. They have fullwidth
	// counterparts.
	narrow

	// fullwidth characters have a compatibility decompositions of type wide
	// that map to a narrow counterpart.
	fullwidth

	// halfwidth characters have a compatibility decomposition of type narrow
	// that map to a wide counterpart, plus U+20A9 ₩ WON SIGN.
	halfwidth
)

func getElem(r rune) elem {
	var buf [4]byte
	n := utf8.EncodeRune(buf[:], r)
	v, _ := newWidthTrie(0).lookup(buf[:n])
	return elem(v)
}

// foldRune maps fullwidth runes to their narrow equivalent and halfwidth runes
// (except for U+20A9 WON SIGN) to their wide equivalent.
func foldRune(r rune) rune {
	v := getElem(r)
	if v&hasMappingMask != 0 && r != wonSign {
		return rune(v &^ tagMappingMask)
	}
	return r
}

// kindOfRune reports the width type of the given rune.
func kindOfRune(r rune) kind {
	v := getElem(r)
	if v < hasMappingMask {
		return kind(v)
	}
	if v&tagMappingMask == tagHalfwidth {
		return halfwidth
	}
	return fullwidth
}
