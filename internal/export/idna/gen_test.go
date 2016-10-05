// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package idna

import (
	"testing"

	"golang.org/x/text/internal/gen"
	"golang.org/x/text/internal/testtext"
	"golang.org/x/text/internal/ucd"
)

func TestTables(t *testing.T) {
	testtext.SkipIfNotLong(t)

	ucd.Parse(gen.OpenUnicodeFile("idna", "", "IdnaMappingTable.txt"), func(p *ucd.Parser) {
		r := p.Rune(0)
		v, _ := trie.lookupString(string(r))
		x := info(v)

		if got, want := x.category(), catFromEntry(p); got != want {
			t.Errorf("%U:category: got %x; want %x", r, got, want)
		}

		mapped := false
		switch p.String(1) {
		case "mapped", "disallowed_STD3_mapped", "deviation":
			mapped = true
		}
		if x.isMapped() != mapped {
			t.Errorf("%U:isMapped: got %v; want %v", r, x.isMapped(), mapped)
		}
		if !mapped {
			return
		}
		want := string(p.Runes(2))
		got := string(x.appendMapping(nil, string(r)))
		if got != want {
			t.Errorf("%U:mapping: got %+q; want %+q", r, got, want)
		}
	})
}
