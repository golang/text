// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bidi

import (
	"flag"
	"testing"

	"golang.org/x/text/internal/gen"
	"golang.org/x/text/internal/ucd"
)

var long = flag.Bool("long", false,
	"run time-consuming tests, such as tests that fetch data online")

var labels = []string{
	classArabicLetter:       "AL",
	classArabicNumber:       "AN",
	classParagraphSeparator: "B",
	classBoundaryNeutral:    "BN",
	classCommonSeparator:    "CS",
	classEuropeanNumber:     "EN",
	classEuropeanSeparator:  "ES",
	classEuropeanTerminator: "ET",
	classLeftToRight:        "L",
	classNonspacingMark:     "NSM",
	classOtherNeutral:       "ON",
	classRightToLeft:        "R",
	classSegmentSeparator:   "S",
	classWhiteSpace:         "WS",

	classLeftToRightOverride:   "LRO",
	classRightToLeftOverride:   "RLO",
	classLeftToRightEmbedding:  "LRE",
	classRightToLeftEmbedding:  "RLE",
	classPopDirectionalFormat:  "PDF",
	classLeftToRightIsolate:    "LRI",
	classRightToLeftIsolate:    "RLI",
	classFirstStrongIsolate:    "FSI",
	classPopDirectionalIsolate: "PDI",
}

func TestTables(t *testing.T) {
	if !*long {
		return
	}

	gen.Init()

	trie := newBidiTrie(0)

	parse("BidiBrackets.txt", func(p *ucd.Parser) {
		r1 := p.Rune(0)
		want := p.Rune(1)

		e, _ := trie.lookupString(string(r1))
		if got := entry(e).reverseBracket(r1); got != want {
			t.Errorf("Reverse(%U) = %U; want %U", r1, got, want)
		}
	})

	done := map[rune]bool{}
	test := func(name string, r rune, want string) {
		e, _ := trie.lookupString(string(r))
		if got := labels[entry(e).class(r)]; got != want {
			t.Errorf("%s:%U: got %s; want %s", name, r, got, want)
		}
		done[r] = true
	}

	// Insert the derived BiDi properties.
	parse("extracted/DerivedBidiClass.txt", func(p *ucd.Parser) {
		r := p.Rune(0)
		test("derived", r, p.String(1))
	})
	visitDefaults(func(r rune, c class) {
		if !done[r] {
			test("default", r, labels[c])
		}
	})

}
