// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package colltab

import (
	"testing"

	"golang.org/x/text/collate/colltab"
)

const (
	defaultSecondary = 0x20
)

func makeCE(w []int) colltab.Elem {
	ce, err := colltab.MakeElem(w[0], w[1], w[2], uint8(w[3]))
	if err != nil {
		panic(err)
	}
	return ce
}

func TestDoNorm(t *testing.T) {
	const div = -1 // The insertion point of the next block.
	tests := []struct {
		in, out []int
	}{{
		in:  []int{4, div, 3},
		out: []int{3, 4},
	}, {
		in:  []int{4, div, 3, 3, 3},
		out: []int{3, 3, 3, 4},
	}, {
		in:  []int{0, 4, div, 3},
		out: []int{0, 3, 4},
	}, {
		in:  []int{0, 0, 4, 5, div, 3, 3},
		out: []int{0, 0, 3, 3, 4, 5},
	}, {
		in:  []int{0, 0, 1, 4, 5, div, 3, 3},
		out: []int{0, 0, 1, 3, 3, 4, 5},
	}, {
		in:  []int{0, 0, 1, 4, 5, div, 4, 4},
		out: []int{0, 0, 1, 4, 4, 4, 5},
	},
	}
	for j, tt := range tests {
		i := Iter{}
		var w, p, s int
		for k, cc := range tt.in {
			if cc == 0 {
				s = 0
			}
			if cc == div {
				w = 100
				p = k
				i.pStarter = s
				continue
			}
			i.Elems = append(i.Elems, makeCE([]int{w, defaultSecondary, 2, cc}))
		}
		i.prevCCC = i.Elems[p-1].CCC()
		i.doNorm(p, i.Elems[p].CCC())
		if len(i.Elems) != len(tt.out) {
			t.Errorf("%d: length was %d; want %d", j, len(i.Elems), len(tt.out))
		}
		prevCCC := uint8(0)
		for k, ce := range i.Elems {
			if int(ce.CCC()) != tt.out[k] {
				t.Errorf("%d:%d: unexpected CCC. Was %d; want %d", j, k, ce.CCC(), tt.out[k])
			}
			if k > 0 && ce.CCC() == prevCCC && i.Elems[k-1].Primary() > ce.Primary() {
				t.Errorf("%d:%d: normalization crossed across CCC boundary.", j, k)
			}
		}
	}
	// test cutoff of large sequence of combining characters.
	result := []uint8{8, 8, 8, 5, 5}
	for o := -2; o <= 2; o++ {
		i := Iter{pStarter: 2, prevCCC: 8}
		n := maxCombiningCharacters + 1 + o
		for j := 1; j < n+i.pStarter; j++ {
			i.Elems = append(i.Elems, makeCE([]int{100, defaultSecondary, 2, 8}))
		}
		p := len(i.Elems)
		i.Elems = append(i.Elems, makeCE([]int{0, defaultSecondary, 2, 5}))
		i.doNorm(p, 5)
		if i.prevCCC != result[o+2] {
			t.Errorf("%d: i.prevCCC was %d; want %d", n, i.prevCCC, result[o+2])
		}
		if result[o+2] == 5 && i.pStarter != p {
			t.Errorf("%d: i.pStarter was %d; want %d", n, i.pStarter, p)
		}
	}
}
