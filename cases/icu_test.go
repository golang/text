// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build icu

package cases

import (
	"path"
	"strings"
	"testing"

	"golang.org/x/text/internal/testtext"
	"golang.org/x/text/language"
)

func TestICUConformance(t *testing.T) {
	// Build test set.
	input := []string{
		"a.a a_a",
		"a\u05d0a",
		"\u05d0'a",
		"a\u03084a",
		"a\u0308a",
		"a3\u30a3a",
		"a\u303aa",
		"a_\u303a_a",
		"1_a..a",
		"1_a.a",
		"a..a.",
		"a--a-",
		"a-a-",
		"a\u200ba",
		"a\u200b\u200ba",
		"a\u00ad\u00ada", // Format
		"a\u00ada",
		"a''a", // SingleQuote
		"a'a",
		"a::a", // MidLetter
		"a:a",
		"a..a", // MidNumLet
		"a.a",
		"a;;a", // MidNum
		"a;a",
		"a__a", // ExtendNumlet
		"a_a",
		"ΟΣ''a",
	}
	add := func(x interface{}) {
		switch v := x.(type) {
		case string:
			input = append(input, v)
		case []string:
			for _, s := range v {
				input = append(input, s)
			}
		}
	}
	for _, tc := range testCases {
		add(tc.src)
		add(tc.lower)
		add(tc.upper)
		add(tc.title)
	}
	for _, tc := range bufferTests {
		add(tc.src)
	}
	for _, tc := range breakTest {
		add(strings.Replace(tc, "|", "", -1))
	}
	for _, tc := range foldTestCases {
		add(tc)
	}

	// Compare ICU to Go.
	for _, c := range []string{"lower", "upper", "title", "fold"} {
		for _, tag := range []string{
			"und", "af", "az", "el", "lt", "nl", "tr",
		} {
			for _, s := range input {
				if exclude(tag, s) {
					continue
				}
				testtext.Run(t, path.Join(c, tag, s), func(t *testing.T) {
					want := doICU(tag, c, s)
					got := doGo(tag, c, s)
					if got != want {
						t.Errorf("\n    in %[3]q (%+[3]q)\n   got %[1]q (%+[1]q)\n  want %[2]q (%+[2]q)", got, want, s)
					}
				})
			}
		}
	}
}

// exclude indicates if a string should be excluded from testing.
func exclude(tag, s string) bool {
	list := []struct{ tags, pattern string }{
		// ICU does not handle leading apostrophe for Dutch and
		// Afrikaans correctly.
		{"af nl", "'n"},
		{"af nl", "'N"},

		// Go terminates the final sigma check after a fixed number of
		// ignorables have been found. This ensures that the algorithm can make
		// progress in a streaming scenario.
		{"", "\u039f\u03a3...............................a"},
		// This also applies to upper in Greek.
		{"el", "\u03bf" + strings.Repeat("\u0321", 29) + "\u0313"},

		// TODO: Go does not handle certain esoteric breaks correctly. This will be
		// fixed once we have a real word break iterator. Alternatively, it
		// seems like we're not too far off from making it work, so we could
		// fix these last steps. But first verify that using a separate word
		// breaker does not hurt performance.
		{"af nl", "a''a"},
		{"", "א'a"},

		// TODO: fix az and tr title.
		{"az tr", ""},

		// TODO: fix lt upper and lower.
		{"lt", ""},

		// TODO: handle Tonos for these letters, which apparently is a thing.
		{"el", "\u0386"}, // GREEK CAPITAL LETTER ALPHA WITH TONOS
		{"el", "\u0389"}, // GREEK CAPITAL LETTER ETA WITH TONOS
		{"el", "\u038A"}, // GREEK CAPITAL LETTER IOTA WITH TONOS

		{"el", "\u0391"}, // GREEK CAPITAL LETTER ALPHA
		{"el", "\u0397"}, // GREEK CAPITAL LETTER ETA
		{"el", "\u0399"}, // GREEK CAPITAL LETTER IOTA

		{"el", "\u03AC"}, // GREEK SMALL LETTER ALPHA WITH TONOS
		{"el", "\u03AE"}, // GREEK SMALL LETTER ALPHA WITH ETA
		{"el", "\u03AF"}, // GREEK SMALL LETTER ALPHA WITH IOTA

		{"el", "\u03B1"}, // GREEK SMALL LETTER ALPHA
		{"el", "\u03B7"}, // GREEK SMALL LETTER ETA
		{"el", "\u03B9"}, // GREEK SMALL LETTER IOTA
	}
	for _, x := range list {
		if x.tags != "" && strings.Index(x.tags, tag) == -1 {
			continue
		}
		if strings.Index(s, x.pattern) != -1 {
			return true
		}
	}
	return false
}

func doGo(tag, caser, input string) string {
	var c Caser
	t := language.MustParse(tag)
	switch caser {
	case "lower":
		c = Lower(t)
	case "upper":
		c = Upper(t)
	case "title":
		c = Title(t)
	case "fold":
		c = Fold()
	}
	return c.String(input)
}
