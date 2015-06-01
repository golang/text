// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package colltab

import (
	"golang.org/x/text/collate/colltab"
)

// An Iter incrementally converts chunks of the input text to collation
// elements, while ensuring that the collation elements are in normalized order
// (that is, they are in the order as if the input text were normalized first).
type Iter struct {
	Weighter colltab.Weighter
	Elems    []colltab.Elem
	// N is the number of elements in Elems that will not be reordered on
	// subsequent iterations, N <= len(Elems).
	N int

	bytes []byte
	str   string

	prevCCC  uint8
	pStarter int
}

func (i *Iter) reset() {
	i.Elems = i.Elems[:0]
	i.N = 0
	i.prevCCC = 0
	i.pStarter = 0
}

// SetInput resets i to input s.
func (i *Iter) SetInput(s []byte) {
	i.bytes = s
	i.str = ""
	i.reset()
}

// SetInputString resets i to input s.
func (i *Iter) SetInputString(s string) {
	i.str = s
	i.bytes = nil
	i.reset()
}

func (i *Iter) done() bool {
	return len(i.str) == 0 && len(i.bytes) == 0
}

func (i *Iter) tail(n int) {
	if i.bytes == nil {
		i.str = i.str[n:]
	} else {
		i.bytes = i.bytes[n:]
	}
}

func (i *Iter) appendNext() int {
	var sz int
	if i.bytes == nil {
		i.Elems, sz = i.Weighter.AppendNextString(i.Elems, i.str)
	} else {
		i.Elems, sz = i.Weighter.AppendNext(i.Elems, i.bytes)
	}
	return sz
}

// Next appends Elems to the internal array until it adds an element with CCC=0.
// In the majority of cases, a Elem with a primary value > 0 will have a CCC of
// 0. The CCC values of collation elements are also used to detect if the input
// string was not normalized and to adjust the result accordingly.
func (i *Iter) Next() bool {
	for !i.done() {
		p0 := len(i.Elems)
		sz := i.appendNext()
		i.tail(sz)
		last := len(i.Elems) - 1
		if ccc := i.Elems[last].CCC(); ccc == 0 {
			i.N = len(i.Elems)
			i.pStarter = last
			i.prevCCC = 0
			return true
		} else if p0 < last && i.Elems[p0].CCC() == 0 {
			// set i.N to only cover part of i.Elems for which ccc == 0 and
			// use rest for the next call to next.
			for p0++; p0 < last && i.Elems[p0].CCC() == 0; p0++ {
			}
			i.N = p0
			i.pStarter = p0 - 1
			i.prevCCC = ccc
			return true
		} else if ccc < i.prevCCC {
			i.doNorm(p0, ccc) // should be rare, never occurs for NFD and FCC.
		} else {
			i.prevCCC = ccc
		}
	}
	if len(i.Elems) != i.N {
		i.N = len(i.Elems)
		return true
	}
	return false
}

// nextNoNorm is the same as next, but does not "normalize" the collation
// elements.
func (i *Iter) nextNoNorm() bool {
	// TODO: remove this function. Using this instead of next does not seem
	// to improve performance in any significant way. We retain this until
	// later for evaluation purposes.
	if i.done() {
		return false
	}
	sz := i.appendNext()
	i.tail(sz)
	i.N = len(i.Elems)
	return true
}

const maxCombiningCharacters = 30

// doNorm reorders the collation elements in i.Elems.
// It assumes that blocks of collation elements added with appendNext
// either start and end with the same CCC or start with CCC == 0.
// This allows for a single insertion point for the entire block.
// The correctness of this assumption is verified in builder.go.
func (i *Iter) doNorm(p int, ccc uint8) {
	if p-i.pStarter > maxCombiningCharacters {
		i.prevCCC = i.Elems[len(i.Elems)-1].CCC()
		i.pStarter = len(i.Elems) - 1
		return
	}
	n := len(i.Elems)
	k := p
	for p--; p > i.pStarter && ccc < i.Elems[p-1].CCC(); p-- {
	}
	i.Elems = append(i.Elems, i.Elems[p:k]...)
	copy(i.Elems[p:], i.Elems[k:])
	i.Elems = i.Elems[:n]
}
