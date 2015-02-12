// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package width

import (
	"flag"
	"testing"
)

var long = flag.Bool("long", false,
	"run time-consuming tests, such as tests that fetch data online")

const (
	loSurrogate = 0xD800
	hiSurrogate = 0xDFFF
)

func TestTables(t *testing.T) {
	if *localDir == "" && !*long {
		t.Skip("skipping test to prevent downloading; to run use -long or use -local to specify a local source")
	}
	// Store the UCD data in a map.
	type data struct {
		tag elem
		alt rune
	}
	runes := map[rune]data{}
	getWidthData(func(r rune, tag elem, alt rune) {
		runes[r] = data{tag, alt}
	})
	for r := rune(0); r < 0x10FFFF; r++ {
		if loSurrogate <= r && r <= hiSurrogate {
			continue
		}
		want := neutral
		switch runes[r].tag {
		case tagAmbiguous:
			want = ambiguous
		case tagNarrow:
			want = narrow
		case tagWide:
			want = wide
		case tagHalfwidth:
			want = halfwidth
		case tagFullwidth:
			want = fullwidth
		}
		if got := kindOfRune(r); got != want {
			t.Errorf("kindOfRune(%U) = %s; want %s.", r, got, want)
		}

		got := foldRune(r)
		if alt := runes[r].alt; alt == 0 {
			if got != r {
				t.Errorf("foldRune(%U) = %+q; want %U", r, got, r)
			}
		} else if got != alt {
			t.Errorf("foldRune(%U) = %+q; want %U", r, got, alt)
		}
	}
}
