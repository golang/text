// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package width

import (
	"flag"
	"testing"

	"golang.org/x/text/internal/gen"
)

var long = flag.Bool("long", false,
	"run time-consuming tests, such as tests that fetch data online")

const (
	loSurrogate = 0xD800
	hiSurrogate = 0xDFFF
)

func TestTables(t *testing.T) {
	if !gen.IsLocal() && !*long {
		t.Skip("skipping test to prevent downloading; to run use -long or use -local to specify a local source")
	}
	runes := map[rune]elem{}
	getWidthData(func(r rune, tag elem, _ rune) {
		runes[r] = tag
	})
	for r := rune(0); r < 0x10FFFF; r++ {
		if loSurrogate <= r && r <= hiSurrogate {
			continue
		}
		want := Neutral
		switch runes[r] {
		case tagAmbiguous:
			want = EastAsianAmbiguous
		case tagNarrow:
			want = EastAsianNarrow
		case tagWide:
			want = EastAsianWide
		case tagHalfwidth:
			want = EastAsianHalfwidth
		case tagFullwidth:
			want = EastAsianFullwidth
		}
		p := LookupRune(r)
		if got := p.Kind(); got != want {
			t.Errorf("Kind of %U was %s; want %s.", r, got, want)
		}
		if got, want := p.Folded(), foldRunes[r]; got != want {
			t.Errorf("fold of %U was %+q; want %U", r, got, want)
		}
	}
}
